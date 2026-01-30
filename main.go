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
// localize decimal.String()

// ROADMAP/TODO/The Future Is In Flux
// ============
// Döp om konto.
//
// 1. Ändra kolumn Benamning i tabellen Konton.
// 2. Ändra alla förekomster (Tillkonto, Frånkonto) i tabellen Transaktioner.
// 3. Ändra alla förekomster (Tillkonto, Frånkonto) i tabellen Överföringar.
//
// Note: Får ej byta namn till existerande konto eller plats. Gäller även för plats. Lägg in spärr vid nytt konto/plats och namnbyte.
//
//
// Vad gör Årskiftesrutin?
//
// 1. Påpekar vikten av säkerhetskopia och att det går att backa till en sådan.
// 2. Fråga om hur budget ska uppdateras:
//    Föregående års utfall
//    Föregående års budget
//    Nollställ budget (sätt alla månader/kategorier till noll)
// 3. Uppdatera budget enligt val
// 4. Uppdatera kolumn SaldoArsskifte i tabellen Konton. Note: Saldot är totalen för kontot, dvs inte bara fram till årsskiftet.
// 5. Uppdatera kolumn ArsskifteManad till "Jan" i tabellen Konton.
//
// Vad gör Säkerhetskopiering/Återföring?
// 1. Gör ny fil med ändelse HBK
// 2. Lägg in headers/strängar i klartext (iso-8859?):
//    Version: "Ver1.00"
//    Path: "C:\...\" fix längd (80bytes) på fältet? Används som förslag på filnamn/path att läsa tillbaka till.
//    En siffra: "1     " (6 bytes)
//    Datum för backupen: "YYMMDDHHMMSS"
//    Storlek på mdb-filen i bytes. (12bytes inkl) Några mellanslag: "     " (troligen fix storlek på fältet för filstorlek)
// 3. kopiera in mdb-filen (troligen byte-för-byte). Börjar med ^A^@ vilket är första icke-skrivbara tecken i HBK-filen också.
//
// BUG: Teckenkodning i lösenord
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
	"embed"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3" // MIT License
	"github.com/pkg/browser"        // BSD-2-Clause
	"github.com/shopspring/decimal" // MIT License
)

// Global variables
const DEBUG_ON bool = true

var db *sql.DB = nil
var nopwDb *sql.DB = nil
var dbtype uint8 = 0 // 0=none, 1=mdb/Access2.0, 2=sqlite3
var currentDatabase = "NONE"
var dbdecimaldot bool = false

//go:embed html/*.html
var htmlTemplates embed.FS

func hello(w http.ResponseWriter, req *http.Request) {
	log.Println("Func hello")

	_, _ = fmt.Fprintf(w, "hello\n")
}

type RootPageData struct {
	FilerFinns bool
	AntalFiler string
	Filnamn    []string
}

