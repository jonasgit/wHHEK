//-*- coding: utf-8 -*-

package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// Exakt svar från LM Studio när en fråga inte får plats i kontexten.
const lmStudioContextError = `{"error":"Engine protocol predict request returned 400: {\"error\":{\"code\":400,\"message\":\"request (13243 tokens) exceeds the available context size (8192 tokens), try increasing it\",\"type\":\"exceed_context_size_error\",\"n_prompt_tokens\":13243,\"n_ctx\":8192}}"}`

// fakeDatabase svarar med samma resultat på alla frågor.
var fakeDatabase = &chatDatabase{Dialect: "SQLite", Query: func(context.Context, string) (*chatResult, error) {
	return &chatResult{Columns: []string{"Summa"}, Rows: [][]string{{"resultatet"}}}, nil
}}

// fakeLLM startar en server som beter sig som LM Studio. Den listar först en
// inbäddningsmodell, som inte får väljas.
func fakeLLM(t *testing.T, chatHandler http.HandlerFunc) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/models", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"data":[{"id":"text-embedding-nomic"},{"id":"chatmodellen"}],"object":"list"}`)
	})
	mux.HandleFunc("/api/v0/models", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"data":[{"id":"text-embedding-nomic","state":"not-loaded"},{"id":"chatmodellen","loaded_context_length":10000}]}`)
	})
	mux.HandleFunc("/v1/chat/completions", chatHandler)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func llmReply(w http.ResponseWriter, content string) {
	_ = json.NewEncoder(w).Encode(map[string]any{
		"choices": []map[string]any{{"message": map[string]string{"role": "assistant", "content": content}}},
	})
}

func postQuestion(t *testing.T, baseURL, question string) string {
	t.Helper()
	return postChat(t, baseURL, url.Values{"question": {question}})
}

func postChat(t *testing.T, baseURL string, form url.Values) string {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/chatask", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	serveChatAsk(rec, req, baseURL, fakeDatabase)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d, htmx visar bara svar med 2xx", rec.Code)
	}
	return rec.Body.String()
}

// closedServerURL returnerar en adress där ingen server lyssnar.
func closedServerURL() string {
	srv := httptest.NewServer(http.NotFoundHandler())
	srv.Close()
	return srv.URL
}

func TestChatPageServerUnavailable(t *testing.T) {
	t.Parallel()
	baseURL := closedServerURL()

	rec := httptest.NewRecorder()
	serveChat(rec, httptest.NewRequest(http.MethodGet, "/chat", nil), baseURL, fakeDatabase)
	body := rec.Body.String()

	for _, want := range []string{baseURL, "dial tcp", "/v1/chat/completions", "LM Studio"} {
		if !strings.Contains(body, want) {
			t.Errorf("felmeddelandet saknar %q:\n%s", want, body)
		}
	}
	if strings.Contains(body, `id="question"`) {
		t.Error("inmatningsrutan ska inte visas när servern saknas")
	}
}

func TestChatPageServerAvailable(t *testing.T) {
	t.Parallel()
	srv := fakeLLM(t, func(http.ResponseWriter, *http.Request) {})

	rec := httptest.NewRecorder()
	serveChat(rec, httptest.NewRequest(http.MethodGet, "/chat", nil), srv.URL, fakeDatabase)
	body := rec.Body.String()

	for _, want := range []string{`id="question"`, `hx-post="/chatask"`, "chatmodellen"} {
		if !strings.Contains(body, want) {
			t.Errorf("sidan saknar %q:\n%s", want, body)
		}
	}
	if strings.Contains(body, "text-embedding") {
		t.Error("inbäddningsmodellen ska inte väljas")
	}
}

func TestChatPageOnlyEmbeddingModel(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"data":[{"id":"text-embedding-nomic"}]}`)
	}))
	t.Cleanup(srv.Close)

	rec := httptest.NewRecorder()
	serveChat(rec, httptest.NewRequest(http.MethodGet, "/chat", nil), srv.URL, fakeDatabase)
	body := rec.Body.String()

	if !strings.Contains(body, "ingen chattmodell") || strings.Contains(body, `id="question"`) {
		t.Errorf("väntade felmeddelande om chattmodell:\n%s", body)
	}
}

func TestChatAskShowsResult(t *testing.T) {
	t.Parallel()
	var got struct {
		Model    string        `json:"model"`
		Stream   bool          `json:"stream"`
		Messages []chatMessage `json:"messages"`
	}
	srv := fakeLLM(t, func(w http.ResponseWriter, req *http.Request) {
		if err := json.NewDecoder(req.Body).Decode(&got); err != nil {
			t.Error(err)
		}
		llmReply(w, "SELECT '<b>Hemlig_SQL</b>'")
	})

	body := postQuestion(t, srv.URL, "Vad är två & två?")

	if got.Model != "chatmodellen" || got.Stream || len(got.Messages) != 2 ||
		got.Messages[0].Role != "system" || got.Messages[1].Role != "user" ||
		got.Messages[1].Content != "Vad är två & två?" {
		t.Errorf("fel begäran till LLM-servern: %+v", got)
	}
	if !strings.Contains(body, "<td>resultatet</td>") || !strings.Contains(body, "<th>Summa</th>") {
		t.Errorf("resultatet ska visas:\n%s", body)
	}
	if !strings.Contains(body, "Vad är två &amp; två?") {
		t.Errorf("frågan ska visas:\n%s", body)
	}
	if !strings.Contains(body, `id="question"`) || !strings.Contains(body, "hx-swap-oob") {
		t.Errorf("inmatningsrutan ska tömmas efter svar:\n%s", body)
	}
	// SQL-frågan ska finnas men vara hopfälld och escapad.
	before, after, found := strings.Cut(body, "<details><summary>Visa SQL</summary>")
	if !found || strings.Contains(before, "Hemlig_SQL") {
		t.Errorf("SQL-frågan ska vara hopfälld under Visa SQL:\n%s", body)
	}
	if !strings.Contains(after, "&lt;b&gt;Hemlig_SQL") || strings.Contains(body, "<b>Hemlig_SQL") {
		t.Errorf("SQL-frågan ska visas escapad:\n%s", body)
	}
}

func TestChatSystemPrompt(t *testing.T) {
	t.Parallel()
	got := chatSystemPrompt("SQLite", time.Date(2026, 11, 2, 12, 0, 0, 0, time.UTC), "Konton: Lönekonto\n")
	for _, want := range []string{"CREATE TABLE Transaktioner", "ENDAST", "Databasen är SQLite", "Dagens datum är 2026-11-02", "Datum >= '2026-01-01' AND Datum <= '2026-12-31'", "Konton: Lönekonto"} {
		if !strings.Contains(got, want) {
			t.Errorf("bakgrunden saknar %q", want)
		}
	}
}

func TestChatSystemPromptDialectRules(t *testing.T) {
	t.Parallel()
	today := time.Date(2026, 11, 2, 12, 0, 0, 0, time.UTC)
	sqlite := chatSystemPrompt("SQLite", today, "")
	access := chatSystemPrompt("Microsoft Access", today, "")

	if !strings.Contains(access, "IIF(") || !strings.Contains(access, "aldrig LIMIT") {
		t.Error("Access-reglerna saknas")
	}
	if strings.Contains(sqlite, "IIF(") {
		t.Error("Access-reglerna ska inte skickas för SQLite")
	}
	if !strings.Contains(sqlite, "CAST(REPLACE") {
		t.Error("SQLite ska få regeln om decimalkomma")
	}
	if strings.Contains(access, "CAST(REPLACE") {
		t.Error("Access kan inte köra CAST")
	}
}

