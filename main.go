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
// kommandorads help-option
// Efter lagt till transaktion, visa den tillagda
// Visa transaktioner, filter: Frånkonto
// Visa transaktioner, filter: Tillkonto
// Visa transaktioner, filter: Summa
// Visa transaktioner, filter: Plats
// Visa transaktioner, filter: Vad
// Visa resultat-tabellen, helår
// Skapa ny fil (sqlite), kompatibel

// ROADMAP/TODO/The Future Is In Flux
// ============
// BUG: Teckenkodning i lösenord
// localize decimal.String()
// escape & in all html. escapeHTML verkar inte fungera. Use template?
// hantera fel: för lång text till comment
// kommandoradsoption för att välja katalog med databas
// kommandoradsoption för att välja databas
// kommandoradsoption för att sätta portnummer
// kommandoradsoption för att lägga till transaktion
// startscript: starta med rätt argument samt starta webläsare
// Installationsinstruktion (ladda ner exe, skapa ikon, brandvägg)
// Visa transaktioner, filter: Person
// Visa resultat-tabellen, aktuell/vald månad
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
	"github.com/pkg/browser" // BSD-2-Clause
)

// Global variables
var db *sql.DB = nil
var nopwDb *sql.DB = nil
var dbtype uint8 = 0 // 0=none, 1=mdb/Access2.0, 2=sqlite3
var currentDatabase = "NONE"
var dbdecimaldot bool = false

func hello(w http.ResponseWriter, req *http.Request) {
	log.Println("Func hello")
	
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
	log.Println("Func root")

	if db != nil {
		generateSummary(w, req)
		return
	}
	
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
	log.Println("Func restapi")
	
	if req.URL.String() == "/r/main/accounts" {
		printSummaryTable(w, db)
	} else {
		_, _ = fmt.Fprintf(w, "Rest API not implemented yet.\n")
		_, _ = fmt.Fprintf(w, "Requested URL "+req.URL.String()+"\n")
	}
}

func headers(w http.ResponseWriter, req *http.Request) {
	log.Println("Func headers")
	
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
		
		konton = append(konton, sumType{acc, Dec2Str(DbSaldo), Dec2Str(DaySaldo), Dec2Str(TotSaldo)})
	}
	res.Close()
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
	log.Println("Func checkpwd")
	
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

//go:embed html/main14.html
var htmlmain14 string
type Main14Data struct {
	Filnamn string
}
//go:embed html/main15.html
var htmlmain15 string
//go:embed html/main16.html
var htmlmain16 string

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
			t := template.New("Main14")
			t, _ = t.Parse(htmlmain14)
			data := Main14Data{
				Filnamn: filename,
			}
			err := t.Execute(w, data)
			if err != nil {
				log.Println("While serving HTTP main14: ", err)
			}
		}
	}
	
	if nopwDb == nil {
		t := template.New("Main15")
		t, _ = t.Parse(htmlmain15)
		err := t.Execute(w, nil)
		if err != nil {
			log.Println("While serving HTTP main15: ", err)
		}
	} else {
		pwd := getdbpw(nopwDb)
		if pwd != " " {
			t := template.New("Main16")
			t, _ = t.Parse(htmlmain16)
			err := t.Execute(w, nil)
			if err != nil {
				log.Println("While serving HTTP main16: ", err)
			}
		} else {
			db = nopwDb
			nopwDb = nil
			showsummary(w)
		}
	}
}

func closeDB() {
	if db != nil {
		_ = db.Close()
		dbtype = 0
		db = nil
		currentDatabase = "NONE"
	} else {
		log.Println("closeDB() trying to close nil")
	}
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
	log.Println("Func generateSummary")
	
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
		_, _ = fmt.Fprintf(w, "<a href=\"/\">Tillbaka</a>.\n")
	}

	_, _ = fmt.Fprintf(w, "</body>\n")
	_, _ = fmt.Fprintf(w, "</html>\n")
}

type TransactionType struct {
	Löpnr   string
	AccName string
	Dest    string
	Typ     string
	Vad     string
	Datum   string
	Vem     string
	Belopp  string
	Saldo   string
	Text    string
	Fixed   string
}

type MonthValueType struct {
	X      string
	Y      string
	Width  string
	Height string
}

type TextType struct {
	X    string
	Y    string
	Text string
}

//go:embed html/main20.html
var htmlmain20 string
type Main20Data struct {
	Filename string
	AccName string
	Year string
	Month string
	Transactions []TransactionType
	ZeroLine string
	MonthValues []MonthValueType
	Texts []TextType
}

