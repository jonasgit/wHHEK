//-*- coding: utf-8 -*-

package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

const (
	chatQueryTimeout = 30 * time.Second
	chatMaxRows      = 1000
	chatMaxPlaces    = 300 // Platser som skickas till modellen, de vanligaste först

	// Modellens svar när frågan inte går att besvara med databasen, se chat_schema.txt.
	chatNoSQLAnswer = "INGEN_SQL"
)

var (
	chatSQLFence = regexp.MustCompile("(?s)```[A-Za-z]*\\s*(.*?)```")
	chatSQLWord  = regexp.MustCompile(`[\p{L}_][\p{L}\p{N}_]*`)
)

// chatForbiddenWords får inte förekomma i en fråga från modellen, förutom i strängar.
// De flesta stoppas även av den skrivskyddade anslutningen. INTO behövs för Access
// (SELECT ... INTO skapar en tabell) och SHELL för att Access kan köra VBA-funktioner.
var chatForbiddenWords = map[string]bool{
	"INSERT": true, "UPDATE": true, "DELETE": true, "DROP": true, "ALTER": true,
	"CREATE": true, "ATTACH": true, "DETACH": true, "PRAGMA": true, "VACUUM": true,
	"REINDEX": true, "INTO": true, "LOAD_EXTENSION": true, "SHELL": true,
}

// chatDialectRules är det som skiljer SQL mellan databaserna.
var chatDialectRules = map[string]string{
	"SQLite": "Beloppen i Budget kan vara sparade med decimalkomma, så räkna med dem som CAST(REPLACE(Jan, ',', '.') AS REAL).",
	"Microsoft Access": "Regler för Access: Använd IIF(villkor, a, b) och aldrig CASE WHEN. Använd aldrig LIMIT. Använd TOP bara när användaren ber om ett visst antal rader, och skriv då talet direkt efter TOP. En vanlig summa är bara SELECT SUM(Belopp) AS Summa utan TOP. Sortera med uttrycket och inte med namnet efter AS, alltså ORDER BY SUM(Belopp) DESC. " +
		"Använd aldrig WITH, CAST, strftime eller COUNT(DISTINCT). Ett namn efter AS med mellanslag eller andra specialtecken ska stå inom hakparenteser, till exempel AS [Totala utgifter], aldrig inom citattecken. " +
		"Beloppen i Budget är tal och används direkt, till exempel SELECT Mar FROM Budget.",
}

// chatResult är resultatet av en SQL-fråga, färdigt att visa.
type chatResult struct {
	Columns []string
	Rows    [][]string
	Cut     bool // Det finns fler rader än de som visas
}

// chatDatabase är databasen som modellens frågor körs mot. Query får inte kunna
// ändra databasen.
type chatDatabase struct {
	Dialect string
	Query   func(ctx context.Context, query string) (*chatResult, error)
}

// chatCurrentDatabase returnerar den öppna databasen, eller nil om ingen är öppen.
// db sätts först när ett eventuellt lösenord är kontrollerat.
func chatCurrentDatabase() *chatDatabase {
	if db == nil {
		return nil
	}
	filename := currentDatabase
	switch dbtype {
	case 1:
		return chatJetDatabase(filename)
	case 2:
		return chatSQLiteDatabase(filename)
	}
	return nil
}

// chatJetDatabase kör frågorna på en tillfällig kopia av filen. Access-drivrutinen
// bryr sig inte om READONLY eller ReadOnly=1: även sådana anslutningar kan ändra
// databasen (TestChatJetDatabase), så det går inte att skydda originalet på annat sätt.
func chatJetDatabase(filename string) *chatDatabase {
	return &chatDatabase{Dialect: "Microsoft Access", Query: func(ctx context.Context, query string) (*chatResult, error) {
		dir, err := os.MkdirTemp("", "wHHEKchat")
		if err != nil {
			return nil, err
		}
		defer func() {
			if err := os.RemoveAll(dir); err != nil {
				log.Println("chat: kunde inte radera tillfällig kopia: ", err)
			}
		}()
		contents, err := os.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		tempCopy := filepath.Join(dir, "kopia.mdb")
		if err := os.WriteFile(tempCopy, contents, 0o600); err != nil {
			return nil, err
		}
		return chatQueryReadOnly(ctx, query, func() (*sql.DB, error) { return connectJetDB(tempCopy, false) })
	}}
}

func chatSQLiteDatabase(filename string) *chatDatabase {
	return &chatDatabase{Dialect: "SQLite", Query: func(ctx context.Context, query string) (*chatResult, error) {
		return chatQueryReadOnly(ctx, query, func() (*sql.DB, error) {
			dsn, err := chatSQLiteReadOnlyDSN(filename)
			if err != nil {
				return nil, err
			}
			return sql.Open("sqlite3", dsn)
		})
	}}
}