func TestChatNoDatabase(t *testing.T) {
	t.Parallel()
	srv := fakeLLM(t, func(http.ResponseWriter, *http.Request) {
		t.Error("frågan ska inte skickas när ingen databas är öppen")
	})

	rec := httptest.NewRecorder()
	serveChat(rec, httptest.NewRequest(http.MethodGet, "/chat", nil), srv.URL, nil)
	if body := rec.Body.String(); !strings.Contains(body, "Ingen databas är öppen") || strings.Contains(body, `id="question"`) {
		t.Errorf("sidan ska säga att databas saknas:\n%s", body)
	}

	form := url.Values{"question": {"Fråga"}}
	req := httptest.NewRequest(http.MethodPost, "/chatask", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	serveChatAsk(rec, req, srv.URL, nil)
	if !strings.Contains(rec.Body.String(), "Ingen databas är öppen") {
		t.Errorf("väntade meddelande:\n%s", rec.Body.String())
	}
}

func TestChatCheckSQL(t *testing.T) {
	t.Parallel()
	allowed := []string{
		"SELECT 1",
		"select Vad, SUM(Belopp) FROM Transaktioner GROUP BY Vad;",
		"WITH x AS (SELECT 1 AS n) SELECT n FROM x",
		"SELECT * FROM Transaktioner WHERE Vad = 'DROP; DELETE'",
		"SELECT REPLACE(Jan, ',', '.') FROM Budget",
		"SELECT [Text] FROM Transaktioner -- en kommentar; DELETE",
		"SELECT 'it''s'",
	}
	for _, query := range allowed {
		if err := chatCheckSQL(query); err != nil {
			t.Errorf("%q ska tillåtas: %v", query, err)
		}
	}
	rejected := []string{
		"",
		"Jag vet inte",
		"DELETE FROM Konton",
		"SELECT 1; DELETE FROM Konton",
		"SELECT 1 /* ; */; DROP TABLE Konton",
		"SELECT ';'; DROP TABLE Konton",
		"WITH x AS (SELECT 1) DELETE FROM Konton",
		"WITH x AS (SELECT 1) INSERT INTO Konton SELECT * FROM x",
		"SELECT * INTO Kopia FROM Konton",
		"PRAGMA writable_schema=1",
		"ATTACH 'x.db' AS y",
		"SELECT * FROM DtbVer",
		"SELECT * FROM [DtbVer]",
		`SELECT * FROM "dtbver"`,
		"SELECT Losenord FROM x",
		"SELECT load_extension('x')",
		"SELECT Shell('calc')",
		"SELECT 'oavslutad",
		"SELECT 1 /* oavslutad",
	}
	for _, query := range rejected {
		if err := chatCheckSQL(query); err == nil {
			t.Errorf("%q ska stoppas", query)
		}
	}
}

func TestChatExtractSQL(t *testing.T) {
	t.Parallel()
	tests := map[string]string{
		"SELECT 1":                           "SELECT 1",
		"  SELECT 1;\n":                      "SELECT 1;",
		"```sql\nSELECT 1\n```":              "SELECT 1",
		"```sqlite\nSELECT 1\n```":           "SELECT 1",
		"Här är frågan:\n```\nSELECT 1\n```": "SELECT 1",
	}
	for reply, want := range tests {
		if got := chatExtractSQL(reply); got != want {
			t.Errorf("%q: got %q", reply, got)
		}
	}
}

// newSQLiteChatDatabase skapar en riktig databas med programmets tabeller.
func newSQLiteChatDatabase(t *testing.T, filename string) (*chatDatabase, *sql.DB) {
	t.Helper()
	path := filepath.Join(t.TempDir(), filename)
	conn, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	InitiateDB(conn)
	return chatSQLiteDatabase(path), conn
}

func countRows(t *testing.T, conn *sql.DB, table string) int {
	t.Helper()
	var n int
	if err := conn.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// Anslutningen ska själv vägra att skriva, även om en fråga skulle slinka förbi chatCheckSQL.
func TestChatQueryConnectionIsReadOnly(t *testing.T) {
	t.Parallel()
	database, conn := newSQLiteChatDatabase(t, "ro.db")
	before := countRows(t, conn, "Budget")

	for _, query := range []string{
		"DELETE FROM Budget",
		"INSERT INTO Konton (Benämning) VALUES ('x')",
		"DROP TABLE Budget",
		"SELECT 1; DELETE FROM Budget",
		"WITH x AS (SELECT 1) DELETE FROM Budget",
		"CREATE TABLE Ny (a)",
	} {
		if _, err := database.Query(context.Background(), query); err == nil {
			t.Errorf("%q ska inte gå att köra", query)
		}
	}

	if after := countRows(t, conn, "Budget"); after != before {
		t.Errorf("Budget ändrades: %d -> %d", before, after)
	}
	if n := countRows(t, conn, "Konton"); n != 1 {
		t.Errorf("Konton ändrades: %d", n)
	}
}

func TestChatQueryFilenameWithSpecialCharacters(t *testing.T) {
	t.Parallel()
	database, _ := newSQLiteChatDatabase(t, "min ekonomi #1 räksmörgås 100%.db")

	result, err := database.Query(context.Background(), "SELECT Benämning FROM Konton")

	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rows) != 1 || result.Rows[0][0] != "Plånboken" {
		t.Errorf("fel resultat: %+v", result)
	}
}

func TestChatQueryFormatsValues(t *testing.T) {
	t.Parallel()
	database, _ := newSQLiteChatDatabase(t, "format.db")

	result, err := database.Query(context.Background(),
		"SELECT NULL AS a, 42 AS b, 1234.5 AS c, 0.1 + 0.2 AS d, 'åäö' AS e")

	if err != nil {
		t.Fatal(err)
	}
	want := []string{"", "42", "1 234,50", "0,30", "åäö"}
	if !slices.Equal(result.Rows[0], want) {
		t.Errorf("got %q, want %q", result.Rows[0], want)
	}
	if !slices.Equal(result.Columns, []string{"a", "b", "c", "d", "e"}) {
		t.Errorf("kolumner: %q", result.Columns)
	}
}

func TestChatQueryAmountColumnIsConsistent(t *testing.T) {
	t.Parallel()
	database, _ := newSQLiteChatDatabase(t, "amount.db")

	result, err := database.Query(context.Background(),
		"SELECT 1 AS id, 8000 AS belopp UNION ALL SELECT 2, 1500.5 ORDER BY id")

	if err != nil {
		t.Fatal(err)
	}
	want := [][]string{{"1", "8 000,00"}, {"2", "1 500,50"}}
	if !slices.EqualFunc(result.Rows, want, slices.Equal[[]string]) {
		t.Errorf("got %q, want %q", result.Rows, want)
	}
}

// SUM ger en rad med NULL när inget matchar, och det ska inte visas som en tom ruta.
func TestChatQueryOnlyNullIsNoResult(t *testing.T) {
	t.Parallel()
	database, _ := newSQLiteChatDatabase(t, "null.db")

	result, err := database.Query(context.Background(), "SELECT SUM(Belopp) AS Summa FROM Transaktioner WHERE Vad = 'Finns inte'")

	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rows) != 0 {
		t.Errorf("väntade inga rader: %q", result.Rows)
	}
}

