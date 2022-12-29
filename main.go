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
// Hantera lösenordsskyddad databas

// ROADMAP/TODO/The Future Is In Flux
// ============
// BUG: Teckenkodning i lösenord
// localize decimal.String()
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
// Kräv inloggning (kräver https?)
// Årsskiftesrutin inkl uppdatera budget på olika sätt
// Graf som i månadsvyn, med valfri grupp av konton
// Graf som i månadsvyn fast för senaste året, med valfri grupp av konton
// Hantera kontokortsföretag
// Visa lån
// Lägg till nytt lån
// Redigera lån

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
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"html"
	"html/template"
	"io/ioutil"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3" // MIT License
	"github.com/shopspring/decimal" // MIT License
)

// Global variables
var db *sql.DB = nil
var nopwDb *sql.DB = nil
var dbtype uint8 = 0 // 0=none, 1=mdb/Access2.0, 2=sqlite3
var currentDatabase = "NONE"
var dbdecimaldot bool = false

func hello(w http.ResponseWriter, req *http.Request) {
	_, _ = fmt.Fprintf(w, "hello\n")
}

type RootPageData struct {
	FilerFinns bool
	AntalFiler string
	Filnamn    []string
}

//go:embed html/root.html
var htmlroot string

func root(w http.ResponseWriter, req *http.Request) {
	tmpl := template.New("wHHEK root")
	tmpl, _ = tmpl.Parse(htmlroot)

	files, err := ioutil.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}

	filer := make([]string, 0)

	if len(files) > 0 {
		for _, file := range files {
			if (JetDBSupport && strings.HasSuffix(strings.ToLower(file.Name()), ".mdb")) ||
				strings.HasSuffix(strings.ToLower(file.Name()), ".db") {
				filer = append(filer, file.Name())
				//log.Println("Hittad fil:", file.Name())

			}
		}
	}
	var antal = strconv.Itoa(len(filer))
	log.Println("Hittade filer", antal)
	data := RootPageData{
		FilerFinns: len(filer) > 0,
		AntalFiler: antal,
		Filnamn:    filer[:],
	}
	_ = tmpl.Execute(w, data)
}

//go:embed htmx.min.js
var htmxjs []byte

func htmx(w http.ResponseWriter, req *http.Request) {
	_, _ = w.Write(htmxjs)
}

//go:embed bars.svg
var barssvg []byte

func imgbars(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml")

	_, _ = w.Write(barssvg)
}

func restapi(w http.ResponseWriter, req *http.Request) {
	if req.URL.String() == "/r/main/accounts" {
		printSummaryTable(w, db)
	} else {
		_, _ = fmt.Fprintf(w, "Rest API not implemented yet.\n")
		_, _ = fmt.Fprintf(w, "Requested URL "+req.URL.String()+"\n")
	}
}