func root(w http.ResponseWriter, req *http.Request) {
	log.Println("Func root")

	if db != nil {
		generateSummary(w, req)
		return
	}

	tmpl, _ := template.New("root.html").ParseFS(htmlTemplates, "html/root.html")

	files, err := os.ReadDir(".")
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
	log.Println("Hittade filer", antal, getCurrentFuncName())
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
		_, _ = fmt.Fprintf(w, "%s", "Requested URL "+req.URL.String()+"\n")
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

type Main4Data struct {
	Antal int
}

type Main5Data struct {
	Antal int
}

func checkÖverföringar(w http.ResponseWriter, db *sql.DB) {
	currentTime := time.Now()
	currDate := currentTime.Format("2006-01-02")
	antal := GetCountPendingÖverföringar(db, currDate)
	if antal > 0 {
		t, _ := template.New("main4.html").ParseFS(htmlTemplates, "html/main4.html")
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
		t, _ := template.New("main5.html").ParseFS(htmlTemplates, "html/main5.html")

		data := Main5Data{
			Antal: antal,
		}
		err := t.Execute(w, data)
		if err != nil {
			log.Println("While serving HTTP main5: ", err)
		}
	}
}

type Main1Data struct {
	CurrDBName string
	CurrDate   string
}

func printSummaryHead(w http.ResponseWriter) {
	currentTime := time.Now()
	currDate := currentTime.Format("2006-01-02")

	t, _ := template.New("main1.html").ParseFS(htmlTemplates, "html/main1.html")
	data := Main1Data{
		CurrDBName: currentDatabase,
		CurrDate:   currDate,
	}
	_ = t.Execute(w, data)
}

func printAccounts(w http.ResponseWriter) {
	t, _ := template.New("main6.html").ParseFS(htmlTemplates, "html/main6.html")
	err := t.Execute(w, t)
	if err != nil {
		log.Println("While serving HTTP main6: ", err)
	}
}

type sumType struct {
	Name     string
	Hidden   bool
	DbSaldo  string
	DaySaldo string
	TotSaldo string
}

type Main11Data struct {
	Konton    []sumType
	DoldaNamn []string
}

func printSummaryTable(w http.ResponseWriter, db *sql.DB) {
	// Record the start time
	duration_start := time.Now()

	// Find date of today
	currentTime := time.Now()
	currDate := currentTime.Format("2006-01-02")

	// Find accounts
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
		konton = append(konton, sumType{acc, false, Dec2Str(DbSaldo), Dec2Str(DaySaldo), Dec2Str(TotSaldo)})
	}
	res.Close()

	// Dölj oanvända konton
	var nolla = decimal.NewFromInt(0)
	var kontonDolda []sumType
	var doldaNamn []string
	for _, konto := range konton {
		DaySaldo, err2 := decimal.NewFromString(
			strings.ReplaceAll(
				strings.ReplaceAll(konto.DaySaldo, ",", "."),
				" ", ""))
		if err2 != nil {
			log.Fatal("Dölj konton:", err2)
		}

		if DaySaldo.Equal(nolla) {
			now := time.Now()
			currentYear, currentMonth, currentDay := now.Date()
			currentLocation := now.Location()

			today := time.Date(currentYear, currentMonth, currentDay, 0, 0, 0, 0, currentLocation)
			firstofperiod := today.AddDate(0, -3, 0)

			trans := getTransactionsInDateRange(db, konto.Name, firstofperiod.Format("2006-01-02"), today.Format("2006-01-02"))
			if len(trans) == 0 {
				konto.Hidden = true
				doldaNamn = append(doldaNamn, konto.Name)
			}
		}
		kontonDolda = append(kontonDolda, konto)
	}

	t, _ := template.New("main11.html").ParseFS(htmlTemplates, "html/main11.html")
	data := Main11Data{
		Konton:    kontonDolda,
		DoldaNamn: doldaNamn,
	}
	err = t.Execute(w, data)
	if err != nil {
		log.Println("While serving HTTP main11: ", err)
	}
	if DEBUG_ON {
		endtime := time.Now()
		log.Println("Duration for print summary: ", endtime.Sub(duration_start))
	}
}

func checkpwd(w http.ResponseWriter, req *http.Request) {
	log.Println("Func checkpwd")

	if nopwDb == nil {
		t, _ := template.New("main12.html").ParseFS(htmlTemplates, "html/main12.html")
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
			t, _ := template.New("main13.html").ParseFS(htmlTemplates, "html/main13.html")
			err := t.Execute(w, nil)
			if err != nil {
				log.Println("While serving HTTP main13: ", err)
			}
		}
	}
}

func showsummary(w http.ResponseWriter) {
	t, _ := template.New("main3.html").ParseFS(htmlTemplates, "html/main3.html")
	err := t.Execute(w, t)
	if err != nil {
		log.Println("While serving HTTP main3: ", err)
	}
}