// chatSQLiteReadOnlyDSN returnerar en adress där SQLite själv vägrar att skriva.
func chatSQLiteReadOnlyDSN(filename string) (string, error) {
	abs, err := filepath.Abs(filename)
	if err != nil {
		return "", err
	}
	path := filepath.ToSlash(abs)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path // Windows: C:/...
	}
	return (&url.URL{Scheme: "file", Path: path}).String() + "?mode=ro&_query_only=1", nil
}

// chatQueryReadOnly kör frågan på en egen anslutning, så att programmets vanliga
// anslutning till databasen inte påverkas. Anslutningen stängs innan funktionen är klar.
func chatQueryReadOnly(ctx context.Context, query string, open func() (*sql.DB, error)) (*chatResult, error) {
	conn, err := open()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(ctx, chatQueryTimeout)
	defer cancel()

	rows, err := conn.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	result := &chatResult{Columns: columns}
	var rawRows [][]any
	for rows.Next() {
		if len(rawRows) == chatMaxRows {
			result.Cut = true
			break
		}
		values := make([]any, len(columns))
		pointers := make([]any, len(columns))
		for i := range values {
			pointers[i] = &values[i]
		}
		if err := rows.Scan(pointers...); err != nil {
			return nil, err
		}
		rawRows = append(rawRows, values)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	result.Rows = chatFormatRows(rawRows, len(columns))
	return result, nil
}

// chatFormatRows gör rader klara att visa. SQLite ger heltal för belopp utan öre, så
// finns det decimaltal i en kolumn visas hela kolumnen som belopp.
func chatFormatRows(rawRows [][]any, columnCount int) [][]string {
	// SUM och liknande ger en enda rad med NULL när inget matchar.
	if len(rawRows) == 1 && !slices.ContainsFunc(rawRows[0], func(v any) bool { return v != nil }) {
		return nil
	}

	amountColumn := make([]bool, columnCount)
	for _, row := range rawRows {
		for i, value := range row {
			if _, isFloat := value.(float64); isFloat {
				amountColumn[i] = true
			}
		}
	}

	formatted := make([][]string, 0, len(rawRows))
	for _, raw := range rawRows {
		row := make([]string, columnCount)
		for i, value := range raw {
			if number, isInt := value.(int64); isInt && amountColumn[i] {
				value = float64(number)
			}
			row[i] = chatFormatValue(value)
		}
		formatted = append(formatted, row)
	}
	return formatted
}

func chatFormatValue(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case []byte:
		return toUtf8(v)
	case float64:
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return fmt.Sprint(v)
		}
		return Dec2Str(decimal.NewFromFloat(v).Round(2))
	default:
		return fmt.Sprint(v)
	}
}

// chatExtractSQL plockar ut SQL-frågan ur modellens svar. Modellen ska bara svara
// med SQL men lägger ibland frågan i ett kodblock.
func chatExtractSQL(reply string) string {
	if m := chatSQLFence.FindStringSubmatch(reply); m != nil {
		reply = m[1]
	}
	return strings.TrimSpace(reply)
}

// chatStripSQL tar bort strängar, citerade namn och kommentarer ur en SQL-fråga, så
// att det som är kvar är sådant som databasen tolkar som kod.
func chatStripSQL(query string) (string, error) {
	var out strings.Builder
	for i := 0; i < len(query); {
		c := query[i]
		var end string
		switch {
		case c == '\'' || c == '"' || c == '`':
			end = string(c)
		case c == '[':
			end = "]"
		case strings.HasPrefix(query[i:], "--"):
			end = "\n"
		case strings.HasPrefix(query[i:], "/*"):
			end = "*/"
		default:
			out.WriteByte(c)
			i++
			continue
		}
		start := i + 1
		if len(end) == 2 || c == '-' {
			start = i + 2
		}
		closing := strings.Index(query[start:], end)
		switch {
		case closing >= 0:
			i = start + closing + len(end)
		case c == '-':
			i = len(query) // En radkommentar får ta slut med frågan
		default:
			return "", errors.New("en sträng eller kommentar avslutas inte")
		}
		out.WriteByte(' ')
	}
	return out.String(), nil
}

// chatCheckSQL kontrollerar att frågan bara läser data. Det kompletterar den
// skrivskyddade anslutningen. go-sqlite3 kör alla satser i en sträng, så frågan
// måste vara en enda sats.
func chatCheckSQL(query string) error {
	// Kontrolleras i hela texten, eftersom namn kan skrivas [DtbVer] eller "DtbVer".
	lower := strings.ToLower(query)
	if strings.Contains(lower, "dtbver") || strings.Contains(lower, "losenord") {
		return errors.New("tabellen DtbVer med databasens lösenord får inte läsas")
	}

	code, err := chatStripSQL(query)
	if err != nil {
		return err
	}
	code = strings.TrimSpace(code)
	if strings.Contains(strings.TrimRight(code, "; \t\r\n"), ";") {
		return errors.New("bara en sats i taget får köras")
	}
	words := chatSQLWord.FindAllString(code, -1)
	if len(words) == 0 {
		return errors.New("frågan är tom")
	}
	if first := strings.ToUpper(words[0]); first != "SELECT" && first != "WITH" {
		return fmt.Errorf("frågan börjar med %s och inte med SELECT", words[0])
	}
	for _, word := range words {
		if chatForbiddenWords[strings.ToUpper(word)] {
			return fmt.Errorf("%s är inte tillåtet", strings.ToUpper(word))
		}
	}
	return nil
}