func printMonthly(w http.ResponseWriter, db *sql.DB, accName string, accYear int, accMonth int) {
	var transactions []TransactionType
	var monthValues []MonthValueType
	var texts []TextType

	var startDate, endDate string
	startDate = fmt.Sprintf("%d-%02d-01", accYear, accMonth)
	endDate = fmt.Sprintf("%d-%02d-01", accYear, accMonth+1)
	//fmt.Println("DEBUG Startdatum: ", startDate)
	//fmt.Println("DEBUG Slutdatum: ", endDate)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var err error
	var res *sql.Rows
	var daySaldo [32]decimal.Decimal
	var dayFound [32]bool
	var rawStart []byte // size 16

	err = db.QueryRowContext(ctx,
		`select startsaldo
  from konton
  where benämning = ?`, accName).Scan(&rawStart)
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
			var transaction TransactionType
			transaction.Löpnr = strconv.Itoa(nummer)
			transaction.AccName = toUtf8(fromAcc)
			transaction.Dest = toUtf8(toAcc)
			transaction.Typ = toUtf8(tType)
			transaction.Vad = toUtf8(what)
			transaction.Datum = toUtf8(date)
			transaction.Vem = toUtf8(who)
			
			str := toUtf8(amount)
			dec, _ := decimal.NewFromString(str)
			transaction.Belopp = Dec2Str(dec)
			transaction.Saldo = Dec2Str(currSaldo)
			transaction.Text = toUtf8(comment)
			transaction.Fixed = strconv.FormatBool(fixed)
			transactions = append(transactions, transaction)

			daynum, _ := strconv.Atoi(toUtf8(date)[8:10])
			daySaldo[daynum] = currSaldo
			dayFound[daynum] = true
		}
	}
	res.Close()

	minSaldo := decimal.NewFromInt(math.MaxInt64)
	maxSaldo := decimal.NewFromInt(math.MinInt64)

	for i := 1; i < 32; i++ {
		if daySaldo[i].GreaterThan(maxSaldo) {
			maxSaldo = daySaldo[i]
		}
		if daySaldo[i].LessThan(minSaldo) {
			minSaldo = daySaldo[i]
		}
	}

	var yf, val float64
	var y int
	var y1 decimal.Decimal
	const colWidth = 20
	var monthValue MonthValueType

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
		
		monthValue.X = strconv.Itoa((i-1)*colWidth)
		monthValue.Y = strconv.Itoa(y)
		monthValue.Width = strconv.Itoa(colWidth)
		monthValue.Height = strconv.Itoa(500-int(y))
		monthValues = append(monthValues, monthValue)
	}
	// zero line
	y1 = maxSaldo //.Sub(0.0)
	val, _ = y1.Float64()
	y2 := maxSaldo.Sub(minSaldo)
	y2f, _ := y2.Float64()
	yf = val / (y2f / (500.0 - 0.0))
	y = int(yf)
	zeroLine := y

	var text TextType
	text.X = "0"
	text.Y = "550"
	text.Text = "1"
	texts = append(texts, text)
	text.X = strconv.Itoa((10-1)*colWidth)
	text.Y = "550"
	text.Text = "10"
	texts = append(texts, text)
	text.X = strconv.Itoa((20-1)*colWidth)
	text.Y = "550"
	text.Text = "20"
	texts = append(texts, text)
	text.X = strconv.Itoa((30-1)*colWidth)
	text.Y = "550"
	text.Text = "30"
	texts = append(texts, text)

	text.X = strconv.Itoa(33*colWidth)
	text.Y = "10"
	text.Text = maxSaldo.String()
	texts = append(texts, text)
	text.X = strconv.Itoa(33*colWidth)
	text.Y = "500"
	text.Text = minSaldo.String()
	texts = append(texts, text)

	t := template.New("Main20")
	t, _ = t.Parse(htmlmain20)
	data := Main20Data{
		Filename: currentDatabase,
		AccName: accName,
		Year: strconv.Itoa(accYear),
		Month: strconv.Itoa(accMonth),
		Transactions: transactions,
		ZeroLine: strconv.Itoa(zeroLine),
		MonthValues: monthValues,
		Texts: texts,
	}
	err = t.Execute(w, data)
	if err != nil {
		log.Println("While serving HTTP main11: ", err)
	}

}

//go:embed html/main17.html
var htmlmain17 string
type Main17Data struct {
	NoDB bool
}

type kontoType struct {
	Name     string
	Selected bool
}
type monthType struct {
	Name     string
	Selected bool
}
type yearType struct {
	Name     string
	Selected bool
}
//go:embed html/main18.html
var htmlmain18 string
type Main18Data struct {
	Konton []kontoType
	Years []yearType
	Months []monthType
	SelectKonto string
	NextYear string
	NextMonth string
	PrevYear string
	PrevMonth string
}
//go:embed html/main19.html
var htmlmain19 string

