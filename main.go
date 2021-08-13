//-*- coding: utf-8 -*-

// This is a personal finance package, inspired by Hogia Hemekonomi.
// Detta är ett hemekonomiprogram, inspirerat av Hogia Hemekonomi från 90-talet.

// System Requirements: Windows 10 (any), Endast om mdb/JetDB-filer ska hanteras
//                      Any, om endast sqlite

// To build on Windows:
// Prepare: install gnu emacs: emacs-26.3-x64_64 (optional)
// Prepare: TDM-GCC from https://jmeubank.github.io/tdm-gcc/
//https://github.com/jmeubank/tdm-gcc/releases/download/v9.2.0-tdm-1/tdm-gcc-9.2.0.exe

// Prepare: install git: Git-2.23.0-64-bit
// Prepare: install golang 32-bits (can't access access/jet/mdb driver using 64-bits)
//   go1.16.3.windows-386.msi
// Build development version: ./make.bat
// Build release version: ./make.bat release
// Run unit tests: ./make.bat test

// Run: ./wHHEK.exe -help
// Run: ./wHHEK.exe -optin=.

// Allow access if Windows Firewall asks. Only allow access from localhost
// if possible.

// Existerande funktioner
// ======================
// Välj databas att hantera. Hanterar både mdb (endast på Windows) från Hogia Hemekonomi och konverterad till sqlite med hhek2sqlite
// Visa konton (enkel HTML)
// Visa kontoutdrag per månad för ett konto med saldo (enkel HTML)
// Visa transaktioner, vald tidsperiod (enkel HTML)
// Visa transaktioner, sök text i kommentar (enkel HTML)
// Visa graf på saldo under månad (enkel HTML)
// Registrera transaktion (enkel HTML)
// Visa platser
// Lägg till ny plats
// Redigera plats
// Visa personer
// Lägg till ny person
// Redigera person
// Lägg till nytt konto
// Redigera konto
// Hantera fasta överföringar/betalningar som passerat datum
// Hantera fasta överföringar/betalningar för hela aktuell månad
// Visa budget
// Redigera budget

// ROADMAP/TODO/The Future Is In Flux
// ============
// escape & in all html. escapeHTML verkar inte fungera. Use template?
// hantera fel: för lång text till comment
// kommandorads help-option
// kommandoradsoption för att välja katalog med databas
// kommandoradsoption för att välja databas
// kommandoradsoption för att sätta portnummer
// kommandoradsoption för att lägga till transaktion
// startscript: starta med rätt argument samt starta webläsare
// Installationsinstruktion (ladda ner exe, skapa ikon, brandvägg)
// Efter lagt till transaktion, visa den tillagda
// Visa transaktioner, filter: Frånkonto
// Visa transaktioner, filter: Tillkonto
// Visa transaktioner, filter: Summa
// Visa transaktioner, filter: Plats
// Visa transaktioner, filter: Person
// Visa transaktioner, filter: Vad
// Visa resultat-tabellen, aktuell/vald månad
// Visa resultat-tabellen, helår
// Visa resultat-tabellen, delår till och med aktuell månad
// Graf som i månadsvyn fast för senaste året
// Visa fasta överföringar
// Lägg till fast överföring
// Redigera fast överföring
// Visa fasta betalningar
// Lägg till fast betalning
// Redigera fast betalning
// Registrera fasta överföringar
// Registrera fasta betalningar
// Skapa ny fil (sqlite), kompatibel
// Skapa budget enligt Konsumentverkets riktlinjer
// REST-api: visa transaktioner
// REST-api: månadskontoutdrag
// REST-api: lägg till transaktion
// REST-api: redigera transaktion
// REST-api: visa/lägg till/redigera platser
// REST-api: visa/lägg till/redigera konton
// REST-api: visa/lägg till/redigera personer
// REST-api: visa/lägg till/redigera fasta överföringar
// REST-api: visa/lägg till/redigera fasta betalningar
// REST-api: registrera överföringar
// REST-api: registrera betalningar
// kommandorads "api"
// REST-api: visa/redigera budget
// Visa budget med jämförelse till resultat
// REST-api: Visa budget med jämförelse till resultat
// Byt till https/SSL
// Kräv inloggning
// Årsskiftesrutin inkl uppdatera budget på olika sätt
// Graf som i månadsvyn, med valfri grupp av konton
// Graf som i månadsvyn fast för senaste året, med valfri grupp av konton
// Visa lån
// Lägg till nytt lån
// Redigera lån

