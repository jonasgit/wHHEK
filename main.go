//-*- coding: utf-8 -*-

// This is a personal finance package, inspired by Hogia Hemekonomi.
// Detta är ett hemekonomiprogram, inspirerat av Hogia Hemekonomi från 90-talet.

// System Requirements: Windows 10 (any)

// To build on Windows:
// Prepare: install gnu emacs: emacs-26.3-x64_64 (optional)
// Prepare: TDM-GCC from https://jmeubank.github.io/tdm-gcc/
//https://github.com/jmeubank/tdm-gcc/releases/download/v9.2.0-tdm-1/tdm-gcc-9.2.0.exe

// Prepare: install git: Git-2.23.0-64-bit
// Prepare: install golang 32-bits (can't access access/jet/mdb driver using 64-bits)
//   go1.13.3.windows-386.msi
// Prepare: go get github.com/alexbrainman/odbc
// Prepare: go get github.com/mattn/go-sqlite3
// Prepare: go get golang.org/x/text/encoding/charmap
// Prepare: go get github.com/shopspring/decimal
// Build: go build -o wHHEK.exe main.go
// Build release: go build -ldflags="-s -w" -o wHHEK.exe main.go
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

// ROADMAP/TODO/The Future Is In Flux
// ============
// escape & in all html. escapeHTML verkar inte fungera. Use template?
// kommandorads help-option
// kommandoradsoption för att välja katalog med databas
// Installationsinstruktion (ladda ner exe, skapa ikon, brandvägg)
// Efter lagt till transaktion, visa den tillagda
// Visa platser
// Lägg till ny plats
// Redigera plats
// Visa personer
// Lägg till ny person
// Redigera person
// Graf som i månadsvyn fast för senaste året
// Lägg till nytt konto
// Redigera konto
// Visa fasta överföringar
// Lägg till fast överföring
// Redigera fast överföring
// Visa fasta betalningar
// Lägg till fast betalning
// Redigera fast betalning
// Registrera fasta överföringar
// Registrera fasta betalningar
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
// Visa budget
// Redigera budget
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
	"fmt"
	"io/ioutil"
	"log"
	"golang.org/x/text/encoding/charmap"
	"net/http"
	"errors"
	"math"
	"net"
	//	"flag"
	"os"
	"strings"
	"strconv"
	"time"
	
	_ "github.com/alexbrainman/odbc"    // BSD-3-Clause License 
	_ "github.com/mattn/go-sqlite3"     // MIT License
	"github.com/shopspring/decimal"     // MIT License
)

// Global variables
var db *sql.DB = nil
var dbtype uint8 = 0 // 0=none, 1=mdb/Access2.0, 2=sqlite3
var currentDatabase string = "NONE"
var (
	ctx context.Context
)

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

func escapeHTML(stringVal string) string {
	stringVal2 := strings.ReplaceAll(stringVal, "&", "&amp;")
	stringVal3 := strings.ReplaceAll(stringVal2, "\"", "&quot;")
	stringVal4 := strings.ReplaceAll(stringVal3, "<", "&lt;")
	stringVal5 := strings.ReplaceAll(stringVal4, ">", "&gt;")
	return stringVal5
}

func hello(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "hello\n")
}

func root(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "<html>\n")
	fmt.Fprintf(w, "<body>\n")
	//fmt.Fprintf(w, "hello root<p>")
	fmt.Fprintf(w, "Välj databas att arbeta med:<br>")
	
	files, err := ioutil.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}

	if (len(files) >0) {
		fmt.Fprintf(w,"<form method=\"POST\" action=\"/open\">\n")
		for _, file := range files {
			
			if(strings.HasSuffix(strings.ToLower(file.Name()), ".mdb") ||
				strings.HasSuffix(strings.ToLower(file.Name()), ".db")) {
				//fmt.Fprintf(w,"%s<br>\n", file.Name())
				fmt.Fprintf(w,"<input type=\"radio\" id=\"%s\" name=\"fname\" value=\"%s\"><label for=\"%s\">%s</label><br>\n", file.Name(), file.Name(), file.Name(), file.Name())
			}
		}
		fmt.Fprintf(w,"<input type=\"submit\" value=\"Submit\"></form>\n")
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

	return fname;
}

func openJetDB(filename string, ro bool) *sql.DB {
	readonlyCommand := ""
	if ro {
		readonlyCommand = "READONLY;"
	}
	
	databaseAccessCommand := "Driver={Microsoft Access Driver (*.mdb)};"+
		readonlyCommand +
		"DBQ="+filename
	//fmt.Println("Database access command: "+databaseAccessCommand)
	db, err := sql.Open("odbc",
		databaseAccessCommand)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	return db
}