func TestChatQueryLimitsRows(t *testing.T) {
	t.Parallel()
	database, _ := newSQLiteChatDatabase(t, "many.db")
	const query = `WITH RECURSIVE n(i) AS (SELECT 1 UNION ALL SELECT i + 1 FROM n WHERE i < %d) SELECT i FROM n`

	result, err := database.Query(context.Background(), fmt.Sprintf(query, chatMaxRows+1))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rows) != chatMaxRows || !result.Cut {
		t.Errorf("rader %d, kapat %v", len(result.Rows), result.Cut)
	}

	result, err = database.Query(context.Background(), fmt.Sprintf(query, chatMaxRows))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rows) != chatMaxRows || result.Cut {
		t.Errorf("exakt max: rader %d, kapat %v", len(result.Rows), result.Cut)
	}
}

// En fråga som aldrig blir klar får inte låsa programmet.
func TestChatQueryStopsAtDeadline(t *testing.T) {
	t.Parallel()
	database, _ := newSQLiteChatDatabase(t, "slow.db")
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		_, err := database.Query(ctx, "WITH RECURSIVE n(i) AS (SELECT 1 UNION ALL SELECT i + 1 FROM n) SELECT COUNT(*) FROM n")
		done <- err
	}()
	select {
	case err := <-done:
		if err == nil {
			t.Error("väntade fel")
		}
	case <-time.After(10 * time.Second):
		t.Fatal("frågan avbröts inte")
	}
}

func TestChatNameLists(t *testing.T) {
	t.Parallel()
	database, conn := newSQLiteChatDatabase(t, "names.db")
	if _, err := conn.Exec(`
		INSERT INTO Konton (Benämning) VALUES ('Lönekonto');
		INSERT INTO Budget (Typ, Inkomst) VALUES ('Egen utgift', 'N'), ('Egen inkomst', 'J');
		INSERT INTO Transaktioner (Typ, Vad, TillKonto) VALUES
			('Inköp', 'Gammal kategori', 'Hemligbutiken'),
			('Inköp', 'Livsmedel', 'Hemligbutiken'),
			('Fast Inkomst', 'Gammal inkomst', 'Lönekonto'),
			('Överföring', '---', 'Lönekonto');`); err != nil {
		t.Fatal(err)
	}

	lists := strings.Split(strings.TrimSpace(chatNameLists(context.Background(), database)), "\n")
	if len(lists) != 4 {
		t.Fatalf("väntade fyra listor: %q", lists)
	}
	checks := []struct {
		prefix  string
		want    []string
		notWant []string
	}{
		{"Inkomstkategorier: ", []string{"Pension", "Egen inkomst", "Gammal inkomst"}, []string{"Livsmedel", "---"}},
		{"Utgiftskategorier: ", []string{"Livsmedel", "Egen utgift", "Gammal kategori"}, []string{"Pension", "---"}},
		{"Konton: ", []string{"Plånboken", "Lönekonto"}, nil},
		{"Platser: ", []string{"Hemligbutiken"}, []string{"Lönekonto", "---"}},
	}
	for i, check := range checks {
		list, found := strings.CutPrefix(lists[i], check.prefix)
		if !found {
			t.Fatalf("väntade %q: %q", check.prefix, lists[i])
		}
		names := strings.Split(list, "; ")
		for _, want := range check.want {
			if !slices.Contains(names, want) {
				t.Errorf("%s saknar %q: %q", check.prefix, want, names)
			}
		}
		for _, notWant := range check.notWant {
			if slices.Contains(names, notWant) {
				t.Errorf("%s ska inte ha %q: %q", check.prefix, notWant, names)
			}
		}
		if n := strings.Count(list, "Livsmedel"); n > 1 {
			t.Errorf("%s har dubbletter: %q", check.prefix, list)
		}
	}
	// Belopp och personer ska inte skickas.
	if all := strings.Join(lists, "\n"); strings.Contains(all, "Gemensamt") {
		t.Errorf("personer ska inte skickas: %q", all)
	}
}

// Vanligaste platserna först, och listan får inte bli hur lång som helst.
func TestChatNameListsPlaces(t *testing.T) {
	t.Parallel()
	database, conn := newSQLiteChatDatabase(t, "places.db")
	if _, err := conn.Exec(`INSERT INTO Transaktioner (Typ, Vad, TillKonto) VALUES
		('Inköp', 'Livsmedel', 'Sällan'), ('Fast Utgift', 'Bil', 'Ofta'), ('Inköp', 'Bil', 'Ofta')`); err != nil {
		t.Fatal(err)
	}
	if got := chatNameLists(context.Background(), database); !strings.Contains(got, "Platser: Ofta; Sällan\n") {
		t.Errorf("fel ordning: %q", got)
	}

	var insert strings.Builder
	insert.WriteString("INSERT INTO Transaktioner (Typ, Vad, TillKonto) VALUES ('Inköp', 'Bil', 'P0')")
	for i := 1; i < chatMaxPlaces+5; i++ {
		fmt.Fprintf(&insert, ", ('Inköp', 'Bil', 'P%d')", i)
	}
	if _, err := conn.Exec(insert.String()); err != nil {
		t.Fatal(err)
	}
	got := chatNameLists(context.Background(), database)
	_, list, found := strings.Cut(got, "Platser (de vanligaste, det finns fler): ")
	if !found {
		t.Fatalf("väntade att listan kapas och att det står så: %q", got)
	}
	if n := len(strings.Split(strings.TrimSpace(list), "; ")); n != chatMaxPlaces {
		t.Errorf("antal platser: %d", n)
	}
}

// fakeLLMReplies svarar med ett svar i taget och sparar vad modellen fick.
func fakeLLMReplies(t *testing.T, replies ...string) (*httptest.Server, func() [][]chatMessage) {
	t.Helper()
	var mu sync.Mutex
	var requests [][]chatMessage
	srv := fakeLLM(t, func(w http.ResponseWriter, req *http.Request) {
		var payload struct {
			Messages []chatMessage `json:"messages"`
		}
		_ = json.NewDecoder(req.Body).Decode(&payload)
		mu.Lock()
		defer mu.Unlock()
		if len(requests) >= len(replies) {
			t.Errorf("oväntad fråga nummer %d till modellen", len(requests)+1)
			llmReply(w, "SELECT 1")
			return
		}
		requests = append(requests, payload.Messages)
		llmReply(w, replies[len(requests)-1])
	})
	return srv, func() [][]chatMessage {
		mu.Lock()
		defer mu.Unlock()
		return slices.Clone(requests)
	}
}