// Testa kompabilitet Linux/Mac (endast med sqlite-databas)
//   Alternate use of: go build -tags withoutODBC
//                    / /  + b u i l d withoutODBC
// Experimental:
// Build on Windows for Linux:
// In powershell:
// $env:GOOS="linux"
// $env:GOARCH="386"
// go build -o wHHEK.elf32 main.go nojetdb.go platser.go transaktioner.go fastatransaktioner.go personer.go konton.go budget.go


// Notes/references/hints for further development
// TODO: https://stackoverflow.com/questions/26345318/how-can-i-prevent-sql-injection-attacks-in-go-while-using-database-sql
// TODO: https://www.calhoun.io/what-is-sql-injection-and-how-do-i-avoid-it-in-go/
// TODO: https://golang.org/pkg/html/template/
// TODO: hasExts from https://stackoverflow.com/questions/45586944/case-insensitive-hassuffix-in-go
// TODO: https://github.com/stripe/safesql
// TODO: https://conroy.org/introducing-sqlc

package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"golang.org/x/text/encoding/charmap"
	"io/ioutil"
	"log"
	"html"
	"math"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"  // MIT License
	"github.com/shopspring/decimal"  // MIT License
)

// Global variables
var db *sql.DB = nil
var dbtype uint8 = 0 // 0=none, 1=mdb/Access2.0, 2=sqlite3
var currentDatabase string = "NONE"
var (
	ctx context.Context
)
var srv *http.Server;

func toUtf8(in_buf []byte) string {
	buf := in_buf
	if dbtype == 1 {
		buf, _ = charmap.Windows1252.NewDecoder().Bytes(in_buf)
	}
	// Escape chars for SQL
	stringVal := string(buf)
	stringVal2 := strings.ReplaceAll(stringVal, "'", "''")
	stringVal3 := strings.ReplaceAll(stringVal2, "\"", "\"\"")
	return stringVal3
}

func unEscapeSQL(in_buf string) string {
	// UnEscape chars for SQL
	stringVal2 := strings.ReplaceAll(in_buf, "''", "'")
	stringVal3 := strings.ReplaceAll(stringVal2, "\"\"", "\"")
	return stringVal3
}

func hello(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "hello\n")
}

func root(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "<html>\n")
	fmt.Fprintf(w, "<body>\n")
	fmt.Fprintf(w, "Välj databas att arbeta med:<br>")

	files, err := ioutil.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}

	if len(files) > 0 {
		fmt.Fprintf(w, "<form method=\"POST\" action=\"/open\">\n")
		for _, file := range files {

			if strings.HasSuffix(strings.ToLower(file.Name()), ".mdb") ||
				strings.HasSuffix(strings.ToLower(file.Name()), ".db") {
				//fmt.Fprintf(w,"%s<br>\n", file.Name())
				fmt.Fprintf(w, "<input type=\"radio\" id=\"%s\" name=\"fname\" value=\"%s\"><label for=\"%s\">%s</label><br>\n", file.Name(), file.Name(), file.Name(), file.Name())
			}
		}
		fmt.Fprintf(w, "<input type=\"submit\" value=\"Submit\"></form>\n")
	} else {
		fmt.Fprintf(w, "No files available.<p>\n")
	}
	//fmt.Fprintf(w, "<p>See also <a href=\"hello\">link hello</a><br>\n")
	//fmt.Fprintf(w, "See also <a href=\"r\">link r</a><br>\n")
	//fmt.Fprintf(w, "See also <a href=\"headers\">link headers</a><br>\n")
	fmt.Fprintf(w, "</body>\n")
	fmt.Fprintf(w, "</html>\n")
}

func restapi(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Rest API not implemented yet.\n")
}