func headers(w http.ResponseWriter, req *http.Request) {
	_, _ = fmt.Fprintf(w, "Host: %s\n\n", req.Host)

	for name, headers := range req.Header {
		for _, h := range headers {
			_, _ = fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}

func GetCountPendingÖverföringar(db *sql.DB, currDate string) int {
	var cnt int
	_ = db.QueryRow(`select count(*) from Överföringar WHERE Datum <= ?`, currDate).Scan(&cnt)
	return cnt
}

//go:embed html/main4.html
var htmlmain4 string
type Main4Data struct {
	Antal int
}

//go:embed html/main5.html
var htmlmain5 string
type Main5Data struct {
	Antal int
}

func checkÖverföringar(w http.ResponseWriter, db *sql.DB) {
	currentTime := time.Now()
	currDate := currentTime.Format("2006-01-02")
	antal := GetCountPendingÖverföringar(db, currDate)
	if antal > 0 {
		t := template.New("Main4")
		t, _ = t.Parse(htmlmain4)
		data := Main4Data{
			Antal: antal,
		}
		err := t.Execute(w, data)
		if err != nil {
			log.Println("While serving HTTP main4: ", err)
		}
	}

	now := time.Now()
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()

	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)
	currDate = lastOfMonth.Format("2006-01-02")
	antal = GetCountPendingÖverföringar(db, currDate)
	if antal > 0 {
		t := template.New("Main5")
		t, _ = t.Parse(htmlmain5)
		data := Main5Data{
			Antal: antal,
		}
		err := t.Execute(w, data)
		if err != nil {
			log.Println("While serving HTTP main5: ", err)
		}
	}
}

//go:embed html/main1.html
var htmlmain1 string

type Main1Data struct {
	CurrDBName string
	CurrDate string
}

func printSummaryHead(w http.ResponseWriter) {
	currentTime := time.Now()
	currDate := currentTime.Format("2006-01-02")

	t := template.New("Lösenordshantering")
	t, _ = t.Parse(htmlmain1)
	data := Main1Data{
		CurrDBName: currentDatabase,
		CurrDate: currDate,
	}
	_ = t.Execute(w, data)
}

//go:embed html/main6.html
var htmlmain6 string

func printAccounts(w http.ResponseWriter) {
	t := template.New("Main6")
	t, _ = t.Parse(htmlmain6)
	err := t.Execute(w, t)
	if err != nil {
		log.Println("While serving HTTP main6: ", err)
	}
}

type sumType struct {
	Name     string
	DbSaldo  string
	DaySaldo string
	TotSaldo string
}

//go:embed html/main11.html
var htmlmain11 string
type Main11Data struct {
	Konton []sumType
}

func printSummaryTable(w http.ResponseWriter, db *sql.DB) {
	currentTime := time.Now()
	currDate := currentTime.Format("2006-01-02")

	res, err := db.Query("SELECT KontoNummer,Benämning,Saldo,StartSaldo,StartManad,Löpnr,SaldoArsskifte,ArsskifteManad FROM Konton ORDER BY Benämning")

	if err != nil {
		log.Fatal("printSummaryTable:", err)
	}

	var KontoNummer []byte    // size 20
	var Benämning []byte      // size 40, index
	var Saldo []byte          // BCD / Decimal Precision 19
	var StartSaldo []byte     // BCD / Decimal Precision 19
	var StartManad []byte     // size 10
	var Löpnr []byte          // autoinc Primary Key
	var SaldoArsskifte []byte // BCD / Decimal Precision 19
	var ArsskifteManad []byte // size 10

	var konton []sumType

	for res.Next() {
		err = res.Scan(&KontoNummer, &Benämning, &Saldo, &StartSaldo, &StartManad, &Löpnr, &SaldoArsskifte, &ArsskifteManad)
		
		acc := toUtf8(Benämning)
		DbSaldo, err2 := decimal.NewFromString(strings.ReplaceAll(toUtf8(Saldo), ",", "."))
		if err2 != nil {
			log.Fatal("printSummaryTable:", err)
		}
		DaySaldo, TotSaldo := saldonKonto(db, acc, currDate)
		
		konton = append(konton, sumType{acc, AmountDec2DBStr(DbSaldo), AmountDec2DBStr(DaySaldo), AmountDec2DBStr(TotSaldo)})
	}
	t := template.New("Main11")
	t, _ = t.Parse(htmlmain11)
	data := Main11Data{
		Konton: konton,
	}
	err = t.Execute(w, data)
	if err != nil {
		log.Println("While serving HTTP main11: ", err)
	}
}

//go:embed html/main12.html
var htmlmain12 string
//go:embed html/main13.html
var htmlmain13 string

func checkpwd(w http.ResponseWriter, req *http.Request) {
	if nopwDb == nil {
		t := template.New("Main12")
		t, _ = t.Parse(htmlmain12)
		err := t.Execute(w, nil)
		if err != nil {
			log.Println("While serving HTTP main12: ", err)
		}
	} else {
		pwd := getdbpw(nopwDb)
		formpwd := req.FormValue("pwd")
		//log.Println("jmf %s %s", pwd, formpwd)
		if pwd == formpwd {
			//fmt.Fprintf(w, "OK.")
			db = nopwDb
			nopwDb = nil
			showsummary(w)
		} else {
			t := template.New("Main13")
			t, _ = t.Parse(htmlmain13)
			err := t.Execute(w, nil)
			if err != nil {
				log.Println("While serving HTTP main13: ", err)
			}
		}
	}
}

//go:embed html/main3.html
var htmlmain3 string

func showsummary(w http.ResponseWriter) {
	t := template.New("Main3")
	t, _ = t.Parse(htmlmain3)
	err := t.Execute(w, t)
	if err != nil {
		log.Println("While serving HTTP main3: ", err)
	}
}

func opendb(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	filename := sanitizeFilename(req.FormValue("fname"))

	if strings.HasSuffix(strings.ToLower(filename), ".mdb") {
		//fmt.Fprintf(w, "Trying to open Access/Jet<br>\n")
		nopwDb = openJetDB(filename, false)
	} else {
		if strings.HasSuffix(strings.ToLower(filename), ".db") {
			//fmt.Fprintf(w, "Trying to open sqlite3<br>\n")
			nopwDb = openSqlite(filename)
		} else {
			_, _ = fmt.Fprintf(w, "Bad filename: %s<br>\n", filename)
			_, _ = fmt.Fprintf(w, "</body>\n")
			_, _ = fmt.Fprintf(w, "</html>\n")
		}
	}

	if nopwDb == nil {
		_, _ = fmt.Fprintf(w, "<html>\n")
		_, _ = fmt.Fprintf(w, "<body>\n")

		_, _ = fmt.Fprintf(w, "Error opening database<p>\n")
		_, _ = fmt.Fprintf(w, "</body>\n")
		_, _ = fmt.Fprintf(w, "</html>\n")
	} else {
		pwd := getdbpw(nopwDb)
		if pwd != " " {
			_, _ = fmt.Fprintf(w, "<!DOCTYPE html>\n")
			_, _ = fmt.Fprintf(w, "<html>\n")
			_, _ = fmt.Fprintf(w, "   <head>\n")
			_, _ = fmt.Fprintf(w, "      <title>Lösenordsskyddad</title>\n")
			_, _ = fmt.Fprintf(w, "   </head>\n")
			_, _ = fmt.Fprintf(w, "   <body>\n")
			_, _ = fmt.Fprintf(w, "      <p>Databasen är lösenordsskyddad. Skriv in lösenordet:</p>\n")
			_, _ = fmt.Fprintf(w, "<form method=\"POST\" action=\"/pwd\">\n")
			_, _ = fmt.Fprintf(w, "      <label for=\"pwd\">Password:</label><input type=\"password\" id=\"pwd\" name=\"pwd\">\n")
			_, _ = fmt.Fprintf(w, "<input type=\"submit\" value=\"Använd\"></form>\n")

			_, _ = fmt.Fprintf(w, "</body>\n")
			_, _ = fmt.Fprintf(w, "</html>\n")
		} else {
			db = nopwDb
			nopwDb = nil
			showsummary(w)
		}
	}
}

func closeDB() {
	_ = db.Close()
	dbtype = 0
	db = nil
	currentDatabase = "NONE"
}

//go:embed html/main7.html
var htmlmain7 string

func closedb(w http.ResponseWriter, req *http.Request) {
	closeDB()

	t := template.New("Main3")
	t, _ = t.Parse(htmlmain3)
	err := t.Execute(w, t)
	if err != nil {
		log.Println("While serving HTTP main3: ", err)
	}
}

//go:embed html/main8.html
var htmlmain8 string

func quitapp(w http.ResponseWriter, req *http.Request) {
	t := template.New("Main8")
	t, _ = t.Parse(htmlmain8)
	err := t.Execute(w, t)
	if err != nil {
		log.Println("While serving HTTP main8: ", err)
	}

	if db != nil {
		closedb(w, req)
	}
	
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
	//time.Sleep(8 * time.Second)
	//srv.Shutdown(ctx);
	os.Exit(0)
}

//go:embed html/main9.html
var htmlmain9 string
type Main9Data struct {
	Filnamn string
}

//go:embed html/main10.html
var htmlmain10 string
type Main10Data struct {
	Filnamn string
}

func createdb(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	filename := req.FormValue("fname")

	if (len([]rune(filename)) < 1) ||
		(strings.ContainsAny(filename, "\\/|:<>.\"'`\x00")) {

		t := template.New("Main9")
		t, _ = t.Parse(htmlmain9)
		data := Main9Data{
			Filnamn: filename,
		}
		err := t.Execute(w, data)
		if err != nil {
			log.Println("While serving HTTP main9: ", err)
		}
		return
	}

	SkapaTomDB(filename + ".db")
	t := template.New("Main10")
	t, _ = t.Parse(htmlmain10)
	data := Main10Data{
		Filnamn: filename,
	}
	err = t.Execute(w, data)
	if err != nil {
		log.Println("While serving HTTP main10: ", err)
	}
}

//go:embed html/main2.html
var htmlmain2 string

func generateSummary(w http.ResponseWriter, req *http.Request) {
	printSummaryHead(w)
	if db != nil {
		checkÖverföringar(w, db)
		t := template.New("Main2")
		t, _ = t.Parse(htmlmain2)
		err := t.Execute(w, t)
		if err != nil {
			log.Println("While serving HTTP main2: ", err)
		}

		printAccounts(w)
		_, _ = fmt.Fprintf(w, "</td></tr></table>\n")
	} else {
		_, _ = fmt.Fprintf(w, "Ingen databas.\n")
	}

	_, _ = fmt.Fprintf(w, "</body>\n")
	_, _ = fmt.Fprintf(w, "</html>\n")
}

func printMonthly(w http.ResponseWriter, db *sql.DB, accName string, accYear int, accMonth int) {
	_, _ = fmt.Fprintf(w, "<h1>%s</h1>\n", currentDatabase)
	_, _ = fmt.Fprintf(w, "Kontonamn: %s<br>\n", accName)
	_, _ = fmt.Fprintf(w, "År: %d<br>\n", accYear)
	_, _ = fmt.Fprintf(w, "Månad: %d<br>\n", accMonth)

	var startDate, endDate string
	startDate = fmt.Sprintf("%d-%02d-01", accYear, accMonth)
	endDate = fmt.Sprintf("%d-%02d-01", accYear, accMonth+1)
	//fmt.Println("DEBUG Startdatum: ", startDate)
	//fmt.Println("DEBUG Slutdatum: ", endDate)

	_, _ = fmt.Fprintf(w, "<p>\n")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var err error
	var res *sql.Rows
	var daySaldo [32]decimal.Decimal
	var dayFound [32]bool

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

	_, _ = fmt.Fprintf(w, "<table style=\"width:100%%\"><tr><th>Löpnr</th><th>Frånkonto</th><th>Tillkonto</th><th>Typ</th><th>Vad</th><th>Datum</th><th>Vem</th><th>Belopp</th><th>Saldo</th><th>Text</th><th>Fast överföring</th>\n")
	for res.Next() {
		err = res.Scan(&fromAcc, &toAcc, &tType, &date, &what, &who, &amount, &nummer, &saldo, &fixed, &comment)
		decAmount, _ := decimal.NewFromString(toUtf8(amount))
		if !dayFound[01] && toUtf8(date) >= startDate {
			daySaldo[01] = currSaldo
			dayFound[01] = true
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
			_, _ = fmt.Fprintf(w, "%s", sqlStmt)

			daynum, _ := strconv.Atoi(toUtf8(date)[8:10])
			daySaldo[daynum] = currSaldo
			dayFound[daynum] = true
		}
	}
	_, _ = fmt.Fprintf(w, "</table>\n")

	minSaldo := decimal.NewFromInt(math.MaxInt64)
	maxSaldo := decimal.NewFromInt(math.MinInt64)

	for i := 1; i < 32; i++ {
		//fmt.Fprintf(w, "%d: %s<br>\n", i, day_saldo[i].String())
		if daySaldo[i].GreaterThan(maxSaldo) {
			maxSaldo = daySaldo[i]
		}
		if daySaldo[i].LessThan(minSaldo) {
			minSaldo = daySaldo[i]
		}
	}
	_, _ = fmt.Fprintf(w, "<svg width=\"900\" height=\"600\">\n")
	var yf, val float64
	var y int
	var y1 decimal.Decimal
	const colWidth = 20
	for i := 1; i < 32; i++ {
		if dayFound[i] {
			y1 = maxSaldo.Sub(daySaldo[i])
			currSaldo = daySaldo[i]
		} else {
			y1 = maxSaldo.Sub(currSaldo)
		}
		val, _ = y1.Float64()
		y2 := maxSaldo.Sub(minSaldo)
		y2f, _ := y2.Float64()
		yf = val / (y2f / (500.0 - 0.0))
		y = int(yf)
		_, _ = fmt.Fprintf(w, "  <rect x=%d y=%d width=\"%d\" height=\"%d\"", (i-1)*colWidth, y, colWidth, 500-int(y))
		_, _ = fmt.Fprintf(w, "  style=\"fill:rgb(0,0,255);stroke-width:1;stroke:rgb(0,0,0)\" />\n")
	}
	// zero line
	y1 = maxSaldo //.Sub(0.0)
	val, _ = y1.Float64()
	y2 := maxSaldo.Sub(minSaldo)
	y2f, _ := y2.Float64()
	yf = val / (y2f / (500.0 - 0.0))
	y = int(yf)
	_, _ = fmt.Fprintf(w, "  <rect x=%d y=%d width=\"%d\" height=\"%d\"", 0, y, 900, 1)
	_, _ = fmt.Fprintf(w, "  style=\"fill:rgb(0,0,255);stroke-width:1;stroke:rgb(0,0,0)\" />\n")
	_, _ = fmt.Fprintf(w, "<text fill=\"#000000\" font-size=\"12\" font-family=\"Verdana\" x=\"0\" y=\"550\">1</text>\n")
	_, _ = fmt.Fprintf(w, "<text fill=\"#000000\" font-size=\"12\" font-family=\"Verdana\" x=\"%d\" y=\"550\">10</text>\n", (10-1)*colWidth)
	_, _ = fmt.Fprintf(w, "<text fill=\"#000000\" font-size=\"12\" font-family=\"Verdana\" x=\"%d\" y=\"550\">20</text>\n", (20-1)*colWidth)
	_, _ = fmt.Fprintf(w, "<text fill=\"#000000\" font-size=\"12\" font-family=\"Verdana\" x=\"%d\" y=\"550\">30</text>\n", (30-1)*colWidth)

	_, _ = fmt.Fprintf(w, "<text fill=\"#000000\" font-size=\"12\" font-family=\"Verdana\" x=\"%d\" y=\"10\">%s</text>\n", 33*colWidth, maxSaldo.String())
	_, _ = fmt.Fprintf(w, "<text fill=\"#000000\" font-size=\"12\" font-family=\"Verdana\" x=\"%d\" y=\"500\">%s</text>\n", 33*colWidth, minSaldo.String())
	_, _ = fmt.Fprintf(w, "Sorry, your browser does not support inline SVG.\n")
	_, _ = fmt.Fprintf(w, "</svg>\n")
}

func monthly(w http.ResponseWriter, req *http.Request) {
	_, _ = fmt.Fprintf(w, "<html>\n")
	_, _ = fmt.Fprintf(w, "<head>\n")
	_, _ = fmt.Fprintf(w, "<style>\n")
	_, _ = fmt.Fprintf(w, "table,th,td { border: 1px solid black }\n")
	_, _ = fmt.Fprintf(w, "</style>\n")
	_, _ = fmt.Fprintf(w, "</head>\n")
	_, _ = fmt.Fprintf(w, "<body>\n")

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
		_, _ = fmt.Fprintf(w, "Monthly: No database open<p>\n")
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
		_, _ = fmt.Fprintf(w, "<form method=\"POST\" action=\"/monthly\">\n")
		_, _ = fmt.Fprintf(w, "<select id=\"accName\" name=\"accName\">\n")
		res, err := db.Query("SELECT KontoNummer,Benämning,Saldo,StartSaldo,StartManad,Löpnr,SaldoArsskifte,ArsskifteManad FROM Konton order by Benämning")

		if err != nil {
			log.Fatal(err)
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

			_, _ = fmt.Fprintf(w, "<option value=\"%s\"", toUtf8(Benämning))
			if toUtf8(Benämning) == accName {
				_, _ = fmt.Fprintf(w, " selected ")
			}
			_, _ = fmt.Fprintf(w, ">%s</option>\n", toUtf8(Benämning))
		}
		_, _ = fmt.Fprintf(w, "</select>\n")
		_, _ = fmt.Fprintf(w, "<select id=\"accYear\" name=\"accYear\">\n")
		for year := firstYear; year <= lastYear; year++ {
			_, _ = fmt.Fprintf(w, "<option value=\"%d\"", year)
			if year == accYear {
				_, _ = fmt.Fprintf(w, " selected ")
			}
			_, _ = fmt.Fprintf(w, ">%d</option>\n", year)
		}
		_, _ = fmt.Fprintf(w, "</select>\n")
		_, _ = fmt.Fprintf(w, "<select id=\"accMonth\" name=\"accMonth\">\n")
		for month := 1; month < 13; month++ {
			_, _ = fmt.Fprintf(w, "<option value=\"%d\"", month)
			if month == accMonth {
				_, _ = fmt.Fprintf(w, " selected ")
			}
			_, _ = fmt.Fprintf(w, ">%d</option>\n", month)
		}
		_, _ = fmt.Fprintf(w, "</select>\n")

		_, _ = fmt.Fprintf(w, "<input type=\"submit\" value=\"Visa\"></form>\n")

		_, _ = fmt.Fprintf(w, "<form method=\"POST\" action=\"/monthly\">\n")
		_, _ = fmt.Fprintf(w, "<input type=\"hidden\" id=\"accName\" name=\"accName\" value=\"%s\">", accName)
		if accMonth+1 > 12 {
			_, _ = fmt.Fprintf(w, "<input type=\"hidden\" id=\"accYear\" name=\"accYear\" value=\"%d\">", accYear+1)
			_, _ = fmt.Fprintf(w, "<input type=\"hidden\" id=\"accMonth\" name=\"accMonth\" value=\"%d\">", 1)
		} else {
			_, _ = fmt.Fprintf(w, "<input type=\"hidden\" id=\"accYear\" name=\"accYear\" value=\"%d\">", accYear)
			_, _ = fmt.Fprintf(w, "<input type=\"hidden\" id=\"accMonth\" name=\"accMonth\" value=\"%d\">", accMonth+1)
		}
		_, _ = fmt.Fprintf(w, "<input type=\"submit\" value=\"Nästa månad\"></form>\n")

		_, _ = fmt.Fprintf(w, "<form method=\"POST\" action=\"/monthly\">\n")
		_, _ = fmt.Fprintf(w, "<input type=\"hidden\" id=\"accName\" name=\"accName\" value=\"%s\">", accName)
		if accMonth-1 < 1 {
			_, _ = fmt.Fprintf(w, "<input type=\"hidden\" id=\"accYear\" name=\"accYear\" value=\"%d\">", accYear-1)
			_, _ = fmt.Fprintf(w, "<input type=\"hidden\" id=\"accMonth\" name=\"accMonth\" value=\"%d\">", 12)
		} else {
			_, _ = fmt.Fprintf(w, "<input type=\"hidden\" id=\"accYear\" name=\"accYear\" value=\"%d\">", accYear)
			_, _ = fmt.Fprintf(w, "<input type=\"hidden\" id=\"accMonth\" name=\"accMonth\" value=\"%d\">", accMonth-1)
		}
		_, _ = fmt.Fprintf(w, "<input type=\"submit\" value=\"Föregående månad\"></form>\n")
	}

	_, _ = fmt.Fprintf(w, "<a href=\"summary\">Översikt</a>\n")
	_, _ = fmt.Fprintf(w, "</body>\n")
	_, _ = fmt.Fprintf(w, "</html>\n")
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
	return "", errors.New("are you connected to the network")
}