func postToDatabase(t *testing.T, baseURL string, database *chatDatabase, question string) string {
	t.Helper()
	form := url.Values{"question": {question}}
	req := httptest.NewRequest(http.MethodPost, "/chatask", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	serveChatAsk(rec, req, baseURL, database)
	return rec.Body.String()
}

func TestChatRetrySucceedsAndIsVisible(t *testing.T) {
	t.Parallel()
	database, _ := newSQLiteChatDatabase(t, "retry.db")
	srv, requests := fakeLLMReplies(t, "SELECT Finns_inte FROM Budget", "SELECT COUNT(*) AS Antal FROM Konton")

	body := postToDatabase(t, srv.URL, database, "Hur många konton?")

	for _, want := range []string{"<td>1</td>", "nytt försök", "Visa första försöket", "no such column: Finns_inte", "SELECT Finns_inte FROM Budget", "Visa SQL", "SELECT COUNT(*) AS Antal FROM Konton"} {
		if !strings.Contains(body, want) {
			t.Errorf("sidan saknar %q:\n%s", want, body)
		}
	}
	if !strings.Contains(body, "hx-swap-oob") {
		t.Errorf("frågan lyckades, så inmatningsrutan ska tömmas:\n%s", body)
	}

	sent := requests()
	if len(sent) != 2 {
		t.Fatalf("antal frågor till modellen: %d", len(sent))
	}
	second := sent[1]
	if len(second) != 4 || second[1].Content != "Hur många konton?" ||
		second[2].Role != "assistant" || second[2].Content != "SELECT Finns_inte FROM Budget" ||
		second[3].Role != "user" || !strings.Contains(second[3].Content, "no such column: Finns_inte") {
		t.Errorf("fel fråga vid nytt försök: %+v", second[1:])
	}

	// Samtalet ska bara ha den fråga som fungerade.
	saved, err := chatParseHistory(historyField(t, body))
	if err != nil {
		t.Fatal(err)
	}
	if len(saved) != 2 || saved[1].Content != "SELECT COUNT(*) AS Antal FROM Konton" {
		t.Errorf("fel samtal: %+v", saved)
	}
}

func TestChatRetryFailsTwice(t *testing.T) {
	t.Parallel()
	database, _ := newSQLiteChatDatabase(t, "retry2.db")
	srv, requests := fakeLLMReplies(t, "SELECT Forsta FROM Budget", "SELECT Andra FROM Budget")

	body := postToDatabase(t, srv.URL, database, "Fråga")

	for _, want := range []string{"inte heller efter ett nytt försök", "no such column: Andra", "SELECT Andra FROM Budget", "Visa första försöket", "no such column: Forsta", "SELECT Forsta FROM Budget"} {
		if !strings.Contains(body, want) {
			t.Errorf("sidan saknar %q:\n%s", want, body)
		}
	}
	if strings.Contains(body, "hx-swap-oob") {
		t.Errorf("frågan ska stå kvar:\n%s", body)
	}
	if n := len(requests()); n != 2 {
		t.Errorf("bara ett nytt försök ska göras, men modellen fick %d frågor", n)
	}
}

// Det andra försöket går genom samma kontroll som det första.
func TestChatRetryCannotWrite(t *testing.T) {
	t.Parallel()
	database, conn := newSQLiteChatDatabase(t, "retry3.db")
	before := countRows(t, conn, "Budget")
	srv, requests := fakeLLMReplies(t, "SELECT Finns_inte FROM Budget", "DELETE FROM Budget")

	body := postToDatabase(t, srv.URL, database, "Fråga")

	if !strings.Contains(body, "så inget kördes") || !strings.Contains(body, "Visa första försöket") {
		t.Errorf("väntade att andra försöket stoppas:\n%s", body)
	}
	if n := countRows(t, conn, "Budget"); n != before {
		t.Errorf("Budget ändrades: %d -> %d", before, n)
	}
	if n := len(requests()); n != 2 {
		t.Errorf("antal frågor till modellen: %d", n)
	}
}

// Stoppade frågor och INGEN_SQL ska inte få något nytt försök, och lyckade frågor
// ska inte visa något om nya försök.
func TestChatNoRetryWhenNotDatabaseError(t *testing.T) {
	t.Parallel()
	database, _ := newSQLiteChatDatabase(t, "noretry.db")
	for _, reply := range []string{"DELETE FROM Budget", "SELECT * FROM DtbVer", "INGEN_SQL", "Jag vet inte", "SELECT 1"} {
		srv, requests := fakeLLMReplies(t, reply)
		body := postToDatabase(t, srv.URL, database, "Fråga")
		if n := len(requests()); n != 1 {
			t.Errorf("%q: modellen fick %d frågor", reply, n)
		}
		if strings.Contains(body, "nytt försök") || strings.Contains(body, "första försöket") {
			t.Errorf("%q: ska inte nämna nytt försök:\n%s", reply, body)
		}
	}
}

func TestChatBackgroundCountsFixedExpenses(t *testing.T) {
	t.Parallel()
	if !strings.Contains(chatBackground, "alltså Typ IN ('Inköp', 'Fast Utgift')") {
		t.Error("bakgrunden ska säga att Fast Utgift räknas som utgift")
	}
}

// Namnen ska följa med till modellen men inte sparas i samtalet.
func TestChatAskSendsNames(t *testing.T) {
	t.Parallel()
	database, conn := newSQLiteChatDatabase(t, "sendnames.db")
	if _, err := conn.Exec(`INSERT INTO Konton (Benämning) VALUES ('Lönekonto')`); err != nil {
		t.Fatal(err)
	}
	var system string
	srv := fakeLLM(t, func(w http.ResponseWriter, req *http.Request) {
		var payload struct {
			Messages []chatMessage `json:"messages"`
		}
		_ = json.NewDecoder(req.Body).Decode(&payload)
		system = payload.Messages[0].Content
		llmReply(w, "SELECT 1")
	})

	form := url.Values{"question": {"Fråga"}}
	req := httptest.NewRequest(http.MethodPost, "/chatask", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	serveChatAsk(rec, req, srv.URL, database)

	if !strings.Contains(system, "Lönekonto") || !strings.Contains(system, "Livsmedel") {
		t.Errorf("namnen saknas i bakgrunden:\n%s", system)
	}
	if strings.Contains(historyField(t, rec.Body.String()), "Lönekonto") {
		t.Error("namnen ska inte sparas i samtalet")
	}
}

// Modellen ska kunna svara även om namnen inte går att läsa.
func TestChatNameListsFailure(t *testing.T) {
	t.Parallel()
	database := &chatDatabase{Dialect: "SQLite", Query: func(context.Context, string) (*chatResult, error) {
		return nil, errors.New("trasig databas")
	}}
	if got := chatNameLists(context.Background(), database); got != "" {
		t.Errorf("got %q", got)
	}
}

// Access: frågorna ska fungera och anslutningen ska inte kunna ändra databasen.
func TestChatJetDatabase(t *testing.T) {
	if !JetDBSupport {
		t.Skip("Jet stöds inte i det här bygget")
	}
	filename := filepath.Join(t.TempDir(), "chat.mdb")
	SkapaTomMDB(t, filename)
	conn, err := connectJetDB(filename, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	for _, stmt := range []string{
		`INSERT INTO Konton (KontoNummer, Benämning) VALUES ('1', 'Lönekonto')`,
		`INSERT INTO Transaktioner (FrånKonto, TillKonto, Typ, Datum, Vad, Vem, Belopp) VALUES ('Lönekonto', 'Räkmacka', 'Inköp', '2026-03-02', 'Livsmedel', 'Gemensamt', 1500)`,
	} {
		if _, err := conn.Exec(stmt); err != nil {
			t.Fatalf("%s: %v", stmt, err)
		}
	}
	database := chatJetDatabase(filename)

	names := chatNameLists(context.Background(), database)
	for _, want := range []string{"Konton: ", "Lönekonto", "Utgiftskategorier: ", "Livsmedel"} {
		if !strings.Contains(names, want) {
			t.Errorf("namnen saknar %q: %q", want, names)
		}
	}

	result, err := database.Query(context.Background(),
		"SELECT TillKonto, Belopp FROM Transaktioner WHERE Vad LIKE '%livsmedel%' AND Datum >= '2026-01-01' AND Datum <= '2026-12-31'")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rows) != 1 || result.Rows[0][0] != "Räkmacka" || result.Rows[0][1] != "1 500,00" {
		t.Errorf("fel resultat: %q", result.Rows)
	}
	t.Logf("kolumner %q, rader %q", result.Columns, result.Rows)

	// Rader som nyss skrivits på programmets anslutning ska synas i kopian.
	for i := range 5 {
		if _, err := conn.Exec(fmt.Sprintf(`INSERT INTO Konton (Benämning) VALUES ('Nytt%d')`, i)); err != nil {
			t.Fatal(err)
		}
		fresh, err := database.Query(context.Background(), "SELECT COUNT(*) FROM Konton WHERE Benämning LIKE 'Nytt%'")
		if err != nil {
			t.Fatal(err)
		}
		if want := strconv.Itoa(i + 1); fresh.Rows[0][0] != want {
			t.Errorf("runda %d: kopian ser %q rader", i, fresh.Rows[0][0])
		}
	}

	before := countRows(t, conn, "Transaktioner")
	accounts := countRows(t, conn, "Konton")
	for _, query := range []string{
		"DELETE FROM Transaktioner",
		"UPDATE Konton SET Benämning = 'x'",
		"INSERT INTO Konton (Benämning) VALUES ('x')",
		"DROP TABLE Personer",
		"CREATE TABLE Ny (a TEXT)",
		"SELECT * INTO Kopia FROM Konton",
	} {
		// Drivrutinen kan köra satsen på kopian och sedan ge fel. Det viktiga är originalet.
		_, _ = database.Query(context.Background(), query)
	}
	if after := countRows(t, conn, "Transaktioner"); after != before {
		t.Errorf("Transaktioner ändrades: %d -> %d", before, after)
	}
	if after := countRows(t, conn, "Konton"); after != accounts {
		t.Errorf("Konton ändrades: %d -> %d", accounts, after)
	}
	if n := countRows(t, conn, "Personer"); n < 1 {
		t.Errorf("Personer: %d", n)
	}

	// Kopian ska vara borta efteråt.
	left, err := filepath.Glob(filepath.Join(os.TempDir(), "wHHEKchat*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(left) != 0 {
		t.Errorf("tillfälliga kopior finns kvar: %q", left)
	}
}

// Hela vägen: modellen svarar med SQL som körs mot en riktig databas.
func TestChatAskAgainstRealDatabase(t *testing.T) {
	t.Parallel()
	database, conn := newSQLiteChatDatabase(t, "real.db")
	if _, err := conn.Exec(`INSERT INTO Transaktioner (FrånKonto,TillKonto,Typ,Datum,Vad,Vem,Belopp) VALUES
		('Lön','ICA','Inköp','2025-03-02','Livsmedel','Gemensamt',500.25),
		('Lön','<b>Coop</b>','Inköp','2025-04-02','Livsmedel','Gemensamt',700)`); err != nil {
		t.Fatal(err)
	}
	budgetRows := countRows(t, conn, "Budget")

	ask := func(reply string) string {
		srv := fakeLLM(t, func(w http.ResponseWriter, _ *http.Request) { llmReply(w, reply) })
		form := url.Values{"question": {"Fråga"}}
		req := httptest.NewRequest(http.MethodPost, "/chatask", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		serveChatAsk(rec, req, srv.URL, database)
		return rec.Body.String()
	}

	body := ask("```sql\nSELECT SUM(Belopp) AS Summa FROM Transaktioner\n```")
	if !strings.Contains(body, "<th>Summa</th>") || !strings.Contains(body, "<td>1 200,25</td>") {
		t.Errorf("fel resultat:\n%s", body)
	}
	saved, err := chatParseHistory(historyField(t, body))
	if err != nil {
		t.Fatal(err)
	}
	if len(saved) != 2 || saved[1].Content != "SELECT SUM(Belopp) AS Summa FROM Transaktioner" {
		t.Errorf("samtalet ska innehålla SQL men inte resultatet: %+v", saved)
	}

	// Data ur databasen ska escapas.
	body = ask("SELECT TillKonto FROM Transaktioner WHERE Belopp = 700")
	if !strings.Contains(body, "&lt;b&gt;Coop") || strings.Contains(body, "<b>Coop") {
		t.Errorf("resultatet ska escapas:\n%s", body)
	}

	body = ask("SELECT * FROM Transaktioner WHERE Vad = 'Finns inte'")
	if !strings.Contains(body, "Inga rader hittades") {
		t.Errorf("väntade meddelande om tomt resultat:\n%s", body)
	}

	// Frågor som stoppas: felet och SQL-frågan visas, frågan står kvar och inget ändras.
	for _, reply := range []string{"DELETE FROM Budget", "SELECT * FROM DtbVer", "SELECT Finns_inte FROM Budget", "Jag kan inte svara på det."} {
		body = ask(reply)
		if !strings.Contains(body, `class="chat_error"`) || !strings.Contains(body, "chat_sql") {
			t.Errorf("%q: väntade fel och SQL:\n%s", reply, body)
		}
		if strings.Contains(body, "hx-swap-oob") {
			t.Errorf("%q: frågan ska stå kvar:\n%s", reply, body)
		}
	}
	if n := countRows(t, conn, "Budget"); n != budgetRows {
		t.Errorf("Budget ändrades: %d -> %d", budgetRows, n)
	}

	body = ask("INGEN_SQL")
	if !strings.Contains(body, "går inte att besvara") || strings.Contains(body, "chat_sql") {
		t.Errorf("väntade meddelande utan SQL:\n%s", body)
	}
}

// historyField plockar ut samtalet ur det dolda fältet i svaret, på samma sätt
// som webbläsaren gör.
func historyField(t *testing.T, body string) string {
	t.Helper()
	_, rest, found := strings.Cut(body, `id="chat_history" name="history" value="`)
	if !found {
		t.Fatalf("svaret saknar historik:\n%s", body)
	}
	value, _, _ := strings.Cut(rest, `"`)
	return html.UnescapeString(value)
}

func TestChatAskKeepsHistory(t *testing.T) {
	t.Parallel()
	var got []chatMessage
	srv := fakeLLM(t, func(w http.ResponseWriter, req *http.Request) {
		var payload struct {
			Messages []chatMessage `json:"messages"`
		}
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			t.Error(err)
		}
		got = payload.Messages[1:] // Utan bakgrunden
		llmReply(w, fmt.Sprintf(`SELECT '"%d" <&>'`, len(got)))
	})

	first := postChat(t, srv.URL, url.Values{"question": {`Fråga "1"`}})
	second := postChat(t, srv.URL, url.Values{"question": {"Fråga 2"}, "history": {historyField(t, first)}})

	want := []chatMessage{
		{Role: "user", Content: `Fråga "1"`},
		{Role: "assistant", Content: `SELECT '"1" <&>'`},
		{Role: "user", Content: "Fråga 2"},
	}
	if !slices.Equal(got, want) {
		t.Errorf("fel samtal till LLM-servern:\n got %+v\nwant %+v", got, want)
	}

	if strings.Contains(historyField(t, second), `"system"`) {
		t.Error("bakgrunden ska inte sparas i samtalet")
	}
	saved, err := chatParseHistory(historyField(t, second))
	if err != nil {
		t.Fatal(err)
	}
	want = append(want, chatMessage{Role: "assistant", Content: `SELECT '"3" <&>'`})
	if !slices.Equal(saved, want) {
		t.Errorf("fel sparat samtal:\n got %+v\nwant %+v", saved, want)
	}
}

func TestChatAskBadHistory(t *testing.T) {
	t.Parallel()
	srv := fakeLLM(t, func(http.ResponseWriter, *http.Request) {
		t.Error("frågan ska inte skickas när historiken är trasig")
	})

	body := postChat(t, srv.URL, url.Values{"question": {"Fråga"}, "history": {"inte json"}})

	if !strings.Contains(body, "Rensa historik") {
		t.Errorf("väntade meddelande:\n%s", body)
	}
}

func TestChatPageHasClearHistory(t *testing.T) {
	t.Parallel()
	srv := fakeLLM(t, func(http.ResponseWriter, *http.Request) {})

	rec := httptest.NewRecorder()
	serveChat(rec, httptest.NewRequest(http.MethodGet, "/chat", nil), srv.URL, fakeDatabase)
	body := rec.Body.String()

	if !strings.Contains(body, `<a href="chat">Rensa historik</a>`) {
		t.Errorf("sidan saknar Rensa historik:\n%s", body)
	}
	if historyField(t, body) != "" {
		t.Error("en ny sida ska ha tom historik")
	}
}

func TestChatAskEmptyQuestion(t *testing.T) {
	t.Parallel()
	srv := fakeLLM(t, func(http.ResponseWriter, *http.Request) {
		t.Error("tom fråga ska inte skickas till LLM-servern")
	})

	body := postQuestion(t, srv.URL, "   ")

	if !strings.Contains(body, "Skriv en fråga") {
		t.Errorf("väntade meddelande:\n%s", body)
	}
}

func TestChatAskServerErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		handler http.HandlerFunc
		want    string
	}{
		{"status 500", func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "modellen kraschade", http.StatusInternalServerError)
		}, "status 500: modellen kraschade"},
		{"fel med status 200", func(w http.ResponseWriter, _ *http.Request) {
			_, _ = io.WriteString(w, `{"error":"No models loaded"}`)
		}, "No models loaded"},
		{"inte JSON", func(w http.ResponseWriter, _ *http.Request) {
			_, _ = io.WriteString(w, "inte json")
		}, "oväntat svar"},
		{"tomt svar", func(w http.ResponseWriter, _ *http.Request) {
			llmReply(w, "  ")
		}, "tomt svar"},
		{"tomt svar utan utrymme", func(w http.ResponseWriter, _ *http.Request) {
			_, _ = io.WriteString(w, `{"choices":[{"message":{"content":""},"finish_reason":"length"}]}`)
		}, "slut på utrymme"},
		{"för lång fråga LM Studio", func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, lmStudioContextError, http.StatusBadRequest)
		}, "13243 tokens, men det finns bara plats för 8192"},
		{"för lång fråga llama.cpp", func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, `{"error":{"code":400,"type":"exceed_context_size_error","n_prompt_tokens":13243,"n_ctx":8192}}`, http.StatusBadRequest)
		}, "13243 tokens, men det finns bara plats för 8192"},
		{"för lång fråga utan storlekar", func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, `{"error":"exceed_context_size_error"}`, http.StatusBadRequest)
		}, "Frågan är för lång"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			body := postQuestion(t, fakeLLM(t, tc.handler).URL, "Fråga")

			if !strings.Contains(body, tc.want) {
				t.Errorf("väntade %q:\n%s", tc.want, body)
			}
			if strings.Contains(body, "hx-swap-oob") {
				t.Errorf("frågan och historiken ska stå kvar i sidan vid fel:\n%s", body)
			}
		})
	}
}