func headers(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Host: %s\n\n", req.Host)

	for name, headers := range req.Header {
		for _, h := range headers {
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}

func sanitizeFilename(fname string) string {
	fname = strings.Replace(fname, "\\", "", -1)
	fname = strings.Replace(fname, "/", "", -1)
	fname = strings.Replace(fname, "'", "", -1)
	fname = strings.Replace(fname, "<", "", -1)
	fname = strings.Replace(fname, ">", "", -1)
	fname = strings.Replace(fname, "\"", "", -1)

	return fname
}

func openSqlite(filename string) *sql.DB {
	currentDatabase = "NONE"
	dbtype = 0

	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		log.Fatal(err)
	}

	currentDatabase = filename
	dbtype = 2

	return db
}

func GetCountPendingÖverföringar(db *sql.DB, currDate string) int {
	var cnt int
	_ = db.QueryRow(`select count(*) from Överföringar WHERE Datum <= ?`, currDate).Scan(&cnt)
	return cnt
}

func checkÖverföringar(w http.ResponseWriter, db *sql.DB) {
	currentTime := time.Now()
	currDate := currentTime.Format("2006-01-02")
	antal := GetCountPendingÖverföringar(db, currDate)
	if antal > 0 {
		fmt.Fprintf(w, "<p>%d fasta transaktioner tills idag väntar på att hanteras. Gå till <a href=\"fixedtrans\">Fasta transaktioner</a>.<p>\n", antal)
	}

	now := time.Now()
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()

	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)
	currDate = lastOfMonth.Format("2006-01-02")
	antal = GetCountPendingÖverföringar(db, currDate)
	if antal > 0 {
		fmt.Fprintf(w, "<p>%d fasta transaktioner till hela denna månaden väntar på att hanteras. Gå till <a href=\"fixedtrans\">Fasta transaktioner</a>.<p>\n", antal)
	}
}

func printAccounts(w http.ResponseWriter, db *sql.DB) {
	fmt.Fprintf(w, "<html>\n")
	fmt.Fprintf(w, "<head>\n")
	fmt.Fprintf(w, "<style>\n")
	fmt.Fprintf(w, "table,th,td { border: 1px solid black ; text-align: center }\n")
	fmt.Fprintf(w, "</style>\n")
	fmt.Fprintf(w, "</head>\n")
	fmt.Fprintf(w, "<body>\n")

	fmt.Fprintf(w, "<h1>Databasnamn: %s</h1>\n", currentDatabase)

	currentTime := time.Now()
	currDate := currentTime.Format("2006-01-02")

	fmt.Fprintf(w, "Dagens datum: %s<p>\n", currDate)

	res, err := db.Query("SELECT KontoNummer,Benämning,Saldo,StartSaldo,StartManad,Löpnr,SaldoArsskifte,ArsskifteManad FROM Konton")

	if err != nil {
		log.Fatal(err)
		os.Exit(2)
	}

	var KontoNummer []byte    // size 20
	var Benämning []byte      // size 40, index
	var Saldo []byte          // BCD / Decimal Precision 19
	var StartSaldo []byte     // BCD / Decimal Precision 19
	var StartManad []byte     // size 10
	var Löpnr []byte          // autoinc Primary Key
	var SaldoArsskifte []byte // BCD / Decimal Precision 19
	var ArsskifteManad []byte // size 10

	fmt.Fprintf(w, "<table style=\"width:100%%\"><tr><th>Kontonamn</th><th>Saldo enligt databas</th><th>Saldo uträknat för idag</th><th>Saldo uträknat totalt</th>\n")
	for res.Next() {
		err = res.Scan(&KontoNummer, &Benämning, &Saldo, &StartSaldo, &StartManad, &Löpnr, &SaldoArsskifte, &ArsskifteManad)

		acc := toUtf8(Benämning)
		dbSaldo , err2 := decimal.NewFromString(toUtf8(Saldo))
		if err2 != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(w, "<tr><td>%s</td>", acc)
		daySaldo, totSaldo := saldonKonto(acc, currDate)
		if dbSaldo.Equals(daySaldo) && dbSaldo.Equals(totSaldo) {
			fmt.Fprintf(w, "<td colspan=\"3\">%s</td>", dbSaldo)
		} else if dbSaldo.Equals(daySaldo) {
			fmt.Fprintf(w, "<td colspan=\"2\">%s</td>", dbSaldo)
			fmt.Fprintf(w, "<td>%s</td>", totSaldo)
		} else if daySaldo.Equals(totSaldo) {
			fmt.Fprintf(w, "<td>%s</td>", dbSaldo)
			fmt.Fprintf(w, "<td colspan=\"2\">%s</td>", totSaldo)
		} else {
			fmt.Fprintf(w, "<td>%s</td>", dbSaldo)
			fmt.Fprintf(w, "<td>%s</td>", daySaldo)
			fmt.Fprintf(w, "<td>%s</td>", totSaldo)
		}
		fmt.Fprintf(w, "</tr>\n")

	}
	fmt.Fprintf(w, "</table>\n")
}