type Main14Data struct {
	Filnamn string
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
			t, _ := template.New("main14.html").ParseFS(htmlTemplates, "html/main14.html")
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
		t, _ := template.New("main15.html").ParseFS(htmlTemplates, "html/main15.html")
		err := t.Execute(w, nil)
		if err != nil {
			log.Println("While serving HTTP main15: ", err)
		}
	} else {
		pwd := getdbpw(nopwDb)
		if pwd != " " {
			t, _ := template.New("main16.html").ParseFS(htmlTemplates, "html/main16.html")
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

func closedb(w http.ResponseWriter, req *http.Request) {
	closeDB()

	t, _ := template.New("main3.html").ParseFS(htmlTemplates, "html/main3.html")
	err := t.Execute(w, t)
	if err != nil {
		log.Println("While serving HTTP main3: ", err)
	}
}

func quitapp(w http.ResponseWriter, req *http.Request) {
	t, _ := template.New("main8.html").ParseFS(htmlTemplates, "html/main8.html")
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

type Main9Data struct {
	Filnamn string
}

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

		t, _ := template.New("main9.html").ParseFS(htmlTemplates, "html/main9.html")
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
	t, _ := template.New("main10.html").ParseFS(htmlTemplates, "html/main10.html")
	data := Main10Data{
		Filnamn: filename,
	}
	err = t.Execute(w, data)
	if err != nil {
		log.Println("While serving HTTP main10: ", err)
	}
}

func generateSummary(w http.ResponseWriter, req *http.Request) {
	log.Println("Func generateSummary")

	printSummaryHead(w)
	if db != nil {
		checkÖverföringar(w, db)
		t, _ := template.New("main2.html").ParseFS(htmlTemplates, "html/main2.html")
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
	X          string
	Y          string
	Width      string
	Height     string
	Day        string
	Saldo      string
	IsPositive string
}

type GridLineType struct {
	X1     string
	Y1     string
	X2     string
	Y2     string
	IsZero bool
}

type AxisLabelType struct {
	X       string
	Y       string
	Text    string
	IsXAxis bool
}

type TextType struct {
	X    string
	Y    string
	Text string
}

type Main20Data struct {
	Filename     string
	AccName      string
	Year         string
	Month        string
	SaldoDatum   string
	SlutSaldo    string
	Transactions []TransactionType
	ZeroLine     string
	MonthValues  []MonthValueType
	Texts        []TextType
	GridLines    []GridLineType
	AxisLabels   []AxisLabelType
	MinSaldo     string
	MaxSaldo     string
	ChartTitle   string
}

func printMonthly(w http.ResponseWriter, db *sql.DB, accName string, accYear int, accMonth int) {
	var transactions []TransactionType
	var monthValues []MonthValueType
	var texts []TextType
	var gridLines []GridLineType
	var axisLabels []AxisLabelType

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

	_ = db.QueryRowContext(ctx,
		`select startsaldo
  from konton
  where benämning = ?`, accName).Scan(&rawStart)
	res2 := toUtf8(rawStart)
	startSaldo, _ := decimal.NewFromString(res2)
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

	// Beräkna saldo från kontot skapades till och med sista dagen i perioden
	for res.Next() {
		_ = res.Scan(&fromAcc, &toAcc, &tType, &date, &what, &who, &amount, &nummer, &saldo, &fixed, &comment)
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

	// Fyll på med saknade dagar
	for i := 2; i < 32; i++ {
		if !dayFound[i] {
			daySaldo[i] = daySaldo[i-1]
			dayFound[i] = true
		}
	}

	minSaldo := decimal.NewFromInt(math.MaxInt64)
	maxSaldo := decimal.NewFromInt(math.MinInt64)

	// Hitta max och min för perioden
	for i := 1; i < 32; i++ {
		if daySaldo[i].GreaterThan(maxSaldo) {
			maxSaldo = daySaldo[i]
		}
		if daySaldo[i].LessThan(minSaldo) {
			minSaldo = daySaldo[i]
		}
	}

	// Se till att 0 finns med
	if decimal.Zero.LessThan(minSaldo) {
		minSaldo = decimal.Zero
	}
	if decimal.Zero.GreaterThan(maxSaldo) {
		maxSaldo = decimal.Zero
	}

	// Round down minSaldo to integer and limit to 3 significant digits
	minSaldo = minSaldo.Floor()
	minVal, _ := minSaldo.Float64()
	if minVal != 0 {
		// Calculate magnitude for 3 significant digits
		magnitude := math.Pow(10, math.Floor(math.Log10(math.Abs(minVal)))-2)
		// Round down to previous multiple that gives 3 significant digits
		rounded := math.Floor(minVal/magnitude) * magnitude
		minSaldo = decimal.NewFromFloat(rounded)
	}

	// Round up maxSaldo to integer and limit to 3 significant digits
	maxSaldo = maxSaldo.Ceil()
	maxVal, _ := maxSaldo.Float64()
	if maxVal != 0 {
		// Calculate magnitude for 3 significant digits
		magnitude := math.Pow(10, math.Floor(math.Log10(math.Abs(maxVal)))-2)
		// Round up to next multiple that gives 3 significant digits
		rounded := math.Ceil(maxVal/magnitude) * magnitude
		maxSaldo = decimal.NewFromFloat(rounded)
	}

	var yf, val float64
	var y int
	var y1 decimal.Decimal
	const colWidth = 20
	const chartHeight = 500
	const chartWidth = 900
	const marginLeft = 60
	const marginRight = 100
	const marginTop = 40
	const marginBottom = 50
	const plotWidth = chartWidth - marginLeft - marginRight
	const plotHeight = chartHeight - marginTop - marginBottom
	var monthValue MonthValueType
	zeroDecimal := decimal.Zero

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
		if y2f > 0 {
			yf = val / (y2f / float64(plotHeight))
		} else {
			yf = 0
		}
		y = int(yf)

		monthValue.X = strconv.Itoa(marginLeft + (i-1)*colWidth)
		monthValue.Y = strconv.Itoa(marginTop + y)
		monthValue.Width = strconv.Itoa(colWidth - 1)
		monthValue.Height = strconv.Itoa(plotHeight - y)
		monthValue.Day = strconv.Itoa(i)
		monthValue.Saldo = Dec2Str(currSaldo)
		if currSaldo.GreaterThanOrEqual(zeroDecimal) {
			monthValue.IsPositive = "true"
		} else {
			monthValue.IsPositive = "false"
		}
		monthValues = append(monthValues, monthValue)
	}
	// zero line - calculate where saldo = 0 should be positioned
	zeroDecimal = decimal.Zero
	var zeroLine int
	if maxSaldo.GreaterThan(zeroDecimal) && minSaldo.LessThan(zeroDecimal) {
		// Zero is between min and max
		y1 = maxSaldo.Sub(zeroDecimal)
		val, _ = y1.Float64()
		y2 := maxSaldo.Sub(minSaldo)
		y2f, _ := y2.Float64()
		if y2f > 0 {
			yf = val / (y2f / float64(plotHeight))
		} else {
			yf = 0
		}
		y = int(yf)
		zeroLine = marginTop + y
	} else if minSaldo.GreaterThanOrEqual(zeroDecimal) {
		// All values are positive, zero line is at bottom
		zeroLine = marginTop + plotHeight
	} else {
		// All values are negative, zero line is at top
		zeroLine = marginTop
	}

	// Generate grid lines (horizontal)
	var gridLine GridLineType
	numGridLines := 5
	for i := 0; i <= numGridLines; i++ {
		gridY := marginTop + (i * plotHeight / numGridLines)
		gridLine.X1 = strconv.Itoa(marginLeft)
		gridLine.Y1 = strconv.Itoa(gridY)
		gridLine.X2 = strconv.Itoa(marginLeft + plotWidth)
		gridLine.Y2 = strconv.Itoa(gridY)
		// Check if this grid line is close to the zero line (within 2 pixels)
		diff := gridY - zeroLine
		if diff < 0 {
			diff = -diff
		}
		if diff <= 2 {
			gridLine.IsZero = true
		} else {
			gridLine.IsZero = false
		}
		gridLines = append(gridLines, gridLine)
	}

	// Always add a dedicated zero line if it's not already in the grid
	zeroLineInGrid := false
	for _, gl := range gridLines {
		if gl.IsZero {
			zeroLineInGrid = true
			break
		}
	}
	if !zeroLineInGrid && zeroLine >= marginTop && zeroLine <= marginTop+plotHeight {
		gridLine.X1 = strconv.Itoa(marginLeft)
		gridLine.Y1 = strconv.Itoa(zeroLine)
		gridLine.X2 = strconv.Itoa(marginLeft + plotWidth)
		gridLine.Y2 = strconv.Itoa(zeroLine)
		gridLine.IsZero = true
		gridLines = append(gridLines, gridLine)
	}

	// Generate vertical grid lines for key days
	for i := 1; i <= 31; i += 5 {
		gridX := marginLeft + (i-1)*colWidth
		gridLine.X1 = strconv.Itoa(gridX)
		gridLine.Y1 = strconv.Itoa(marginTop)
		gridLine.X2 = strconv.Itoa(gridX)
		gridLine.Y2 = strconv.Itoa(marginTop + plotHeight)
		gridLine.IsZero = false
		gridLines = append(gridLines, gridLine)
	}

	// Generate axis labels for days (X-axis)
	var axisLabel AxisLabelType
	// Day labels: 1, 5, 10, 15, 20, 25, 30
	dayLabels := []int{1, 5, 10, 15, 20, 25, 30}
	for _, day := range dayLabels {
		axisLabel.X = strconv.Itoa(marginLeft + (day-1)*colWidth + colWidth/2)
		axisLabel.Y = strconv.Itoa(marginTop + plotHeight + 20)
		axisLabel.Text = strconv.Itoa(day)
		axisLabel.IsXAxis = true
		axisLabels = append(axisLabels, axisLabel)
	}

	// Generate Y-axis labels (saldo values)
	numYLabels := 6
	for i := 0; i <= numYLabels; i++ {
		labelY := marginTop + (i * plotHeight / numYLabels)
		ratio := decimal.NewFromInt(int64(i)).Div(decimal.NewFromInt(int64(numYLabels)))
		rangeVal := maxSaldo.Sub(minSaldo)
		offset := rangeVal.Mul(ratio)
		saldoValue := maxSaldo.Sub(offset)
		axisLabel.X = strconv.Itoa(marginLeft - 10)
		axisLabel.Y = strconv.Itoa(labelY + 5)
		axisLabel.Text = Dec2Str(saldoValue)
		axisLabel.IsXAxis = false
		axisLabels = append(axisLabels, axisLabel)
	}

	// Keep old text labels for backward compatibility (but we'll use axisLabels in template)
	var text TextType
	text.X = "0"
	text.Y = "550"
	text.Text = "1"
	texts = append(texts, text)
	text.X = strconv.Itoa((10 - 1) * colWidth)
	text.Y = "550"
	text.Text = "10"
	texts = append(texts, text)
	text.X = strconv.Itoa((20 - 1) * colWidth)
	text.Y = "550"
	text.Text = "20"
	texts = append(texts, text)
	text.X = strconv.Itoa((30 - 1) * colWidth)
	text.Y = "550"
	text.Text = "30"
	texts = append(texts, text)

	text.X = strconv.Itoa(33 * colWidth)
	text.Y = "10"
	text.Text = maxSaldo.String()
	texts = append(texts, text)
	text.X = strconv.Itoa(33 * colWidth)
	text.Y = "500"
	text.Text = minSaldo.String()
	texts = append(texts, text)

	now := time.Now()
	currentLocation := now.Location()
	firstOfMonth := time.Date(accYear, time.Month(accMonth), 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)
	saldoDatum := lastOfMonth
	DaySaldo, _ := saldonKonto(db, accName, lastOfMonth.Format("2006-01-02"))

	// Generate chart title
	monthNames := []string{"", "Januari", "Februari", "Mars", "April", "Maj", "Juni",
		"Juli", "Augusti", "September", "Oktober", "November", "December"}
	chartTitle := fmt.Sprintf("Saldo per dag - %s %d", monthNames[accMonth], accYear)

	t, _ := template.New("main20.html").ParseFS(htmlTemplates, "html/main20.html")
	data := Main20Data{
		Filename:     currentDatabase,
		AccName:      accName,
		Year:         strconv.Itoa(accYear),
		Month:        strconv.Itoa(accMonth),
		SaldoDatum:   saldoDatum.Format("2006-01-02"),
		SlutSaldo:    Dec2Str(DaySaldo),
		Transactions: transactions,
		ZeroLine:     strconv.Itoa(zeroLine),
		MonthValues:  monthValues,
		Texts:        texts,
		GridLines:    gridLines,
		AxisLabels:   axisLabels,
		MinSaldo:     Dec2Str(minSaldo),
		MaxSaldo:     Dec2Str(maxSaldo),
		ChartTitle:   chartTitle,
	}
	err = t.Execute(w, data)
	if err != nil {
		log.Println("While serving HTTP main20: ", err)
	}

}

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

type Main18Data struct {
	Konton      []kontoType
	Years       []yearType
	Months      []monthType
	SelectKonto string
	NextYear    string
	NextMonth   string
	PrevYear    string
	PrevMonth   string
}

func monthly(w http.ResponseWriter, req *http.Request) {
	t, _ := template.New("main17.html").ParseFS(htmlTemplates, "html/main17.html")
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
		// data from form
		accYear, _ = strconv.Atoi(req.FormValue("accYear"))
		accMonth, _ = strconv.Atoi(req.FormValue("accMonth"))
		accName = req.FormValue("accName")
	} else {
		// default data
		var date []byte // size 10
		// find date of last transaction
		_ = db.QueryRow("SELECT MAX(Datum) FROM Transaktioner").Scan(&date)
		accYear, _ = strconv.Atoi(toUtf8(date)[0:4])
		accMonth, _ = strconv.Atoi(toUtf8(date)[5:7])

		// adjust date to today if last is later
		now := time.Now()
		currentYear, currentMonth, _ := now.Date()
		if accYear > currentYear {
			accYear = currentYear
			accMonth = int(currentMonth)
		} else if (accYear == currentYear) && (accMonth > int(currentMonth)) {
			accMonth = int(currentMonth)
		}

		var namn []byte // size 10
		_ = db.QueryRow("SELECT TOP 1 Benämning FROM Konton").Scan(&namn)
		accName = toUtf8(namn)
	}

	if db != nil {
		var date []byte // size 10
		_ = db.QueryRow("SELECT MIN(Datum) FROM Transaktioner").Scan(&date)
		firstYear, _ := strconv.Atoi(toUtf8(date)[0:4])
		_ = db.QueryRow("SELECT MAX(Datum) FROM Transaktioner").Scan(&date)
		lastYear, _ := strconv.Atoi(toUtf8(date)[0:4])

		printMonthly(w, db, accName, accYear, accMonth)

		kontolista := getAccNames()
		var konton []kontoType
		for _, j := range kontolista {
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

		t, _ := template.New("main18.html").ParseFS(htmlTemplates, "html/main18.html")
		data := Main18Data{
			Konton:      konton,
			Years:       years,
			Months:      months,
			SelectKonto: accName,
			NextYear:    strconv.Itoa(nextYear),
			NextMonth:   strconv.Itoa(nextMonth),
			PrevYear:    strconv.Itoa(prevYear),
			PrevMonth:   strconv.Itoa(prevMonth),
		}
		err = t.Execute(w, data)
		if err != nil {
			log.Println("While serving HTTP main18: ", err)
		}
	}

	t, _ = template.New("main19.html").ParseFS(htmlTemplates, "html/main19.html")
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
		_ = res.Scan(&Typ)
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
		_ = res.Scan(&Typ)
		names = append(names, toUtf8(Typ))
	}
	res.Close()
	return names
}

func help1(w http.ResponseWriter, req *http.Request) {
	log.Println("Func help1")

	t, _ := template.New("help1.html").ParseFS(htmlTemplates, "html/help1.html")
	err := t.Execute(w, t)
	if err != nil {
		return
	}
}

func faq1(w http.ResponseWriter, req *http.Request) {
	log.Println("Func faq1")

	t, _ := template.New("faq1.html").ParseFS(htmlTemplates, "html/faq1.html")
	err := t.Execute(w, t)
	if err != nil {
		return
	}
}

func comps(w http.ResponseWriter, req *http.Request) {
	log.Println("Func comps")

	t, _ := template.New("components.html").ParseFS(htmlTemplates, "html/components.html")
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
	//http.HandleFunc("/r/e/fastutg", r_e_fastutg)
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
	http.HandleFunc("/editfixedtrans", editfixedtransactionHTML)
	http.HandleFunc("/editfastutg", editfastutgHTML)
	http.HandleFunc("/editfastink", editfastinkHTML)
	http.HandleFunc("/editfastovf", editfastovfHTML)
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