func openSqlite(filename string) *sql.DB {
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func printAccounts(w http.ResponseWriter, db *sql.DB) {
	fmt.Fprintf(w, "<html>\n")
	fmt.Fprintf(w, "<head>\n")
	fmt.Fprintf(w, "<style>\n")
	fmt.Fprintf(w, "table,th,td { border: 1px solid black }\n")
	fmt.Fprintf(w, "</style>\n")
	fmt.Fprintf(w, "</head>\n")
	fmt.Fprintf(w, "<body>\n")

	fmt.Fprintf(w, "<h1>%s</h1>\n", currentDatabase)

	res, err := db.Query("SELECT KontoNummer,Benämning,Saldo,StartSaldo,StartManad,Löpnr,SaldoArsskifte,ArsskifteManad FROM Konton")

	if err != nil {
		log.Fatal(err)
		os.Exit(2)
	}

	var KontoNummer []byte  // size 20
	var Benämning  []byte  // size 40, index
	var Saldo []byte  // BCD / Decimal Precision 19
	var StartSaldo []byte  // BCD / Decimal Precision 19
	var StartManad []byte  // size 10
	var Löpnr  []byte  // autoinc Primary Key
	var SaldoArsskifte []byte  // BCD / Decimal Precision 19
	var ArsskifteManad []byte  // size 10

	fmt.Fprintf(w, "<table style=\"width:100%%\"><tr><th>Kontonamn</th><th>Saldo (kanske?)</th>\n")
	for res.Next() {
		err = res.Scan(&KontoNummer,&Benämning,&Saldo,&StartSaldo,&StartManad,&Löpnr,&SaldoArsskifte,&ArsskifteManad)

		fmt.Fprintf(w, "<tr><td>%s</td><td>%s</td>\n", toUtf8(Benämning), toUtf8(Saldo))
	}
	fmt.Fprintf(w, "</table>\n")
}

func opendb(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	filename := sanitizeFilename(req.FormValue("fname"))

	if(strings.HasSuffix(strings.ToLower(filename), ".mdb")) {
		//fmt.Fprintf(w, "Trying to open Access/Jet<br>\n")
		db = openJetDB(filename, true)
		currentDatabase = filename
		dbtype=1;
	} else {
		if(strings.HasSuffix(strings.ToLower(filename), ".db")) {
			//fmt.Fprintf(w, "Trying to open sqlite3<br>\n")
			db = openSqlite(filename)
			currentDatabase = filename
			dbtype=2;
		} else {
			fmt.Fprintf(w, "Bad filename: %s<br>\n", filename)
		}
	}

	if(db==nil) {
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
		fmt.Fprintf(w, "      <p>Hello HTML5!</p>\n")
		fmt.Fprintf(w, "   </body>\n")
		fmt.Fprintf(w, "</html>\n")
	}
	fmt.Fprintf(w, "</body>\n")
	fmt.Fprintf(w, "</html>\n")
}

func closedb(w http.ResponseWriter, req *http.Request) {
	db.Close()
	dbtype=0;
	currentDatabase = "NONE"
	
	fmt.Fprintf(w, "<!DOCTYPE html>\n")
	fmt.Fprintf(w, "<html>\n")
	fmt.Fprintf(w, "   <head>\n")
	fmt.Fprintf(w, "      <title>HTML Meta Tag</title>\n")
	fmt.Fprintf(w, "      <meta http-equiv = \"refresh\" content = \"0; url = /\" />\n")
	fmt.Fprintf(w, "   </head>\n")
	fmt.Fprintf(w, "   <body>\n")
	fmt.Fprintf(w, "      <p>Hello HTML5!</p>\n")
	fmt.Fprintf(w, "   </body>\n")
	fmt.Fprintf(w, "</html>\n")
}

func generateSummary(w http.ResponseWriter, req *http.Request) {
	printAccounts(w, db)
	fmt.Fprintf(w, "<a href=\"monthly\">Månads kontoutdrag</a><p>\n")
	fmt.Fprintf(w, "<a href=\"transactions\">Transaktionslista</a><p>\n")
	fmt.Fprintf(w, "<a href=\"newtrans\">Ny transaktion</a><p>\n")
	fmt.Fprintf(w, "<a href=\"close\">Stäng databas</a><p>\n")
	fmt.Fprintf(w, "</body>\n")
	fmt.Fprintf(w, "</html>\n")
}

func printMonthly(w http.ResponseWriter, db *sql.DB, accName string, accYear int, accMonth int) {
	fmt.Fprintf(w, "<h1>%s</h1>\n", currentDatabase)
	fmt.Fprintf(w, "Kontonamn: %s<br>\n", accName)
	fmt.Fprintf(w, "År: %d<br>\n", accYear)
	fmt.Fprintf(w, "Månad: %d<br>\n", accMonth)

	var startDate,endDate string
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
	var rawStart []byte     // size 16
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

	var fromAcc []byte  // size 40
	var toAcc []byte    // size 40
	var tType []byte    // size 40
	var date []byte     // size 10
	var what []byte     // size 40
	var who []byte      // size 50
	var amount []byte   // BCD / Decimal Precision 19
	var nummer int      // Autoinc Primary Key, index
	var saldo []byte    // BCD / Decimal Precision 19
	var fixed bool      // Boolean
	var comment []byte  // size 60

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
			( (toUtf8(tType) == "Uttag") ||
				(toUtf8(tType) == "Inköp") ||
				(toUtf8(tType) == "Fast Utgift") ||
				(toUtf8(tType) == "Överföring")) {
			currSaldo = currSaldo.Sub(decAmount)
		}
 		if toUtf8(date) >= startDate {
			sqlStmt:=""
			sqlStmt+="<tr><td>" + strconv.Itoa(nummer) + "</td>"
			sqlStmt+="<td>" + toUtf8(fromAcc) + "</td>"
			sqlStmt+="<td>" + toUtf8(toAcc) + "</td>"
			sqlStmt+="<td>" + toUtf8(tType) + "</td>"
			sqlStmt+="<td>" + toUtf8(what) + "</td>"
			sqlStmt+="<td>" + toUtf8(date) + "</td>"
			sqlStmt+="<td>" + toUtf8(who) + "</td>"
			sqlStmt+="<td>" + toUtf8(amount) + "</td>"
			sqlStmt+="<td>" + currSaldo.String() + "</td>"
			sqlStmt+="<td>" + escapeHTML(toUtf8(comment)) + "</td>\n"
			sqlStmt+="<td>" + strconv.FormatBool(fixed) + "</td></tr>"
			fmt.Fprintf(w, sqlStmt)
			
			daynum,_ := strconv.Atoi(toUtf8(date)[8:10])
			day_saldo[daynum] = currSaldo
			day_found[daynum] = true
		}
	}
	fmt.Fprintf(w, "</table>\n")

	min_saldo := decimal.NewFromInt(math.MaxInt64)
	max_saldo := decimal.NewFromInt(math.MinInt64)

	for i := 1; i<32; i++ {
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
	for i := 1; i<32; i++ {
		if day_found[i] {
			y1 = max_saldo.Sub(day_saldo[i])
			currSaldo = day_saldo[i]
		} else {
			y1 = max_saldo.Sub(currSaldo)
		}
		val,_ = y1.Float64()
		y2 := max_saldo.Sub(min_saldo)
		y2f,_ := y2.Float64()
		yf = val / (y2f / (500.0-0.0))
		y = int(yf)
		fmt.Fprintf(w, "  <rect x=%d y=%d width=\"%d\" height=\"%d\"", (i-1)*colWidth, int(y), colWidth, 500-int(y))
		fmt.Fprintf(w, "  style=\"fill:rgb(0,0,255);stroke-width:1;stroke:rgb(0,0,0)\" />\n")
	}
	// zero line
	y1 = max_saldo //.Sub(0.0)
	val,_ = y1.Float64()
	y2 := max_saldo.Sub(min_saldo)
	y2f,_ := y2.Float64()
	yf = val / (y2f / (500.0-0.0))
	y = int(yf)
	fmt.Fprintf(w, "  <rect x=%d y=%d width=\"%d\" height=\"%d\"", 0, y, 900, 1)
	fmt.Fprintf(w, "  style=\"fill:rgb(0,0,255);stroke-width:1;stroke:rgb(0,0,0)\" />\n")
	fmt.Fprintf(w, "<text fill=\"#000000\" font-size=\"12\" font-family=\"Verdana\" x=\"0\" y=\"550\">1</text>\n")
	fmt.Fprintf(w, "<text fill=\"#000000\" font-size=\"12\" font-family=\"Verdana\" x=\"%d\" y=\"550\">10</text>\n", (10-1)*colWidth)
	fmt.Fprintf(w, "<text fill=\"#000000\" font-size=\"12\" font-family=\"Verdana\" x=\"%d\" y=\"550\">20</text>\n", (20-1)*colWidth)
	fmt.Fprintf(w, "<text fill=\"#000000\" font-size=\"12\" font-family=\"Verdana\" x=\"%d\" y=\"550\">30</text>\n", (30-1)*colWidth)

	fmt.Fprintf(w, "<text fill=\"#000000\" font-size=\"12\" font-family=\"Verdana\" x=\"%d\" y=\"10\">%s</text>\n", 33*colWidth, max_saldo.String())
	fmt.Fprintf(w, "<text fill=\"#000000\" font-size=\"12\" font-family=\"Verdana\" x=\"%d\" y=\"500\">%s</text>\n", 33*colWidth, min_saldo.String())
	fmt.Fprintf(w, "Sorry, your browser does not support inline SVG.\n");
	fmt.Fprintf(w, "</svg>\n");
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

	var accYear,accMonth int
	var accName string
	
	if len(req.FormValue("accYear"))>3 {
		accYear,err = strconv.Atoi(req.FormValue("accYear"))
		accMonth,err = strconv.Atoi(req.FormValue("accMonth"))
		accName = req.FormValue("accName")
	} else {
		res1 := db.QueryRow("SELECT MAX(Datum) FROM Transaktioner")
		var date []byte     // size 10
		err = res1.Scan(&date)
		accYear,err = strconv.Atoi(toUtf8(date)[0:4])
		accMonth,err = strconv.Atoi(toUtf8(date)[5:7])
		res1 = db.QueryRow("SELECT TOP 1 Benämning FROM Konton")
		var namn []byte     // size 10
		err = res1.Scan(&namn)
		accName = toUtf8(namn)
	}	

	if(db==nil) {
		fmt.Fprintf(w, "Monthly: No database open<p>\n")
	} else {
		res1 := db.QueryRow("SELECT MIN(Datum) FROM Transaktioner")
		var date []byte     // size 10
		err = res1.Scan(&date)
		firstYear,err := strconv.Atoi(toUtf8(date)[0:4])
		res1 = db.QueryRow("SELECT MAX(Datum) FROM Transaktioner")
		err = res1.Scan(&date)
		lastYear,err := strconv.Atoi(toUtf8(date)[0:4])
		
		printMonthly(w, db, accName, accYear, accMonth)
		//fmt.Fprintf(w, "<a href=\"monthly\">Månads kontoutdrag</a>\n")
		fmt.Fprintf(w,"<form method=\"POST\" action=\"/monthly\">\n")
		fmt.Fprintf(w,"<select id=\"accName\" name=\"accName\">\n")
		res, err := db.Query("SELECT KontoNummer,Benämning,Saldo,StartSaldo,StartManad,Löpnr,SaldoArsskifte,ArsskifteManad FROM Konton order by Benämning")
		
		if err != nil {
			log.Fatal(err)
			os.Exit(2)
		}
		
		var KontoNummer []byte  // size 20
		var Benämning  []byte  // size 40, index
		var Saldo []byte  // BCD / Decimal Precision 19
		var StartSaldo []byte  // BCD / Decimal Precision 19
		var StartManad []byte  // size 10
		var Löpnr  []byte  // autoinc Primary Key
		var SaldoArsskifte []byte  // BCD / Decimal Precision 19
		var ArsskifteManad []byte  // size 10
		for res.Next() {
			err = res.Scan(&KontoNummer,&Benämning,&Saldo,&StartSaldo,&StartManad,&Löpnr,&SaldoArsskifte,&ArsskifteManad)

			fmt.Fprintf(w,"<option value=\"%s\"", toUtf8(Benämning))
			if toUtf8(Benämning) == accName {
				fmt.Fprintf(w," selected ")
			}
			fmt.Fprintf(w,">%s</option>\n", toUtf8(Benämning))
		}
		fmt.Fprintf(w,"</select>\n")
		fmt.Fprintf(w,"<select id=\"accYear\" name=\"accYear\">\n")
		for year := firstYear; year<=lastYear; year++ {
			fmt.Fprintf(w,"<option value=\"%d\"", year)
			if year == accYear {
				fmt.Fprintf(w," selected ")
			}
			fmt.Fprintf(w,">%d</option>\n", year)
		}
		fmt.Fprintf(w,"</select>\n")
		fmt.Fprintf(w,"<select id=\"accMonth\" name=\"accMonth\">\n")
		for month := 1; month<13; month++ {
			fmt.Fprintf(w,"<option value=\"%d\"", month)
			if month == accMonth {
				fmt.Fprintf(w," selected ")
			}
			fmt.Fprintf(w,">%d</option>\n", month)
		}
		fmt.Fprintf(w,"</select>\n")

		fmt.Fprintf(w,"<input type=\"submit\" value=\"Visa\"></form>\n")

		fmt.Fprintf(w,"<form method=\"POST\" action=\"/monthly\">\n")
		fmt.Fprintf(w,"<input type=\"hidden\" id=\"accName\" name=\"accName\" value=\"%s\">", accName)
		if accMonth+1 > 12 {
			fmt.Fprintf(w,"<input type=\"hidden\" id=\"accYear\" name=\"accYear\" value=\"%d\">", accYear+1)
			fmt.Fprintf(w,"<input type=\"hidden\" id=\"accMonth\" name=\"accMonth\" value=\"%d\">", 1)
		} else {
			fmt.Fprintf(w,"<input type=\"hidden\" id=\"accYear\" name=\"accYear\" value=\"%d\">", accYear)
			fmt.Fprintf(w,"<input type=\"hidden\" id=\"accMonth\" name=\"accMonth\" value=\"%d\">", accMonth+1)
		}
		fmt.Fprintf(w,"<input type=\"submit\" value=\"Nästa månad\"></form>\n")

		fmt.Fprintf(w,"<form method=\"POST\" action=\"/monthly\">\n")
		fmt.Fprintf(w,"<input type=\"hidden\" id=\"accName\" name=\"accName\" value=\"%s\">", accName)
		if accMonth-1 < 1 {
			fmt.Fprintf(w,"<input type=\"hidden\" id=\"accYear\" name=\"accYear\" value=\"%d\">", accYear-1)
			fmt.Fprintf(w,"<input type=\"hidden\" id=\"accMonth\" name=\"accMonth\" value=\"%d\">", 12)
		} else {
			fmt.Fprintf(w,"<input type=\"hidden\" id=\"accYear\" name=\"accYear\" value=\"%d\">", accYear)
			fmt.Fprintf(w,"<input type=\"hidden\" id=\"accMonth\" name=\"accMonth\" value=\"%d\">", accMonth-1)
		}
		fmt.Fprintf(w,"<input type=\"submit\" value=\"Föregående månad\"></form>\n")
	}
	
	fmt.Fprintf(w, "<a href=\"summary\">Översikt</a>\n")
	fmt.Fprintf(w, "</body>\n")
	fmt.Fprintf(w, "</html>\n")
}

func printTransactions(w http.ResponseWriter, db *sql.DB, startDate string, endDate string, limitcomment string) {
	fmt.Println("printTransactions startDate:", startDate)
	fmt.Println("printTransactions endDate:", endDate)
	fmt.Println("printTransactions comment:", limitcomment)
	
	fmt.Fprintf(w, "<h1>%s</h1>\n", currentDatabase)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var err error
	var res *sql.Rows

	if len(limitcomment)>0 {
		res, err = db.QueryContext(ctx,
`SELECT FrånKonto,TillKonto,Typ,Datum,Vad,Vem,Belopp,Löpnr,Saldo,Fastöverföring,Text from transaktioner
  where (datum < ?) and (datum >= ?) and (Text like ?)
order by datum,löpnr`, endDate, startDate, limitcomment)
	} else {
		res, err = db.QueryContext(ctx,
`SELECT FrånKonto,TillKonto,Typ,Datum,Vad,Vem,Belopp,Löpnr,Saldo,Fastöverföring,Text from transaktioner
  where (datum < ?) and (datum >= ?)
order by datum,löpnr`, endDate, startDate)
	}
	if err != nil {
		log.Fatal(err)
		os.Exit(2)
	}

	var fromAcc []byte  // size 40
	var toAcc []byte    // size 40
	var tType []byte    // size 40
	var date []byte     // size 10
	var what []byte     // size 40
	var who []byte      // size 50
	var amount []byte   // BCD / Decimal Precision 19
	var nummer int      // Autoinc Primary Key, index
	var saldo []byte    // BCD / Decimal Precision 19
	var fixed bool      // Boolean
	var comment []byte  // size 60

	fmt.Fprintf(w, "<table style=\"width:100%%\"><tr><th>Löpnr</th><th>Frånkonto</th><th>Tillkonto/Plats</th><th>Typ</th><th>Vad</th><th>Datum</th><th>Vem</th><th>Belopp</th><th>Text</th><th>Fast överföring</th>\n")
	for res.Next() {
		err = res.Scan(&fromAcc, &toAcc, &tType, &date, &what, &who, &amount, &nummer, &saldo, &fixed, &comment)
		
		sqlStmt:=""
		sqlStmt+="<tr><td>" + strconv.Itoa(nummer) + "</td>"
		sqlStmt+="<td>" + toUtf8(fromAcc) + "</td>"
		sqlStmt+="<td>" + toUtf8(toAcc) + "</td>"
		sqlStmt+="<td>" + toUtf8(tType) + "</td>"
		sqlStmt+="<td>" + toUtf8(what) + "</td>"
		sqlStmt+="<td>" + toUtf8(date) + "</td>"
		sqlStmt+="<td>" + toUtf8(who) + "</td>"
		sqlStmt+="<td>" + toUtf8(amount) + "</td>"
		sqlStmt+="<td>" + escapeHTML(toUtf8(comment)) + "</td>\n"
		sqlStmt+="<td>" + strconv.FormatBool(fixed) + "</td></tr>"
		fmt.Fprintf(w, sqlStmt)
	}
	fmt.Fprintf(w, "</table>\n")
}

func transactions(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "<html>\n")
	fmt.Fprintf(w, "<head>\n")
	fmt.Fprintf(w, "<style>\n")
	fmt.Fprintf(w, "table,th,td { border: 1px solid black }\n")
	fmt.Fprintf(w, "</style>\n")
	fmt.Fprintf(w, "</head>\n")
	fmt.Fprintf(w, "<body>\n")

	currentTime := time.Now()
	startDate := currentTime.Format("2006-01-02")
	startDate = startDate[0:8] + "01"
	endDay := currentTime.AddDate(0, 1, 0)
	endDate := endDay.Format("2006-01-02")
	endDate = endDate[0:8] + "01"

	err := req.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	if len(req.FormValue("startdate"))>3 {
		startDate = req.FormValue("startdate")
	}
	if len(req.FormValue("enddate"))>3 {
		endDate = req.FormValue("enddate")
	}

	if(db==nil) {
		fmt.Fprintf(w, "Transactions: No database open<p>\n")
	} else {
		res1 := db.QueryRow("SELECT MIN(Datum) FROM Transaktioner")
		var date []byte     // size 10
		err = res1.Scan(&date)
		res1 = db.QueryRow("SELECT MAX(Datum) FROM Transaktioner")
		err = res1.Scan(&date)

		printTransactions(w, db, startDate, endDate, req.FormValue("comment"))
		fmt.Fprintf(w,"<form method=\"POST\" action=\"/transactions\">\n")
		fmt.Fprintf(w,"<label for=\"startdate\">Startdatum:</label>")
		fmt.Fprintf(w,"	<input type=\"date\" id=\"startdate\" name=\"startdate\" value=\"%s\" title=\"Inklusive\">", startDate)
		fmt.Fprintf(w,"<label for=\"enddate\">Slutdatum:</label>")
		fmt.Fprintf(w,"	<input type=\"date\" id=\"enddate\" name=\"enddate\" value=\"%s\" title=\"Exclusive\">", endDate)
		fmt.Fprintf(w,"<label for=\"comment\">Kommentar:</label>")
		fmt.Fprintf(w,"	<input id=\"comment\" name=\"comment\" value=\"%s\" placeholder=\"wildcards %%_\" title=\"Söktext\n%% är noll, ett eller många tecken.\n_ är ett tecken.\nTomt fält betyder inget filtreras.\">", req.FormValue("comment"))

		fmt.Fprintf(w,"<input type=\"submit\" value=\"Visa\"></form>\n")

		fmt.Fprintf(w,"<form method=\"POST\" action=\"/transactions\">\n")
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

func getAccNames() []string {
	names := make([]string, 0)

	res, err := db.Query("SELECT Benämning FROM Konton ORDER BY Benämning")

	if err != nil {
		log.Fatal(err)
		os.Exit(2)
	}

	var Benämning  []byte  // size 40, index
	for res.Next() {
		err = res.Scan(&Benämning)
		names = append(names, toUtf8(Benämning))
	}
	return names
}

func getPlaceNames() []string {
	names := make([]string, 0)

	res, err := db.Query("SELECT Namn FROM Platser ORDER BY Namn")

	if err != nil {
		log.Fatal(err)
		os.Exit(2)
	}

	var Namn  []byte  // size 40, index
	for res.Next() {
		err = res.Scan(&Namn)
		names = append(names, toUtf8(Namn))
	}
	return names
}

func getPersonNames() []string {
	names := make([]string, 0)

	res, err := db.Query("SELECT Namn FROM Personer ORDER BY Namn")

	if err != nil {
		log.Fatal(err)
		os.Exit(2)
	}

	var Namn  []byte  // size 50, index
	for res.Next() {
		err = res.Scan(&Namn)
		names = append(names, toUtf8(Namn))
	}
	return names
}

func getTypeInNames() []string {
	names := make([]string, 0)

	res, err := db.Query("SELECT Typ FROM Budget where Inkomst = 'J' ORDER BY Typ")

	if err != nil {
		log.Fatal(err)
		os.Exit(2)
	}

	var Typ  []byte  // size 40, index
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

	var Typ  []byte  // size 40, index
	for res.Next() {
		err = res.Scan(&Typ)
		names = append(names, toUtf8(Typ))
	}
	return names
}

func newtransaction(w http.ResponseWriter, req *http.Request) {
	fmt.Println("newtrans 10")
	// Common
	kontonamn := getAccNames()
	fmt.Println("newtrans 20")
	platser := getPlaceNames()
	personer := getPersonNames()
	vad_inkomst := getTypeInNames()
	vad_utgift := getTypeOutNames()
	fmt.Println("newtrans 30")
	
	fmt.Fprintf(w, "<html>\n")
	fmt.Fprintf(w, "<head>\n")
	fmt.Fprintf(w, "<style>\n")
	fmt.Fprintf(w, "table,th,td { border: 1px solid black }\n")
	fmt.Fprintf(w, "</style>\n")
	fmt.Fprintf(w, "</head>\n")
	fmt.Fprintf(w, "<body>\n")
	fmt.Fprintf(w, "<h1>%s</h1>\n", currentDatabase)
	
	// Inköp
	fmt.Fprintf(w, "<h3>Inköp</h3>\n")
	fmt.Fprintf(w,"<form method=\"POST\" action=\"/addtrans\">\n")
	fmt.Fprintf(w,"<input type=\"hidden\" id=\"transtyp\" name=\"transtyp\" value=\"Inköp\">\n")
	fmt.Fprintf(w,"  <label for=\"fromacc\">Från:</label>")
	fmt.Fprintf(w,"  <select name=\"fromacc\" id=\"fromacc\">")
	for _,s := range kontonamn {
		fmt.Fprintf(w,"    <option value=\"%s\">%s</option>", s, s)
	}
	fmt.Println("newtrans 40")
	fmt.Fprintf(w,"  </select>\n")
	fmt.Fprintf(w,"  <label for=\"place\">Plats:</label>")
	fmt.Fprintf(w,"  <select name=\"place\" id=\"place\">")
	for _,s := range platser {
		fmt.Fprintf(w,"    <option value=\"%s\">%s</option>", s, s)
	}
	fmt.Fprintf(w,"  </select>\n")
	fmt.Fprintf(w,"<label for=\"date\">Datum:</label>")
	fmt.Fprintf(w,"	<input type=\"date\" id=\"date\" name=\"date\">")
	fmt.Fprintf(w,"  <label for=\"what\">Vad:</label>")
	fmt.Fprintf(w,"  <select name=\"what\" id=\"what\">")
	for _,s := range vad_utgift {
		fmt.Fprintf(w,"    <option value=\"%s\">%s</option>", s, s)
	}
	fmt.Fprintf(w,"  </select>\n")
	fmt.Fprintf(w,"  <label for=\"who\">Vem:</label>")
	fmt.Fprintf(w,"  <select name=\"who\" id=\"who\">")
	for _,s := range personer {
		fmt.Fprintf(w,"    <option value=\"%s\">%s</option>", s, s)
	}
	fmt.Fprintf(w,"  </select>\n")
	fmt.Fprintf(w,"<label for=\"amount\">Belopp:</label>")
	fmt.Fprintf(w,"<input type=\"number\" id=\"amount\" name=\"amount\" min=0 step=\"0.01\">")
	fmt.Fprintf(w,"<label for=\"text\">Text:</label>")
	fmt.Fprintf(w,"<input type=\"text\" id=\"text\" name=\"text\" >")
	 fmt.Fprintf(w,"<input type=\"submit\" value=\"Submit\"></form>\n")
	// Insättning
	fmt.Fprintf(w, "<h3>Insättning</h3>\n")
	fmt.Fprintf(w,"<form method=\"POST\" action=\"/addtrans\">\n")
	fmt.Fprintf(w,"<input type=\"hidden\" id=\"transtyp\" name=\"transtyp\" value=\"Insättning\">\n")
	fmt.Fprintf(w,"  <label for=\"toacc\">Till:</label>")
	fmt.Fprintf(w,"  <select name=\"toacc\" id=\"fromacc\">")
	for _,s := range kontonamn {
		fmt.Fprintf(w,"    <option value=\"%s\">%s</option>", s, s)
	}
	fmt.Println("newtrans 40")
	fmt.Fprintf(w,"  </select>\n")
	fmt.Fprintf(w,"<label for=\"date\">Datum:</label>")
	fmt.Fprintf(w,"	<input type=\"date\" id=\"date\" name=\"date\">")
	fmt.Fprintf(w,"  <label for=\"what\">Vad:</label>")
	fmt.Fprintf(w,"  <select name=\"what\" id=\"what\">")
	for _,s := range vad_inkomst {
		fmt.Fprintf(w,"    <option value=\"%s\">%s</option>", s, s)
	}
	fmt.Fprintf(w,"  </select>\n")
	fmt.Fprintf(w,"  <label for=\"who\">Vem:</label>")
	fmt.Fprintf(w,"  <select name=\"who\" id=\"who\">")
	for _,s := range personer {
		fmt.Fprintf(w,"    <option value=\"%s\">%s</option>", s, s)
	}
	fmt.Fprintf(w,"  </select>\n")
	fmt.Fprintf(w,"<label for=\"amount\">Belopp:</label>")
	fmt.Fprintf(w,"<input type=\"number\" id=\"amount\" name=\"amount\" min=0 step=\"0.01\">")
	fmt.Fprintf(w,"<label for=\"text\">Text:</label>")
	fmt.Fprintf(w,"<input type=\"text\" id=\"text\" name=\"text\" >")
	fmt.Fprintf(w,"<input type=\"submit\" value=\"Submit\"></form>\n")
	// Uttag
	fmt.Fprintf(w, "<h3>Uttag</h3>\n")
	fmt.Fprintf(w,"<form method=\"POST\" action=\"/addtrans\">\n")
	fmt.Fprintf(w,"<input type=\"hidden\" id=\"transtyp\" name=\"transtyp\" value=\"Uttag\">\n")
	fmt.Fprintf(w,"  <label for=\"fromacc\">Från:</label>")
	fmt.Fprintf(w,"  <select name=\"fromacc\" id=\"fromacc\">")
	for _,s := range kontonamn {
		fmt.Fprintf(w,"    <option value=\"%s\">%s</option>", s, s)
	}
	fmt.Println("newtrans 40")
	fmt.Fprintf(w,"  </select>\n")
	fmt.Fprintf(w,"<label for=\"date\">Datum:</label>")
	fmt.Fprintf(w,"	<input type=\"date\" id=\"date\" name=\"date\">")
	fmt.Fprintf(w,"  <label for=\"who\">Vem:</label>")
	fmt.Fprintf(w,"  <select name=\"who\" id=\"who\">")
	for _,s := range personer {
		fmt.Fprintf(w,"    <option value=\"%s\">%s</option>", s, s)
	}
	fmt.Fprintf(w,"  </select>\n")
	fmt.Fprintf(w,"<label for=\"amount\">Belopp:</label>")
	fmt.Fprintf(w,"<input type=\"number\" id=\"amount\" name=\"amount\" min=0 step=\"0.01\">")
	fmt.Fprintf(w,"<label for=\"text\">Text:</label>")
	fmt.Fprintf(w,"<input type=\"text\" id=\"text\" name=\"text\" >")
	fmt.Fprintf(w,"<input type=\"submit\" value=\"Submit\"></form>\n")
	// Överföring
	fmt.Fprintf(w, "<h3>Överföring</h3>\n")
	fmt.Fprintf(w,"<form method=\"POST\" action=\"/addtrans\">\n")
	fmt.Fprintf(w,"<input type=\"hidden\" id=\"transtyp\" name=\"transtyp\" value=\"Överföring\">\n")
	fmt.Fprintf(w,"  <label for=\"fromacc\">Från:</label>")
	fmt.Fprintf(w,"  <select name=\"fromacc\" id=\"fromacc\">")
	for _,s := range kontonamn {
		fmt.Fprintf(w,"    <option value=\"%s\">%s</option>", s, s)
	}
	fmt.Fprintf(w,"  </select>\n")
	fmt.Fprintf(w,"  <label for=\"toacc\">Till:</label>")
	fmt.Fprintf(w,"  <select name=\"toacc\" id=\"toacc\">")
	for _,s := range kontonamn {
		fmt.Fprintf(w,"    <option value=\"%s\">%s</option>", s, s)
	}
	fmt.Println("newtrans 40")
	fmt.Fprintf(w,"  </select>\n")
	fmt.Fprintf(w,"<label for=\"date\">Datum:</label>")
	fmt.Fprintf(w,"	<input type=\"date\" id=\"date\" name=\"date\">")
	fmt.Fprintf(w,"  <label for=\"who\">Vem:</label>")
	fmt.Fprintf(w,"  <select name=\"who\" id=\"who\">")
	for _,s := range personer {
		fmt.Fprintf(w,"    <option value=\"%s\">%s</option>", s, s)
	}
	fmt.Fprintf(w,"  </select>\n")
	fmt.Fprintf(w,"<label for=\"amount\">Belopp:</label>")
	fmt.Fprintf(w,"<input type=\"number\" id=\"amount\" name=\"amount\" min=0 step=\"0.01\">")
	fmt.Fprintf(w,"<label for=\"text\">Text:</label>")
	fmt.Fprintf(w,"<input type=\"text\" id=\"text\" name=\"text\" >")
	fmt.Fprintf(w,"<input type=\"submit\" value=\"Submit\"></form>\n")
	fmt.Println("newtrans 50")

	fmt.Fprintf(w, "<a href=\"summary\">Översikt</a>\n")
	fmt.Fprintf(w, "</body>\n")
	fmt.Fprintf(w, "</html>\n")
}

func addtransaction(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "<html>\n")
	fmt.Fprintf(w, "<head>\n")
	fmt.Fprintf(w, "<style>\n")
	fmt.Fprintf(w, "table,th,td { border: 1px solid black }\n")
	fmt.Fprintf(w, "</style>\n")
	fmt.Fprintf(w, "</head>\n")
	fmt.Fprintf(w, "<body>\n")
	fmt.Fprintf(w, "<h1>%s</h1>\n", currentDatabase)

	err := req.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	transtyp := req.FormValue("transtyp")
	date := req.FormValue("date")
	who := req.FormValue("who")
	amount := req.FormValue("amount")
	text := req.FormValue("text")
	fmt.Println("Val: ", transtyp)
	fmt.Println("Val: ", date)
	fmt.Println("Val: ", who)
	fmt.Println("Val: ", amount)
	fmt.Println("Val: ", text)

	if transtyp == "Inköp" {
		fromacc := req.FormValue("fromacc")
		place := req.FormValue("place")
		what := req.FormValue("what")
		fmt.Println("Val: ", fromacc)
		fmt.Println("Val: ", place)
		fmt.Println("Val: ", what)

		fmt.Fprintf(w, "Inserting... ")

		sqlStatement := `
INSERT INTO Transaktioner (FrånKonto,TillKonto,Typ,Datum,Vad,Vem,Belopp,"Text")
VALUES (?,?,?,?,?,?,?,?)`
		_, err = db.Exec(sqlStatement, fromacc, place, transtyp, date, what, who, strings.ReplaceAll(amount, ".", ","), text)
		if err != nil {
			panic(err)
		}

		fmt.Fprintf(w, "Inserted.\n")
		
		fmt.Fprintf(w," Insert res:\n", err)
	}
	if transtyp == "Insättning" {
		toacc := req.FormValue("toacc")
		what := req.FormValue("what")
		fmt.Println("Val: ", toacc)
		fmt.Println("Val: ", what)

		fmt.Fprintf(w, "Inserting... ")

		sqlStatement := `
INSERT INTO Transaktioner (FrånKonto,TillKonto,Typ,Datum,Vad,Vem,Belopp,"Text")
VALUES (?,?,?,?,?,?,?,?)`
		_, err = db.Exec(sqlStatement, "---", toacc, transtyp, date, what, who, strings.ReplaceAll(amount, ".", ","), text)
		if err != nil {
			panic(err)
		}

		fmt.Fprintf(w, "Inserted.\n")
		
		fmt.Fprintf(w," Insert res:\n", err)
	}
	if transtyp == "Uttag" {
		fromacc := req.FormValue("fromacc")
		what := req.FormValue("what")
		fmt.Println("Val: ", fromacc)
		fmt.Println("Val: ", what)

		fmt.Fprintf(w, "Inserting... ")

		sqlStatement := `
INSERT INTO Transaktioner (FrånKonto,TillKonto,Typ,Datum,Vad,Vem,Belopp,"Text")
VALUES (?,?,?,?,?,?,?,?)`
		_, err = db.Exec(sqlStatement, fromacc, "Plånboken", transtyp, date, "---", who, strings.ReplaceAll(amount, ".", ","), text)
		if err != nil {
			panic(err)
		}
		fmt.Fprintf(w, "Inserted.\n")
		
		fmt.Fprintf(w," Insert res:\n", err)
	}
	if transtyp == "Överföring" {
		fromacc := req.FormValue("fromacc")
		toacc := req.FormValue("toacc")
		fmt.Println("Val: ", fromacc)
		fmt.Println("Val: ", toacc)

		fmt.Fprintf(w, "Inserting... ")

		sqlStatement := `
INSERT INTO Transaktioner (FrånKonto,TillKonto,Typ,Datum,Vad,Vem,Belopp,"Text")
VALUES (?,?,?,?,?,?,?,?)`
		_, err = db.Exec(sqlStatement, fromacc, toacc, transtyp, date, "---", who, strings.ReplaceAll(amount, ".", ","), text)
		if err != nil {
			panic(err)
		}
		fmt.Fprintf(w, "Inserted.\n")
		
		fmt.Fprintf(w," Insert res:\n", err)
	}

	fmt.Fprintf(w, "<a href=\"summary\">Översikt</a>\n")
	fmt.Fprintf(w, "</body>\n")
	fmt.Fprintf(w, "</html>\n")
}

func main() {
	http.HandleFunc("/hello", hello)
	http.HandleFunc("/r", restapi)
	http.HandleFunc("/headers", headers)
	http.HandleFunc("/open", opendb)
	http.HandleFunc("/close", closedb)
	http.HandleFunc("/newtrans", newtransaction)
	http.HandleFunc("/addtrans", addtransaction)
	http.HandleFunc("/monthly", monthly)
	http.HandleFunc("/transactions", transactions)
	http.HandleFunc("/summary", generateSummary)
	http.HandleFunc("/", root)

	ip, _ := externalIP()
	fmt.Println("Öppna URL i webläsaren:  http://localhost:8090/")
	fmt.Printf(" eller :  http://%s:8090/\n", ip)
	http.ListenAndServe(":8090", nil)
}