func opendb(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	filename := sanitizeFilename(req.FormValue("fname"))

	if strings.HasSuffix(strings.ToLower(filename), ".mdb") {
		//fmt.Fprintf(w, "Trying to open Access/Jet<br>\n")
		db = openJetDB(filename, false)
	} else {
		if strings.HasSuffix(strings.ToLower(filename), ".db") {
			//fmt.Fprintf(w, "Trying to open sqlite3<br>\n")
			db = openSqlite(filename)
		} else {
			fmt.Fprintf(w, "Bad filename: %s<br>\n", filename)
		}
	}

	if db == nil {
		fmt.Fprintf(w, "<html>\n")
		fmt.Fprintf(w, "<body>\n")

		fmt.Fprintf(w, "Error opening database<p>\n")
	} else {
		fmt.Fprintf(w, "<!DOCTYPE html>\n")
		fmt.Fprintf(w, "<html>\n")
		fmt.Fprintf(w, "   <head>\n")
		fmt.Fprintf(w, "      <title>HTML Meta Tag</title>\n")
		fmt.Fprintf(w, "      <meta http-equiv = \"refresh\" content = \"0; url = /summary\" />\n")
		fmt.Fprintf(w, "   </head>\n")
		fmt.Fprintf(w, "   <body>\n")
		fmt.Fprintf(w, "      <p>Arbetar...</p>\n")
		fmt.Fprintf(w, "   </body>\n")
		fmt.Fprintf(w, "</html>\n")
	}
	fmt.Fprintf(w, "</body>\n")
	fmt.Fprintf(w, "</html>\n")
}

func closeDB() {
	db.Close()
	dbtype = 0
	db = nil
	currentDatabase = "NONE"
}

func closedb(w http.ResponseWriter, req *http.Request) {
	closeDB();
	
	fmt.Fprintf(w, "<!DOCTYPE html>\n")
	fmt.Fprintf(w, "<html>\n")
	fmt.Fprintf(w, "   <head>\n")
	fmt.Fprintf(w, "      <title>HTML Meta Tag</title>\n")
	fmt.Fprintf(w, "      <meta http-equiv = \"refresh\" content = \"10; url = /\" />\n")
	fmt.Fprintf(w, "   </head>\n")
	fmt.Fprintf(w, "   <body>\n")
	fmt.Fprintf(w, "      <p>Closing database!</p>\n")
	fmt.Fprintf(w, "   </body>\n")
	fmt.Fprintf(w, "</html>\n")
}

func quitapp(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "<!DOCTYPE html>\n")
	fmt.Fprintf(w, "<html>\n")
	fmt.Fprintf(w, "   <head>\n")
	fmt.Fprintf(w, "      <title>HTML Meta Tag</title>\n")
	fmt.Fprintf(w, "      <meta http-equiv = \"refresh\" content = \"10; url = /\" />\n")
	fmt.Fprintf(w, "   </head>\n")
	fmt.Fprintf(w, "   <body>\n")
	if db != nil {
		closedb(w, req)
	}
	fmt.Fprintf(w, "      <p>Avslutar. Hej då!</p>\n")
	fmt.Fprintf(w, "   </body>\n")
	fmt.Fprintf(w, "</html>\n")
	time.Sleep(8 * time.Second)
	srv.Shutdown(ctx);
}

func generateSummary(w http.ResponseWriter, req *http.Request) {
	printAccounts(w, db)
	checkÖverföringar(w, db)
	fmt.Fprintf(w, "<a href=\"monthly\">Månads kontoutdrag</a><p>\n")
	fmt.Fprintf(w, "<a href=\"transactions\">Transaktionslista</a><p>\n")
	fmt.Fprintf(w, "<a href=\"platser\">Platser</a><p>\n")
	fmt.Fprintf(w, "<a href=\"personer\">Personer</a><p>\n")
	fmt.Fprintf(w, "<a href=\"konton\">Konton</a><p>\n")
	fmt.Fprintf(w, "<a href=\"budget\">Budget</a><p>\n")
	fmt.Fprintf(w, "<a href=\"newtrans\">Ny transaktion</a><p>\n")
	fmt.Fprintf(w, "<a href=\"fixedtrans\">Fasta transaktioner/överföringar</a><p>\n")
	fmt.Fprintf(w, "<a href=\"close\">Stäng databas</a><p>\n")
	fmt.Fprintf(w, "<a href=\"quit\">Avsluta program</a><p>\n")
	fmt.Fprintf(w, "</body>\n")
	fmt.Fprintf(w, "</html>\n")
}