func monthly(w http.ResponseWriter, req *http.Request) {
	t := template.New("Main17")
	t, _ = t.Parse(htmlmain17)
	data := Main17Data{
		NoDB: db == nil,
	}
	err := t.Execute(w, data)
	if err != nil {
		log.Println("While serving HTTP main17: ", err)
	}

	err = req.ParseForm()
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
		var date []byte // size 10
		err = db.QueryRow("SELECT MAX(Datum) FROM Transaktioner").Scan(&date)
		accYear, err = strconv.Atoi(toUtf8(date)[0:4])
		accMonth, err = strconv.Atoi(toUtf8(date)[5:7])
		var namn []byte // size 10
		err = db.QueryRow("SELECT TOP 1 Benämning FROM Konton").Scan(&namn)
		accName = toUtf8(namn)
	}

	if db != nil {
		var date []byte // size 10
		err := db.QueryRow("SELECT MIN(Datum) FROM Transaktioner").Scan(&date)
		firstYear, err := strconv.Atoi(toUtf8(date)[0:4])
		err = db.QueryRow("SELECT MAX(Datum) FROM Transaktioner").Scan(&date)
		lastYear, err := strconv.Atoi(toUtf8(date)[0:4])

		printMonthly(w, db, accName, accYear, accMonth)

		kontolista := getAccNames()
		var konton []kontoType
		for _,j := range kontolista {
			var k kontoType
			k.Name = j
			if k.Name == accName {
				k.Selected = true
			}
			konton = append(konton, k)
		}
		var years []yearType
		for i := firstYear; i <= lastYear; i++ {
			var k yearType
			k.Name = strconv.Itoa(i)
			if i == accYear {
				k.Selected = true
			}
			years = append(years, k)
		}
		var months []monthType
		for i := 1; i <= 12; i++ {
			var k monthType
			k.Name = strconv.Itoa(i)
			if i == accMonth {
				k.Selected = true
			}
			months = append(months, k)
		}

		nextYear := accYear
		nextMonth := accMonth + 1
		if nextMonth > 12 {
			nextMonth = 1
			nextYear = nextYear + 1
		}
		prevYear := accYear
		prevMonth := accMonth - 1
		if prevMonth < 1 {
			prevMonth = 12
			prevYear = prevYear - 1
		}
		
		t := template.New("Main18")
		t, _ = t.Parse(htmlmain18)
		data := Main18Data{
			Konton: konton,
			Years: years,
			Months: months,
			SelectKonto: accName,
			NextYear: strconv.Itoa(nextYear),
			NextMonth: strconv.Itoa(nextMonth),
			PrevYear: strconv.Itoa(prevYear),
			PrevMonth: strconv.Itoa(prevMonth),
		}
		err = t.Execute(w, data)
		if err != nil {
			log.Println("While serving HTTP main18: ", err)
		}
	}

	t = template.New("Main19")
	t, _ = t.Parse(htmlmain19)
	err = t.Execute(w, nil)
	if err != nil {
		log.Println("While serving HTTP main19: ", err)
	}
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
	res.Close()
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
	res.Close()
	return names
}

//go:embed html/help1.html
var html1 string

func help1(w http.ResponseWriter, req *http.Request) {
	log.Println("Func help1")
	
	t := template.New("Hjälp example")
	t, _ = t.Parse(html1)
	err := t.Execute(w, t)
	if err != nil {
		return
	}
}

//go:embed html/faq1.html
var htmlfaq1 string

func faq1(w http.ResponseWriter, req *http.Request) {
	log.Println("Func faq1")
	
	t := template.New("FAQ example")
	t, _ = t.Parse(htmlfaq1)
	err := t.Execute(w, t)
	if err != nil {
		return
	}
}

//go:embed html/components.html
var htmlcomps string

func comps(w http.ResponseWriter, req *http.Request) {
	log.Println("Func comps")
	
	t := template.New("Components example")
	t, _ = t.Parse(htmlcomps)
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
	http.HandleFunc("/r/e/transaction", r_e_transaction)
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
	http.HandleFunc("/hnewtrans", addtransaction)
	http.HandleFunc("/fixedtrans", fixedtransactionHTML)
	//	http.HandleFunc("/addtrans", addtransaction)
	http.HandleFunc("/monthly", monthly)
	http.HandleFunc("/transactions", transactions)
	http.HandleFunc("/htransactions", htransactions)
	http.HandleFunc("/platser", hanteraplatser)
	http.HandleFunc("/personer", hanterapersoner)
	http.HandleFunc("/konton", hanterakonton)
	http.HandleFunc("/budget", hanteraBudget)
	http.HandleFunc("/yresult", hanteraYResult)
	http.HandleFunc("/ybr", hanteraYBR)
	http.HandleFunc("/summary", generateSummary)
	http.HandleFunc("/acccmp", compareaccount)
	http.HandleFunc("/passwd", passwordmgmt)
	http.HandleFunc("/help1", help1)
	http.HandleFunc("/faq1", faq1)
	http.HandleFunc("/components", comps)
	http.HandleFunc("/", root)

	//ip, _ := externalIP()
	fmt.Println("Öppna URL i webläsaren:  http://localhost:8090/")
	//fmt.Printf(" eller :  http://%s:8090/\n", ip)
	browser.OpenURL("http://localhost:8090/")
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