// chatNameLists hämtar användarens kategorier, konton och platser, så att modellen kan
// använda de riktiga namnen. Kategorierna kommer från budgeten och från transaktionerna,
// som kan ha kategorier som saknas i budgeten. Platserna kommer från transaktionerna,
// eftersom inte alla finns i tabellen Platser. Går det inte att läsa dem returneras det
// som gick, eftersom modellen kan svara ändå.
func chatNameLists(ctx context.Context, database *chatDatabase) string {
	lists := []struct {
		title, query string
		limit        int
	}{
		{"Inkomstkategorier", `SELECT Typ FROM Budget WHERE Inkomst = 'J'
			UNION SELECT Vad FROM Transaktioner WHERE Typ IN ('Insättning', 'Fast Inkomst') AND Vad <> '---'`, chatMaxRows},
		{"Utgiftskategorier", `SELECT Typ FROM Budget WHERE Inkomst = 'N'
			UNION SELECT Vad FROM Transaktioner WHERE Typ IN ('Inköp', 'Fast Utgift') AND Vad <> '---'`, chatMaxRows},
		{"Konton", `SELECT Benämning FROM Konton`, chatMaxRows},
		{"Platser", `SELECT TillKonto FROM Transaktioner WHERE Typ IN ('Inköp', 'Fast Utgift')
			GROUP BY TillKonto ORDER BY COUNT(*) DESC`, chatMaxPlaces},
	}
	var out strings.Builder
	for _, list := range lists {
		result, err := database.Query(ctx, list.query)
		if err != nil {
			log.Println("chat: kan inte läsa namn: ", err)
			continue
		}
		var names []string
		for _, row := range result.Rows {
			if name := strings.TrimSpace(row[0]); name != "" {
				names = append(names, name)
			}
		}
		title := list.title
		if len(names) > list.limit {
			names = names[:list.limit]
			title += " (de vanligaste, det finns fler)"
		}
		if len(names) > 0 {
			fmt.Fprintf(&out, "%s: %s\n", title, strings.Join(names, "; "))
		}
	}
	return out.String()
}

// chatSystemPrompt är bakgrunden som skickas med varje fråga.
func chatSystemPrompt(dialect string, today time.Time, names string) string {
	year := today.Year()
	return fmt.Sprintf("%s\nDatabasen är %s. %s\nDagens datum är %s. \"I år\" betyder hela året, alltså Datum >= '%d-01-01' AND Datum <= '%d-12-31'. Om användaren inte nämner någon tid, filtrera inte på datum.\n%s",
		chatBackground, dialect, chatDialectRules[dialect], today.Format("2006-01-02"), year, year, names)
}

// chatQueryError betyder att databasen inte kunde köra modellens fråga. Det är det enda
// felet som får ett nytt försök: en fråga som stoppas av chatCheckSQL ska inte bli en
// omskriven fråga, och INGEN_SQL är modellens eget svar.
type chatQueryError struct {
	query string
	err   error
}

func (e *chatQueryError) Error() string { return e.err.Error() }

// chatRunReply tar modellens svar, kör SQL-frågan och fyller i resultatet i data.
// Returnerar frågan som kördes.
func chatRunReply(ctx context.Context, database *chatDatabase, reply string, data *ChatReplyData) (string, error) {
	data.SQL = ""
	if strings.Contains(reply, chatNoSQLAnswer) {
		return "", llmProblem("Frågan går inte att besvara med uppgifterna i databasen.")
	}

	query := chatExtractSQL(reply)
	data.SQL = query
	if err := chatCheckSQL(query); err != nil {
		return "", llmProblem(fmt.Sprintf("Modellen svarade inte med en fråga som går att köra, så inget kördes: %v.", err))
	}

	result, err := database.Query(ctx, query)
	if err != nil {
		return "", &chatQueryError{query: query, err: err}
	}
	data.Result = result
	return query, nil
}

// chatRetryMessages är det som läggs till i samtalet när databasen inte kunde köra
// modellens fråga. Det sparas inte i samtalet.
func chatRetryMessages(failed *chatQueryError) []chatMessage {
	return []chatMessage{
		{Role: "assistant", Content: failed.query},
		{Role: "user", Content: fmt.Sprintf("Databasen kunde inte köra den frågan. Felet var: %v\nSkriv om frågan så att den fungerar. Svara ENDAST med den rättade SQL-frågan.", failed.err)},
	}
}