func getTypeInNames() []string {
	names := make([]string, 0)

	res, err := db.Query("SELECT Typ FROM Budget where Inkomst = 'J' ORDER BY Typ")

	if err != nil {
		log.Fatal(err)
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
	}

	var Typ []byte // size 40, index
	for res.Next() {
		err = res.Scan(&Typ)
		names = append(names, toUtf8(Typ))
	}
	return names
}

//go:embed html/help1.html
var html1 string

func help1(w http.ResponseWriter, req *http.Request) {
	t := template.New("Hjälp example")
	t, _ = t.Parse(html1)
	err := t.Execute(w, t)
	if err != nil {
		return
	}
}

func main() {
	helpPtr := flag.Bool("help", false, "Skriv ut hjälptext.")
	
	flag.Parse()
	
	if *helpPtr {
		flag.Usage()
		os.Exit(1)
	}

	dbdecimaldot = !detectdbdec()
	log.Println("Setting dbdecimaldot: ", dbdecimaldot)

	http.HandleFunc("/hello", hello)
	http.HandleFunc("/r/", restapi)
	http.HandleFunc("/htmx.js", htmx)
	http.HandleFunc("/img/bars.svg", imgbars)
	http.HandleFunc("/headers", headers)
	http.HandleFunc("/open", opendb)
	http.HandleFunc("/createdb", createdb)
	http.HandleFunc("/pwd", checkpwd)
	http.HandleFunc("/close", closedb)
	http.HandleFunc("/quit", quitapp)
	http.HandleFunc("/newtrans", newtransaction)
	http.HandleFunc("/fixedtrans", fixedtransactionHTML)
	//	http.HandleFunc("/addtrans", addtransaction)
	http.HandleFunc("/monthly", monthly)
	http.HandleFunc("/transactions", transactions)
	http.HandleFunc("/platser", hanteraplatser)
	http.HandleFunc("/personer", hanterapersoner)
	http.HandleFunc("/konton", hanterakonton)
	http.HandleFunc("/budget", hanteraBudget)
	http.HandleFunc("/yresult", hanteraYResult)
	http.HandleFunc("/summary", generateSummary)
	http.HandleFunc("/acccmp", compareaccount)
	http.HandleFunc("/passwd", passwordmgmt)
	http.HandleFunc("/help1", help1)
	http.HandleFunc("/", root)

	//ip, _ := externalIP()
	fmt.Println("Öppna URL i webläsaren:  http://localhost:8090/")
	//fmt.Printf(" eller :  http://%s:8090/\n", ip)
	_ = context.Background()
	srv := &http.Server{
		Addr:           "127.0.0.1:8090",
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