// Felet ska inte beskrivas som att servern saknas när den bara inte kan svara.
func TestChatAskProblemIsNotServerError(t *testing.T) {
	t.Parallel()
	srv := fakeLLM(t, func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, lmStudioContextError, http.StatusBadRequest)
	})

	body := postQuestion(t, srv.URL, "Fråga")

	if strings.Contains(body, "Chat kräver") {
		t.Errorf("ska inte visa krav på servern:\n%s", body)
	}
}

func TestChatAskTruncatedReply(t *testing.T) {
	t.Parallel()
	srv := fakeLLM(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"choices":[{"message":{"content":"SELECT Vad FROM Trans"},"finish_reason":"length"}]}`)
	})

	body := postQuestion(t, srv.URL, "Fråga")

	if !strings.Contains(body, "slut på utrymme") || strings.Contains(body, "resultatet") {
		t.Errorf("en kapad SQL-fråga ska inte köras:\n%s", body)
	}
	if strings.Contains(body, "hx-swap-oob") {
		t.Errorf("frågan ska stå kvar:\n%s", body)
	}
}

func TestChatAskNormalResultHasNoNotice(t *testing.T) {
	t.Parallel()
	srv := fakeLLM(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"choices":[{"message":{"content":"SELECT 1"},"finish_reason":"stop"}]}`)
	})

	if body := postQuestion(t, srv.URL, "Fråga"); strings.Contains(body, `class="chat_notice"`) {
		t.Errorf("väntade inget meddelande:\n%s", body)
	}
}

