//-*- coding: utf-8 -*-

package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"html"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/alexbrainman/odbc" // BSD-3-Clause License
	_ "github.com/mattn/go-sqlite3"  // MIT License
)

func printTransactions(w http.ResponseWriter, db *sql.DB, startDate string, endDate string, limitcomment string) {
	fmt.Println("printTransactions startDate:", startDate)
	fmt.Println("printTransactions endDate:", endDate)
	fmt.Println("printTransactions comment:", limitcomment)

	fmt.Fprintf(w, "<h1>%s</h1>\n", currentDatabase)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var err error
	var res *sql.Rows

	if len(limitcomment) > 0 {
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

	fmt.Fprintf(w, "<table style=\"width:100%%\"><tr><th>Löpnr</th><th>Frånkonto</th><th>Tillkonto/Plats</th><th>Typ</th><th>Vad</th><th>Datum</th><th>Vem</th><th>Belopp</th><th>Text</th><th>Fast överföring</th>\n")
	for res.Next() {
		err = res.Scan(&fromAcc, &toAcc, &tType, &date, &what, &who, &amount, &nummer, &saldo, &fixed, &comment)

		sqlStmt := ""
		sqlStmt += "<tr><td>" + strconv.Itoa(nummer) + "</td>"
		sqlStmt += "<td>" + toUtf8(fromAcc) + "</td>"
		sqlStmt += "<td>" + toUtf8(toAcc) + "</td>"
		sqlStmt += "<td>" + toUtf8(tType) + "</td>"
		sqlStmt += "<td>" + toUtf8(what) + "</td>"
		sqlStmt += "<td>" + toUtf8(date) + "</td>"
		sqlStmt += "<td>" + toUtf8(who) + "</td>"
		sqlStmt += "<td>" + toUtf8(amount) + "</td>"
		sqlStmt += "<td>" + html.EscapeString(toUtf8(comment)) + "</td>\n"
		sqlStmt += "<td>" + strconv.FormatBool(fixed) + "</td></tr>"
		fmt.Fprintf(w, "%s", sqlStmt)
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

	if len(req.FormValue("startdate")) > 3 {
		startDate = req.FormValue("startdate")
	}
	if len(req.FormValue("enddate")) > 3 {
		endDate = req.FormValue("enddate")
	}

	if db == nil {
		fmt.Fprintf(w, "Transactions: No database open<p>\n")
	} else {
		res1 := db.QueryRow("SELECT MIN(Datum) FROM Transaktioner")
		var date []byte // size 10
		err = res1.Scan(&date)
		res1 = db.QueryRow("SELECT MAX(Datum) FROM Transaktioner")
		err = res1.Scan(&date)

		printTransactions(w, db, startDate, endDate, req.FormValue("comment"))
		fmt.Fprintf(w, "<form method=\"POST\" action=\"/transactions\">\n")
		fmt.Fprintf(w, "<label for=\"startdate\">Startdatum:</label>")
		fmt.Fprintf(w, "	<input type=\"date\" id=\"startdate\" name=\"startdate\" value=\"%s\" title=\"Inklusive\">", startDate)
		fmt.Fprintf(w, "<label for=\"enddate\">Slutdatum:</label>")
		fmt.Fprintf(w, "	<input type=\"date\" id=\"enddate\" name=\"enddate\" value=\"%s\" title=\"Exclusive\">", endDate)
		fmt.Fprintf(w, "<label for=\"comment\">Kommentar:</label>")
		fmt.Fprintf(w, "	<input id=\"comment\" name=\"comment\" value=\"%s\" placeholder=\"wildcards %%_\" title=\"Söktext\n%% är noll, ett eller många tecken.\n_ är ett tecken.\nTomt fält betyder inget filtreras.\">", req.FormValue("comment"))

		fmt.Fprintf(w, "<input type=\"submit\" value=\"Visa\"></form>\n")

		fmt.Fprintf(w, "<form method=\"POST\" action=\"/transactions\">\n")
	}

	fmt.Fprintf(w, "<a href=\"summary\">Översikt</a>\n")
	fmt.Fprintf(w, "</body>\n")
	fmt.Fprintf(w, "</html>\n")
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
	fmt.Fprintf(w, "<form method=\"POST\" action=\"/addtrans\">\n")
	fmt.Fprintf(w, "<input type=\"hidden\" id=\"transtyp\" name=\"transtyp\" value=\"Inköp\">\n")
	fmt.Fprintf(w, "  <label for=\"fromacc\">Från:</label>")
	fmt.Fprintf(w, "  <select name=\"fromacc\" id=\"fromacc\">")
	for _, s := range kontonamn {
		fmt.Fprintf(w, "    <option value=\"%s\">%s</option>", s, s)
	}
	fmt.Println("newtrans 40")
	fmt.Fprintf(w, "  </select>\n")
	fmt.Fprintf(w, "  <label for=\"place\">Plats:</label>")
	fmt.Fprintf(w, "  <select name=\"place\" id=\"place\">")
	for _, s := range platser {
		fmt.Fprintf(w, "    <option value=\"%s\">%s</option>", s, s)
	}
	fmt.Fprintf(w, "  </select>\n")
	fmt.Fprintf(w, "<label for=\"date\">Datum:</label>")
	fmt.Fprintf(w, "	<input type=\"date\" id=\"date\" name=\"date\">")
	fmt.Fprintf(w, "  <label for=\"what\">Vad:</label>")
	fmt.Fprintf(w, "  <select name=\"what\" id=\"what\">")
	for _, s := range vad_utgift {
		fmt.Fprintf(w, "    <option value=\"%s\">%s</option>", s, s)
	}
	fmt.Fprintf(w, "  </select>\n")
	fmt.Fprintf(w, "  <label for=\"who\">Vem:</label>")
	fmt.Fprintf(w, "  <select name=\"who\" id=\"who\">")
	for _, s := range personer {
		fmt.Fprintf(w, "    <option value=\"%s\">%s</option>", s, s)
	}
	fmt.Fprintf(w, "  </select>\n")
	fmt.Fprintf(w, "<label for=\"amount\">Belopp:</label>")
	fmt.Fprintf(w, "<input type=\"number\" id=\"amount\" name=\"amount\" min=0 step=\"0.01\">")
	fmt.Fprintf(w, "<label for=\"text\">Text:</label>")
	fmt.Fprintf(w, "<input type=\"text\" id=\"text\" name=\"text\" >")
	fmt.Fprintf(w, "<input type=\"submit\" value=\"Submit\"></form>\n")
	// Insättning
	fmt.Fprintf(w, "<h3>Insättning</h3>\n")
	fmt.Fprintf(w, "<form method=\"POST\" action=\"/addtrans\">\n")
	fmt.Fprintf(w, "<input type=\"hidden\" id=\"transtyp\" name=\"transtyp\" value=\"Insättning\">\n")
	fmt.Fprintf(w, "  <label for=\"toacc\">Till:</label>")
	fmt.Fprintf(w, "  <select name=\"toacc\" id=\"fromacc\">")
	for _, s := range kontonamn {
		fmt.Fprintf(w, "    <option value=\"%s\">%s</option>", s, s)
	}
	fmt.Println("newtrans 40")
	fmt.Fprintf(w, "  </select>\n")
	fmt.Fprintf(w, "<label for=\"date\">Datum:</label>")
	fmt.Fprintf(w, "	<input type=\"date\" id=\"date\" name=\"date\">")
	fmt.Fprintf(w, "  <label for=\"what\">Vad:</label>")
	fmt.Fprintf(w, "  <select name=\"what\" id=\"what\">")
	for _, s := range vad_inkomst {
		fmt.Fprintf(w, "    <option value=\"%s\">%s</option>", s, s)
	}
	fmt.Fprintf(w, "  </select>\n")
	fmt.Fprintf(w, "  <label for=\"who\">Vem:</label>")
	fmt.Fprintf(w, "  <select name=\"who\" id=\"who\">")
	for _, s := range personer {
		fmt.Fprintf(w, "    <option value=\"%s\">%s</option>", s, s)
	}
	fmt.Fprintf(w, "  </select>\n")
	fmt.Fprintf(w, "<label for=\"amount\">Belopp:</label>")
	fmt.Fprintf(w, "<input type=\"number\" id=\"amount\" name=\"amount\" min=0 step=\"0.01\">")
	fmt.Fprintf(w, "<label for=\"text\">Text:</label>")
	fmt.Fprintf(w, "<input type=\"text\" id=\"text\" name=\"text\" >")
	fmt.Fprintf(w, "<input type=\"submit\" value=\"Submit\"></form>\n")
	// Uttag
	fmt.Fprintf(w, "<h3>Uttag</h3>\n")
	fmt.Fprintf(w, "<form method=\"POST\" action=\"/addtrans\">\n")
	fmt.Fprintf(w, "<input type=\"hidden\" id=\"transtyp\" name=\"transtyp\" value=\"Uttag\">\n")
	fmt.Fprintf(w, "  <label for=\"fromacc\">Från:</label>")
	fmt.Fprintf(w, "  <select name=\"fromacc\" id=\"fromacc\">")
	for _, s := range kontonamn {
		fmt.Fprintf(w, "    <option value=\"%s\">%s</option>", s, s)
	}
	fmt.Println("newtrans 40")
	fmt.Fprintf(w, "  </select>\n")
	fmt.Fprintf(w, "<label for=\"date\">Datum:</label>")
	fmt.Fprintf(w, "	<input type=\"date\" id=\"date\" name=\"date\">")
	fmt.Fprintf(w, "  <label for=\"who\">Vem:</label>")
	fmt.Fprintf(w, "  <select name=\"who\" id=\"who\">")
	for _, s := range personer {
		fmt.Fprintf(w, "    <option value=\"%s\">%s</option>", s, s)
	}
	fmt.Fprintf(w, "  </select>\n")
	fmt.Fprintf(w, "<label for=\"amount\">Belopp:</label>")
	fmt.Fprintf(w, "<input type=\"number\" id=\"amount\" name=\"amount\" min=0 step=\"0.01\">")
	fmt.Fprintf(w, "<label for=\"text\">Text:</label>")
	fmt.Fprintf(w, "<input type=\"text\" id=\"text\" name=\"text\" >")
	fmt.Fprintf(w, "<input type=\"submit\" value=\"Submit\"></form>\n")
	// Överföring
	fmt.Fprintf(w, "<h3>Överföring</h3>\n")
	fmt.Fprintf(w, "<form method=\"POST\" action=\"/addtrans\">\n")
	fmt.Fprintf(w, "<input type=\"hidden\" id=\"transtyp\" name=\"transtyp\" value=\"Överföring\">\n")
	fmt.Fprintf(w, "  <label for=\"fromacc\">Från:</label>")
	fmt.Fprintf(w, "  <select name=\"fromacc\" id=\"fromacc\">")
	for _, s := range kontonamn {
		fmt.Fprintf(w, "    <option value=\"%s\">%s</option>", s, s)
	}
	fmt.Fprintf(w, "  </select>\n")
	fmt.Fprintf(w, "  <label for=\"toacc\">Till:</label>")
	fmt.Fprintf(w, "  <select name=\"toacc\" id=\"toacc\">")
	for _, s := range kontonamn {
		fmt.Fprintf(w, "    <option value=\"%s\">%s</option>", s, s)
	}
	fmt.Println("newtrans 40")
	fmt.Fprintf(w, "  </select>\n")
	fmt.Fprintf(w, "<label for=\"date\">Datum:</label>")
	fmt.Fprintf(w, "	<input type=\"date\" id=\"date\" name=\"date\">")
	fmt.Fprintf(w, "  <label for=\"who\">Vem:</label>")
	fmt.Fprintf(w, "  <select name=\"who\" id=\"who\">")
	for _, s := range personer {
		fmt.Fprintf(w, "    <option value=\"%s\">%s</option>", s, s)
	}
	fmt.Fprintf(w, "  </select>\n")
	fmt.Fprintf(w, "<label for=\"amount\">Belopp:</label>")
	fmt.Fprintf(w, "<input type=\"number\" id=\"amount\" name=\"amount\" min=0 step=\"0.01\">")
	fmt.Fprintf(w, "<label for=\"text\">Text:</label>")
	fmt.Fprintf(w, "<input type=\"text\" id=\"text\" name=\"text\" >")
	fmt.Fprintf(w, "<input type=\"submit\" value=\"Submit\"></form>\n")
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

		fmt.Fprintf(w, " Insert res:\n", err)
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

		fmt.Fprintf(w, " Insert res:\n", err)
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

		fmt.Fprintf(w, " Insert res:\n", err)
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

		fmt.Fprintf(w, " Insert res:\n", err)
	}

	fmt.Fprintf(w, "<a href=\"summary\">Översikt</a>\n")
	fmt.Fprintf(w, "</body>\n")
	fmt.Fprintf(w, "</html>\n")
}