func printMonthly(w http.ResponseWriter, db *sql.DB, accName string, accYear int, accMonth int) {
	fmt.Fprintf(w, "<h1>%s</h1>\n", currentDatabase)
	fmt.Fprintf(w, "Kontonamn: %s<br>\n", accName)
	fmt.Fprintf(w, "År: %d<br>\n", accYear)
	fmt.Fprintf(w, "Månad: %d<br>\n", accMonth)

	var startDate, endDate string
	startDate = fmt.Sprintf("%d-%02d-01", accYear, accMonth)
	endDate = fmt.Sprintf("%d-%02d-01", accYear, accMonth+1)
	//fmt.Println("DEBUG Startdatum: ", startDate)
	//fmt.Println("DEBUG Slutdatum: ", endDate)

	fmt.Fprintf(w, "<p>\n")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var err error
	var res *sql.Rows
	var day_saldo [32]decimal.Decimal
	var day_found [32]bool

	res1 := db.QueryRowContext(ctx,
		`select startsaldo
  from konton
  where benämning = ?`, accName)
	var rawStart []byte // size 16
	err = res1.Scan(&rawStart)
	res2 := toUtf8(rawStart)
	startSaldo, err := decimal.NewFromString(res2)
	currSaldo := startSaldo

	res, err = db.QueryContext(ctx,
		`SELECT FrånKonto,TillKonto,Typ,Datum,Vad,Vem,Belopp,Löpnr,Saldo,Fastöverföring,Text from transaktioner
  where (datum < ?)
    and ((tillkonto = ?)
         or (frånkonto = ?))
order by datum,löpnr`, endDate, accName, accName)

	if err != nil {
		log.Fatal(err)
		os.Exit(2)
	}

	var fromAcc []byte // size 40
	var toAcc []byte   // size 40
	var tType []byte   // size 40
	var date []byte    // size 10
	var what []byte    // size 40
	var who []byte     // size 50
	var amount []byte  // BCD / Decimal Precision 19
	var nummer int     // Autoinc Primary Key, index
	var saldo []byte   // BCD / Decimal Precision 19
	var fixed bool     // Boolean
	var comment []byte // size 60

	fmt.Fprintf(w, "<table style=\"width:100%%\"><tr><th>Löpnr</th><th>Frånkonto</th><th>Tillkonto</th><th>Typ</th><th>Vad</th><th>Datum</th><th>Vem</th><th>Belopp</th><th>Saldo</th><th>Text</th><th>Fast överföring</th>\n")
	for res.Next() {
		err = res.Scan(&fromAcc, &toAcc, &tType, &date, &what, &who, &amount, &nummer, &saldo, &fixed, &comment)
		decAmount, _ := decimal.NewFromString(toUtf8(amount))
		if !day_found[01] && toUtf8(date) >= startDate {
			day_saldo[01] = currSaldo
			day_found[01] = true
		}
		if (accName == toUtf8(toAcc)) &&
			((toUtf8(tType) == "Uttag") ||
				(toUtf8(tType) == "Fast Inkomst") ||
				(toUtf8(tType) == "Insättning") ||
				(toUtf8(tType) == "Överföring")) {
			currSaldo = currSaldo.Add(decAmount)
		}
		if (accName == toUtf8(fromAcc)) &&
			((toUtf8(tType) == "Uttag") ||
				(toUtf8(tType) == "Inköp") ||
				(toUtf8(tType) == "Fast Utgift") ||
				(toUtf8(tType) == "Överföring")) {
			currSaldo = currSaldo.Sub(decAmount)
		}
		if toUtf8(date) >= startDate {
			sqlStmt := ""
			sqlStmt += "<tr><td>" + strconv.Itoa(nummer) + "</td>"
			sqlStmt += "<td>" + toUtf8(fromAcc) + "</td>"
			sqlStmt += "<td>" + toUtf8(toAcc) + "</td>"
			sqlStmt += "<td>" + toUtf8(tType) + "</td>"
			sqlStmt += "<td>" + toUtf8(what) + "</td>"
			sqlStmt += "<td>" + toUtf8(date) + "</td>"
			sqlStmt += "<td>" + toUtf8(who) + "</td>"
			sqlStmt += "<td>" + toUtf8(amount) + "</td>"
			sqlStmt += "<td>" + currSaldo.String() + "</td>"
			sqlStmt += "<td>" + html.EscapeString(toUtf8(comment)) + "</td>\n"
			sqlStmt += "<td>" + strconv.FormatBool(fixed) + "</td></tr>"
			fmt.Fprintf(w, "%s", sqlStmt)

			daynum, _ := strconv.Atoi(toUtf8(date)[8:10])
			day_saldo[daynum] = currSaldo
			day_found[daynum] = true
		}
	}
	fmt.Fprintf(w, "</table>\n")

	min_saldo := decimal.NewFromInt(math.MaxInt64)
	max_saldo := decimal.NewFromInt(math.MinInt64)

	for i := 1; i < 32; i++ {
		//fmt.Fprintf(w, "%d: %s<br>\n", i, day_saldo[i].String())
		if day_saldo[i].GreaterThan(max_saldo) {
			max_saldo = day_saldo[i]
		}
		if day_saldo[i].LessThan(min_saldo) {
			min_saldo = day_saldo[i]
		}
	}
	fmt.Fprintf(w, "<svg width=\"900\" height=\"600\">\n")
	var yf, val float64
	var y int
	var y1 decimal.Decimal
	const colWidth = 20
	for i := 1; i < 32; i++ {
		if day_found[i] {
			y1 = max_saldo.Sub(day_saldo[i])
			currSaldo = day_saldo[i]
		} else {
			y1 = max_saldo.Sub(currSaldo)
		}
		val, _ = y1.Float64()
		y2 := max_saldo.Sub(min_saldo)
		y2f, _ := y2.Float64()
		yf = val / (y2f / (500.0 - 0.0))
		y = int(yf)
		fmt.Fprintf(w, "  <rect x=%d y=%d width=\"%d\" height=\"%d\"", (i-1)*colWidth, int(y), colWidth, 500-int(y))
		fmt.Fprintf(w, "  style=\"fill:rgb(0,0,255);stroke-width:1;stroke:rgb(0,0,0)\" />\n")
	}
	// zero line
	y1 = max_saldo //.Sub(0.0)
	val, _ = y1.Float64()
	y2 := max_saldo.Sub(min_saldo)
	y2f, _ := y2.Float64()
	yf = val / (y2f / (500.0 - 0.0))
	y = int(yf)
	fmt.Fprintf(w, "  <rect x=%d y=%d width=\"%d\" height=\"%d\"", 0, y, 900, 1)
	fmt.Fprintf(w, "  style=\"fill:rgb(0,0,255);stroke-width:1;stroke:rgb(0,0,0)\" />\n")
	fmt.Fprintf(w, "<text fill=\"#000000\" font-size=\"12\" font-family=\"Verdana\" x=\"0\" y=\"550\">1</text>\n")
	fmt.Fprintf(w, "<text fill=\"#000000\" font-size=\"12\" font-family=\"Verdana\" x=\"%d\" y=\"550\">10</text>\n", (10-1)*colWidth)
	fmt.Fprintf(w, "<text fill=\"#000000\" font-size=\"12\" font-family=\"Verdana\" x=\"%d\" y=\"550\">20</text>\n", (20-1)*colWidth)
	fmt.Fprintf(w, "<text fill=\"#000000\" font-size=\"12\" font-family=\"Verdana\" x=\"%d\" y=\"550\">30</text>\n", (30-1)*colWidth)

	fmt.Fprintf(w, "<text fill=\"#000000\" font-size=\"12\" font-family=\"Verdana\" x=\"%d\" y=\"10\">%s</text>\n", 33*colWidth, max_saldo.String())
	fmt.Fprintf(w, "<text fill=\"#000000\" font-size=\"12\" font-family=\"Verdana\" x=\"%d\" y=\"500\">%s</text>\n", 33*colWidth, min_saldo.String())
	fmt.Fprintf(w, "Sorry, your browser does not support inline SVG.\n")
	fmt.Fprintf(w, "</svg>\n")
}