// Fake-servern anger kontextlängden 10000 tokens, så samtalet med bakgrund får bli ungefär 5000.
func TestChatAskTrimsOldestWhenLong(t *testing.T) {
	t.Parallel()
	var got []chatMessage
	srv := fakeLLM(t, func(w http.ResponseWriter, req *http.Request) {
		var payload struct {
			Messages []chatMessage `json:"messages"`
		}
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			t.Error(err)
		}
		got = payload.Messages[1:] // Utan bakgrunden
		llmReply(w, "SELECT 1")
	})

	var history []chatMessage
	for i := range 10 {
		history = append(history,
			chatMessage{Role: "user", Content: fmt.Sprintf("Fråga%c %s", 'A'+i, strings.Repeat("x", 3000))},
			chatMessage{Role: "assistant", Content: strings.Repeat("y", 3000)})
	}
	encoded, err := json.Marshal(history)
	if err != nil {
		t.Fatal(err)
	}

	body := postChat(t, srv.URL, url.Values{"question": {"Sista frågan"}, "history": {string(encoded)}})

	if got[len(got)-1].Content != "Sista frågan" || got[len(got)-2].Role != "assistant" {
		t.Errorf("den nya frågan ska komma sist, efter senaste svaret: %+v", got[len(got)-2:])
	}
	if !strings.HasPrefix(got[0].Content, "Fråga") || got[0].Role != "user" {
		t.Errorf("samtalet ska börja med en fråga: %+v", got[0])
	}
	sent := len(got) / 2
	if sent >= 10 || sent == 0 {
		t.Fatalf("antal frågor kvar: %d", sent)
	}
	if !strings.HasPrefix(got[len(got)-3].Content, "FrågaJ") {
		t.Errorf("den senaste frågan ska finnas kvar: %q", got[len(got)-3].Content[:10])
	}
	if want := fmt.Sprintf("modellen minns: %d.", sent+1); !strings.Contains(body, want) {
		t.Errorf("väntade %q:\n%s", want, body)
	}

	saved, err := chatParseHistory(historyField(t, body))
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(saved[:len(got)], got) {
		t.Error("det sparade samtalet ska vara det som skickades, utan de borttagna frågorna")
	}
}

func TestChatEstimateTokens(t *testing.T) {
	t.Parallel()
	if got := chatEstimateTokens("123"); got != 3 {
		t.Errorf("siffror: %d", got)
	}
	if got := chatEstimateTokens("åäöåäö"); got != 2 {
		t.Errorf("text ska räknas i tecken, inte byte: %d", got)
	}
}

func TestChatTrimHistory(t *testing.T) {
	t.Parallel()
	pair := func(s string) []chatMessage {
		return []chatMessage{{Role: "user", Content: s}, {Role: "assistant", Content: s}}
	}
	long := strings.Repeat("1", 100) // 106 tokens per meddelande med tillägget
	history := slices.Concat(pair("a"+long), pair("b"+long))

	tests := []struct {
		name        string
		question    string
		context     int
		wantDropped int
		wantKept    int
	}{
		{"ryms", "", 1000, 0, 4},
		{"äldsta bort", "", 500, 1, 2},
		{"allt bort", "", 100, 2, 0},
		{"lång fråga tränger undan", strings.Repeat("1", 200), 400, 2, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			kept, dropped := chatTrimHistory(history, tc.question, "", tc.context)
			if dropped != tc.wantDropped || len(kept) != tc.wantKept {
				t.Errorf("dropped=%d kept=%d", dropped, len(kept))
			}
			if tc.wantKept == 2 && !strings.HasPrefix(kept[0].Content, "b") {
				t.Errorf("den senaste frågan ska finnas kvar: %q", kept[0].Content[:5])
			}
		})
	}
}