func monthly(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "<html>\n")
	fmt.Fprintf(w, "<head>\n")
	fmt.Fprintf(w, "<style>\n")
	fmt.Fprintf(w, "table,th,td { border: 1px solid black }\n")
	fmt.Fprintf(w, "</style>\n")
	fmt.Fprintf(w, "</head>\n")
	fmt.Fprintf(w, "<body>\n")

	err := req.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	var accYear, accMonth int
	var accName string

	if len(req.FormValue("accYear")) > 3 {
		accYear, err = strconv.Atoi(req.FormValue("accYear"))
		accMonth, err = strconv.Atoi(req.FormValue("accMonth"))
		accName = req.FormValue("accName")
	} else {
		res1 := db.QueryRow("SELECT MAX(Datum) FROM Transaktioner")
		var date []byte // size 10
		err = res1.Scan(&date)
		accYear, err = strconv.Atoi(toUtf8(date)[0:4])
		accMonth, err = strconv.Atoi(toUtf8(date)[5:7])
		res1 = db.QueryRow("SELECT TOP 1 Benämning FROM Konton")
		var namn []byte // size 10
		err = res1.Scan(&namn)
		accName = toUtf8(namn)
	}

	if db == nil {
		fmt.Fprintf(w, "Monthly: No database open<p>\n")
	} else {
		res1 := db.QueryRow("SELECT MIN(Datum) FROM Transaktioner")
		var date []byte // size 10
		err = res1.Scan(&date)
		firstYear, err := strconv.Atoi(toUtf8(date)[0:4])
		res1 = db.QueryRow("SELECT MAX(Datum) FROM Transaktioner")
		err = res1.Scan(&date)
		lastYear, err := strconv.Atoi(toUtf8(date)[0:4])

		printMonthly(w, db, accName, accYear, accMonth)
		//fmt.Fprintf(w, "<a href=\"monthly\">Månads kontoutdrag</a>\n")
		fmt.Fprintf(w, "<form method=\"POST\" action=\"/monthly\">\n")
		fmt.Fprintf(w, "<select id=\"accName\" name=\"accName\">\n")
		res, err := db.Query("SELECT KontoNummer,Benämning,Saldo,StartSaldo,StartManad,Löpnr,SaldoArsskifte,ArsskifteManad FROM Konton order by Benämning")

		if err != nil {
			log.Fatal(err)
			os.Exit(2)
		}

		var KontoNummer []byte    // size 20
		var Benämning []byte      // size 40, index
		var Saldo []byte          // BCD / Decimal Precision 19
		var StartSaldo []byte     // BCD / Decimal Precision 19
		var StartManad []byte     // size 10
		var Löpnr []byte          // autoinc Primary Key
		var SaldoArsskifte []byte // BCD / Decimal Precision 19
		var ArsskifteManad []byte // size 10
		for res.Next() {
			err = res.Scan(&KontoNummer, &Benämning, &Saldo, &StartSaldo, &StartManad, &Löpnr, &SaldoArsskifte, &ArsskifteManad)

			fmt.Fprintf(w, "<option value=\"%s\"", toUtf8(Benämning))
			if toUtf8(Benämning) == accName {
				fmt.Fprintf(w, " selected ")
			}
			fmt.Fprintf(w, ">%s</option>\n", toUtf8(Benämning))
		}
		fmt.Fprintf(w, "</select>\n")
		fmt.Fprintf(w, "<select id=\"accYear\" name=\"accYear\">\n")
		for year := firstYear; year <= lastYear; year++ {
			fmt.Fprintf(w, "<option value=\"%d\"", year)
			if year == accYear {
				fmt.Fprintf(w, " selected ")
			}
			fmt.Fprintf(w, ">%d</option>\n", year)
		}
		fmt.Fprintf(w, "</select>\n")
		fmt.Fprintf(w, "<select id=\"accMonth\" name=\"accMonth\">\n")
		for month := 1; month < 13; month++ {
			fmt.Fprintf(w, "<option value=\"%d\"", month)
			if month == accMonth {
				fmt.Fprintf(w, " selected ")
			}
			fmt.Fprintf(w, ">%d</option>\n", month)
		}
		fmt.Fprintf(w, "</select>\n")

		fmt.Fprintf(w, "<input type=\"submit\" value=\"Visa\"></form>\n")

		fmt.Fprintf(w, "<form method=\"POST\" action=\"/monthly\">\n")
		fmt.Fprintf(w, "<input type=\"hidden\" id=\"accName\" name=\"accName\" value=\"%s\">", accName)
		if accMonth+1 > 12 {
			fmt.Fprintf(w, "<input type=\"hidden\" id=\"accYear\" name=\"accYear\" value=\"%d\">", accYear+1)
			fmt.Fprintf(w, "<input type=\"hidden\" id=\"accMonth\" name=\"accMonth\" value=\"%d\">", 1)
		} else {
			fmt.Fprintf(w, "<input type=\"hidden\" id=\"accYear\" name=\"accYear\" value=\"%d\">", accYear)
			fmt.Fprintf(w, "<input type=\"hidden\" id=\"accMonth\" name=\"accMonth\" value=\"%d\">", accMonth+1)
		}
		fmt.Fprintf(w, "<input type=\"submit\" value=\"Nästa månad\"></form>\n")

		fmt.Fprintf(w, "<form method=\"POST\" action=\"/monthly\">\n")
		fmt.Fprintf(w, "<input type=\"hidden\" id=\"accName\" name=\"accName\" value=\"%s\">", accName)
		if accMonth-1 < 1 {
			fmt.Fprintf(w, "<input type=\"hidden\" id=\"accYear\" name=\"accYear\" value=\"%d\">", accYear-1)
			fmt.Fprintf(w, "<input type=\"hidden\" id=\"accMonth\" name=\"accMonth\" value=\"%d\">", 12)
		} else {
			fmt.Fprintf(w, "<input type=\"hidden\" id=\"accYear\" name=\"accYear\" value=\"%d\">", accYear)
			fmt.Fprintf(w, "<input type=\"hidden\" id=\"accMonth\" name=\"accMonth\" value=\"%d\">", accMonth-1)
		}
		fmt.Fprintf(w, "<input type=\"submit\" value=\"Föregående månad\"></form>\n")
	}

	fmt.Fprintf(w, "<a href=\"summary\">Översikt</a>\n")
	fmt.Fprintf(w, "</body>\n")
	fmt.Fprintf(w, "</html>\n")
}

func externalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("are you connected to the network?")
}

func getTypeInNames() []string {
	names := make([]string, 0)

	res, err := db.Query("SELECT Typ FROM Budget where Inkomst = 'J' ORDER BY Typ")

	if err != nil {
		log.Fatal(err)
		os.Exit(2)
	}

	var Typ []byte // size 40, index
	for res.Next() {
		err = res.Scan(&Typ)
		names = append(names, toUtf8(Typ))
	}
	return names
}

func getTypeOutNames() []string {
	names := make([]string, 0)

	res, err := db.Query("SELECT Typ FROM Budget where Inkomst = 'N' ORDER BY Typ")

	if err != nil {
		log.Fatal(err)
		os.Exit(2)
	}

	var Typ []byte // size 40, index
	for res.Next() {
		err = res.Scan(&Typ)
		names = append(names, toUtf8(Typ))
	}
	return names
}

func main() {
	http.HandleFunc("/hello", hello)
	http.HandleFunc("/r", restapi)
	http.HandleFunc("/headers", headers)
	http.HandleFunc("/open", opendb)
	http.HandleFunc("/close", closedb)
	http.HandleFunc("/quit", quitapp)
	http.HandleFunc("/newtrans", newtransaction)
	http.HandleFunc("/fixedtrans", fixedtransaction)
	//	http.HandleFunc("/addtrans", addtransaction)
	http.HandleFunc("/monthly", monthly)
	http.HandleFunc("/transactions", transactions)
	http.HandleFunc("/platser", hanteraplatser)
	http.HandleFunc("/personer", hanterapersoner)
	http.HandleFunc("/konton", hanterakonton)
	http.HandleFunc("/budget", hanteraBudget)
	http.HandleFunc("/summary", generateSummary)
	http.HandleFunc("/", root)

	ip, _ := externalIP()
	fmt.Println("Öppna URL i webläsaren:  http://localhost:8090/")
	fmt.Printf(" eller :  http://%s:8090/\n", ip)
	ctx = context.Background()
	srv := &http.Server{
		Addr:           ":8090",
		Handler:        nil,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	httpError := srv.ListenAndServe()
	if httpError != nil {
		log.Println("While serving HTTP: ", httpError)
	}
}