func TestChatTrimHistoryCountsBackground(t *testing.T) {
	t.Parallel()
	history := []chatMessage{
		{Role: "user", Content: strings.Repeat("1", 100)},
		{Role: "assistant", Content: strings.Repeat("1", 100)},
	}
	if _, dropped := chatTrimHistory(history, "", "", 1000); dropped != 0 {
		t.Errorf("utan bakgrund: dropped=%d", dropped)
	}
	if _, dropped := chatTrimHistory(history, "", strings.Repeat("1", 700), 1000); dropped != 1 {
		t.Errorf("med bakgrund: dropped=%d", dropped)
	}
	// Med riktig bakgrund och okänd kontextlängd ska det gå att ställa följdfrågor.
	pair := []chatMessage{{Role: "user", Content: "Hur mycket lade jag på mat?"}, {Role: "assistant", Content: "SELECT SUM(Belopp) FROM Transaktioner"}}
	background := chatSystemPrompt("SQLite", time.Now(), "Konton: Lönekonto\n")
	if _, dropped := chatTrimHistory(pair, "Och i fjol?", background, llmDefaultContextTokens); dropped != 0 {
		t.Errorf("följdfråga vid %d tokens: dropped=%d", llmDefaultContextTokens, dropped)
	}
}

// Beskrivningen av databasen ska stämma med tabellerna som InitiateDB skapar.
func TestChatBackgroundMatchesDatabase(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "schema.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	InitiateDB(db)

	actual := map[string]string{}
	tables, err := db.Query("SELECT name FROM sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite_%'")
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for tables.Next() {
		var name string
		if err := tables.Scan(&name); err != nil {
			t.Fatal(err)
		}
		names = append(names, name)
	}
	_ = tables.Close()
	for _, name := range names {
		rows, err := db.Query("SELECT name, type FROM pragma_table_info(?)", name)
		if err != nil {
			t.Fatal(err)
		}
		var columns []string
		for rows.Next() {
			var column, columnType string
			if err := rows.Scan(&column, &columnType); err != nil {
				t.Fatal(err)
			}
			columns = append(columns, column+" "+strings.ToUpper(columnType))
		}
		_ = rows.Close()
		actual[name] = strings.TrimSpace(strings.Join(columns, ", "))
	}
	if len(actual) == 0 {
		t.Fatal("inga tabeller i databasen")
	}

	described := map[string]string{}
	for _, m := range regexp.MustCompile(`(?m)^CREATE TABLE (\S+) \((.*)\);$`).FindAllStringSubmatch(chatBackground, -1) {
		columns := strings.Split(strings.ReplaceAll(m[2], "DECIMAL(19,4)", "DECIMAL(19#4)"), ", ")
		for i, column := range columns {
			columns[i] = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(column, "#", ","), " PRIMARY KEY", ""))
		}
		described[m[1]] = strings.Join(columns, ", ")
	}

	for name, columns := range actual {
		if described[name] != strings.ReplaceAll(columns, " ,", ",") {
			t.Errorf("tabell %s:\n databas     %q\n beskrivning %q", name, columns, described[name])
		}
	}
	for name := range described {
		if _, found := actual[name]; !found {
			t.Errorf("tabell %s finns inte i databasen", name)
		}
	}
}

func TestLLMContextTokens(t *testing.T) {
	t.Parallel()
	srv := fakeLLM(t, func(http.ResponseWriter, *http.Request) {})
	ctx := context.Background()

	if got := llmContextTokens(ctx, srv.URL, "chatmodellen"); got != 10000 {
		t.Errorf("laddad modell: %d", got)
	}
	if got := llmContextTokens(ctx, srv.URL, "text-embedding-nomic"); got != llmDefaultContextTokens {
		t.Errorf("modell som inte är laddad: %d", got)
	}
	if got := llmContextTokens(ctx, closedServerURL(), "chatmodellen"); got != llmDefaultContextTokens {
		t.Errorf("ingen server: %d", got)
	}
}

// fakeLMStudio beter sig som LM Studios API för att ladda modeller.
type fakeLMStudio struct {
	*httptest.Server
	mu        sync.Mutex
	loaded    []int // Kontextlängd för varje laddad instans
	calls     []string
	failLoads map[int]bool // Kontextlängder som inte går att ladda
	roundUpTo int          // LM Studio avrundar längden uppåt
}

func newFakeLMStudio(t *testing.T, loaded ...int) *fakeLMStudio {
	t.Helper()
	f := &fakeLMStudio{loaded: loaded, failLoads: map[int]bool{}}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/models", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"data":[{"id":"chatmodellen"}]}`)
	})
	mux.HandleFunc("GET /api/v1/models", func(w http.ResponseWriter, _ *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		instances := []map[string]any{}
		for i, contextLength := range f.loaded {
			instances = append(instances, map[string]any{
				"id":     fmt.Sprintf("instans%d", i),
				"config": map[string]int{"context_length": contextLength},
			})
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"models": []map[string]any{
			{"key": "annan", "max_context_length": 2048, "loaded_instances": []map[string]any{}},
			{"key": "chatmodellen", "max_context_length": 131072, "loaded_instances": instances},
		}})
	})
	mux.HandleFunc("POST /api/v1/models/unload", func(w http.ResponseWriter, req *http.Request) {
		var payload struct {
			InstanceID string `json:"instance_id"`
		}
		_ = json.NewDecoder(req.Body).Decode(&payload)
		f.mu.Lock()
		defer f.mu.Unlock()
		f.calls = append(f.calls, "unload "+payload.InstanceID)
		f.loaded = f.loaded[1:]
		_, _ = fmt.Fprintf(w, `{"instance_id":%q}`, payload.InstanceID)
	})
	mux.HandleFunc("POST /api/v1/models/load", func(w http.ResponseWriter, req *http.Request) {
		var payload struct {
			Model         string `json:"model"`
			ContextLength int    `json:"context_length"`
		}
		_ = json.NewDecoder(req.Body).Decode(&payload)
		f.mu.Lock()
		defer f.mu.Unlock()
		f.calls = append(f.calls, fmt.Sprintf("load %s %d", payload.Model, payload.ContextLength))
		if f.failLoads[payload.ContextLength] {
			http.Error(w, `{"error":"insufficient memory"}`, http.StatusInternalServerError)
			return
		}
		length := payload.ContextLength
		if f.roundUpTo > 0 {
			length = (length + f.roundUpTo - 1) / f.roundUpTo * f.roundUpTo
		}
		f.loaded = append(f.loaded, length)
		_, _ = io.WriteString(w, `{"status":"loaded"}`)
	})
	f.Server = httptest.NewServer(mux)
	t.Cleanup(f.Close)
	return f
}

func (f *fakeLMStudio) checkCalls(t *testing.T, want ...string) {
	t.Helper()
	f.mu.Lock()
	defer f.mu.Unlock()
	if !slices.Equal(f.calls, want) {
		t.Errorf("fel anrop:\n got %q\nwant %q", f.calls, want)
	}
}

func TestLLMEnsureContextReloadsWithHalfOfMax(t *testing.T) {
	t.Parallel()
	f := newFakeLMStudio(t, 8192)

	got, err := llmEnsureContext(context.Background(), f.URL, "chatmodellen")

	if err != nil || got != 65536 {
		t.Errorf("got %d, %v", got, err)
	}
	f.checkCalls(t, "unload instans0", "load chatmodellen 65536")
}

func TestLLMEnsureContextPercent(t *testing.T) {
	t.Parallel()
	for percent, want := range map[int]int{25: 32768, 33: 43253, 100: 131072, 1: 1310} {
		f := newFakeLMStudio(t, 8192)

		got, err := llmEnsureContextPercent(context.Background(), f.URL, "chatmodellen", percent)

		if err != nil || got != want {
			t.Errorf("%d%%: got %d, %v", percent, got, err)
		}
		f.checkCalls(t, "unload instans0", fmt.Sprintf("load chatmodellen %d", want))

		// Andra gången stämmer längden, så modellen ska inte laddas om.
		if _, err := llmEnsureContextPercent(context.Background(), f.URL, "chatmodellen", percent); err != nil {
			t.Error(err)
		}
		f.checkCalls(t, "unload instans0", fmt.Sprintf("load chatmodellen %d", want))
	}
}

// LM Studio avrundar längden uppåt. Det får inte leda till omladdning varje gång.
func TestLLMEnsureContextServerRounds(t *testing.T) {
	t.Parallel()
	f := newFakeLMStudio(t, 8192)
	f.roundUpTo = 256

	for range 3 {
		got, err := llmEnsureContextPercent(context.Background(), f.URL, "chatmodellen", 33)
		if err != nil || got != 43264 {
			t.Errorf("got %d, %v", got, err)
		}
	}
	f.checkCalls(t, "unload instans0", "load chatmodellen 43253")
}

func TestLLMCheckOptions(t *testing.T) {
	t.Parallel()
	for _, ok := range [][2]int{{1234, 50}, {1, 1}, {65535, 100}} {
		if err := llmCheckOptions(ok[0], ok[1]); err != nil {
			t.Errorf("%v: %v", ok, err)
		}
	}
	for _, bad := range [][2]int{{0, 50}, {-1, 50}, {65536, 50}, {1234, 0}, {1234, -5}, {1234, 101}} {
		if err := llmCheckOptions(bad[0], bad[1]); err == nil {
			t.Errorf("%v ska inte godkännas", bad)
		}
	}
}

func TestLLMURLForPort(t *testing.T) {
	t.Parallel()
	if got := llmURLForPort(4321); got != "http://localhost:4321" {
		t.Errorf("got %q", got)
	}
}

func TestLLMEnsureContextAlreadyCorrect(t *testing.T) {
	t.Parallel()
	f := newFakeLMStudio(t, 65536)

	got, err := llmEnsureContext(context.Background(), f.URL, "chatmodellen")

	if err != nil || got != 65536 {
		t.Errorf("got %d, %v", got, err)
	}
	f.checkCalls(t)
}

func TestLLMEnsureContextNotLoaded(t *testing.T) {
	t.Parallel()
	f := newFakeLMStudio(t)

	if _, err := llmEnsureContext(context.Background(), f.URL, "chatmodellen"); err != nil {
		t.Error(err)
	}
	f.checkCalls(t, "load chatmodellen 65536")
}

func TestLLMEnsureContextLoadFails(t *testing.T) {
	t.Parallel()
	f := newFakeLMStudio(t, 8192)
	f.failLoads[65536] = true

	got, err := llmEnsureContext(context.Background(), f.URL, "chatmodellen")
	if err == nil || !strings.Contains(err.Error(), "insufficient memory") || got != 8192 {
		t.Errorf("got %d, %v", got, err)
	}
	f.checkCalls(t, "unload instans0", "load chatmodellen 65536", "load chatmodellen 8192")

	// Modellen får inte laddas om igen vid nästa sidvisning eller fråga.
	got, err = llmEnsureContext(context.Background(), f.URL, "chatmodellen")
	if err == nil || got != 8192 {
		t.Errorf("andra försöket: got %d, %v", got, err)
	}
	f.checkCalls(t, "unload instans0", "load chatmodellen 65536", "load chatmodellen 8192")

	rec := httptest.NewRecorder()
	serveChat(rec, httptest.NewRequest(http.MethodGet, "/chat", nil), f.URL, fakeDatabase)
	body := rec.Body.String()
	if !strings.Contains(body, "insufficient memory") || !strings.Contains(body, `id="question"`) {
		t.Errorf("sidan ska visa varningen och ändå gå att använda:\n%s", body)
	}
}

// En webbläsare som lämnar sidan får inte lämna modellen urladdad.
func TestLLMEnsureContextSurvivesCancelledRequest(t *testing.T) {
	t.Parallel()
	f := newFakeLMStudio(t, 8192)
	ctx, cancel := context.WithCancel(context.Background())
	unloaded := make(chan struct{})

	// Avbryt begäran precis när modellen laddas ur.
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		f.Config.Handler.ServeHTTP(w, req)
		if strings.HasSuffix(req.URL.Path, "/unload") {
			cancel()
			close(unloaded)
		}
	}))
	t.Cleanup(proxy.Close)

	if _, err := llmEnsureContext(ctx, proxy.URL, "chatmodellen"); err != nil {
		t.Error(err)
	}
	<-unloaded
	f.checkCalls(t, "unload instans0", "load chatmodellen 65536")
}

func TestLLMEnsureContextOtherServers(t *testing.T) {
	t.Parallel()
	// fakeLLM saknar LM Studios API, som llama.cpp.
	srv := fakeLLM(t, func(http.ResponseWriter, *http.Request) {})

	got, err := llmEnsureContext(context.Background(), srv.URL, "chatmodellen")

	if err != nil || got != 0 {
		t.Errorf("got %d, %v", got, err)
	}
}

func TestChatPageShowsContextLength(t *testing.T) {
	t.Parallel()
	f := newFakeLMStudio(t, 65536)

	rec := httptest.NewRecorder()
	serveChat(rec, httptest.NewRequest(http.MethodGet, "/chat", nil), f.URL, fakeDatabase)

	if !strings.Contains(rec.Body.String(), "kontextlängd 65536 tokens") {
		t.Errorf("sidan saknar kontextlängden:\n%s", rec.Body.String())
	}
}

func TestChatAskServerGone(t *testing.T) {
	t.Parallel()

	body := postQuestion(t, closedServerURL(), "Fråga")

	if !strings.Contains(body, "dial tcp") || !strings.Contains(body, "LM Studio") {
		t.Errorf("väntade felmeddelande med krav på server:\n%s", body)
	}
}

func TestChatAskRequiresPost(t *testing.T) {
	t.Parallel()
	rec := httptest.NewRecorder()
	serveChatAsk(rec, httptest.NewRequest(http.MethodGet, "/chatask?question=x", nil), closedServerURL(), fakeDatabase)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status %d", rec.Code)
	}
}

// Webbservern i main() har korta tidsgränser, men ett svar från en språkmodell
// tar ofta längre tid än så.
func TestChatAskOutlivesServerTimeouts(t *testing.T) {
	t.Parallel()
	llm := fakeLLM(t, func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(500 * time.Millisecond)
		llmReply(w, "SELECT 1")
	})

	app := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		serveChatAsk(w, req, llm.URL, fakeDatabase)
	}))
	app.Config.ReadTimeout = 100 * time.Millisecond
	app.Config.WriteTimeout = 100 * time.Millisecond
	app.Start()
	t.Cleanup(app.Close)

	resp, err := http.PostForm(app.URL, url.Values{"question": {"Fråga"}})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "<td>resultatet</td>") {
		t.Errorf("svaret kom inte fram:\n%s", body)
	}
}
