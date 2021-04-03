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
	fmt.Fprintf(w, "<th>Redigera</th><th>Radera</th>\n")
	for res.Next() {
		err = res.Scan(&fromAcc, &toAcc, &tType, &date, &what, &who, &amount, &nummer, &saldo, &fixed, &comment)

		fmt.Fprintf(w, "<tr><td>" + strconv.Itoa(nummer) + "</td>")
		fmt.Fprintf(w, "<td>" + toUtf8(fromAcc) + "</td>")
		fmt.Fprintf(w, "<td>" + toUtf8(toAcc) + "</td>")
		fmt.Fprintf(w, "<td>" + toUtf8(tType) + "</td>")
		fmt.Fprintf(w, "<td>" + toUtf8(what) + "</td>")
		fmt.Fprintf(w, "<td>" + toUtf8(date) + "</td>")
		fmt.Fprintf(w, "<td>" + toUtf8(who) + "</td>")
		fmt.Fprintf(w, "<td>" + toUtf8(amount) + "</td>")
		fmt.Fprintf(w, "<td>" + html.EscapeString(toUtf8(comment)) + "</td>\n")
		fmt.Fprintf(w, "<td>" + strconv.FormatBool(fixed) + "</td>")
		fmt.Fprintf(w, "<td><form method=\"POST\" action=\"/transactions\"><input type=\"hidden\" id=\"lopnr\" name=\"lopnr\" value=\"%d\"><input type=\"hidden\" id=\"action\" name=\"action\" value=\"editform\"><input type=\"submit\" value=\"Redigera\"></form></td>\n", nummer)
		fmt.Fprintf(w, "<td><form method=\"POST\" action=\"/transactions\"><input type=\"hidden\" id=\"lopnr\" name=\"lopnr\" value=\"%d\"><input type=\"hidden\" id=\"action\" name=\"action\" value=\"radera\"><input type=\"checkbox\" id=\"OK\" name=\"OK\" required><label for=\"OK\">OK</label><input type=\"submit\" value=\"Radera\"></form></td></tr>\n", nummer)
	}
	fmt.Fprintf(w, "</table>\n")
}

func handletransactions(w http.ResponseWriter, req *http.Request) {
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
}

func transactions(w http.ResponseWriter, req *http.Request) {
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

	formaction := req.FormValue("action")
	var lopnr int = -1
	if len(req.FormValue("lopnr")) > 0 {
		lopnr, err = strconv.Atoi(req.FormValue("lopnr"))
	}

	switch formaction {
	case "radera":
		raderaTransaction(w, lopnr, db)
	case "editform":
		editformTransaction(w, lopnr, db)
	case "update":
		updateTransaction(w, lopnr, req, db)
	default:
		fmt.Println("Okänd action: ", formaction)
	}

	handletransactions(w, req)

	fmt.Fprintf(w, "<a href=\"summary\">Översikt</a>\n")
	fmt.Fprintf(w, "</body>\n")
	fmt.Fprintf(w, "</html>\n")
}

func raderaTransaction(w http.ResponseWriter, lopnr int, db *sql.DB) {
	fmt.Println("raderaTransaction lopnr: ", lopnr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := db.ExecContext(ctx,
		`DELETE FROM transaktioner WHERE (Löpnr=?)`, lopnr)

	if err != nil {
		log.Fatal(err)
		os.Exit(2)
	}
	fmt.Fprintf(w, "Transaktion med löpnr %d raderad.<br>", lopnr)
}


func newtransaction(w http.ResponseWriter, req *http.Request) {
	// Common
	kontonamn := getAccNames()

	platser := getPlaceNames()
	personer := getPersonNames()
	vad_inkomst := getTypeInNames()
	vad_utgift := getTypeOutNames()

	fmt.Fprintf(w, "<html>\n")
	fmt.Fprintf(w, "<head>\n")
	fmt.Fprintf(w, "<style>\n")
	fmt.Fprintf(w, "table,th,td { border: 1px solid black }\n")
	fmt.Fprintf(w, "</style>\n")
	fmt.Fprintf(w, "</head>\n")
	fmt.Fprintf(w, "<body>\n")
	fmt.Fprintf(w, "<h1>%s</h1>\n", currentDatabase)

	addtransaction(w, req)
	showFastaTransaktioner(w, req)
	
	// Inköp
	fmt.Fprintf(w, "<h3>Inköp</h3>\n")
	fmt.Fprintf(w, "<form method=\"POST\" action=\"/newtrans\">\n")
	fmt.Fprintf(w, "<input type=\"hidden\" id=\"transtyp\" name=\"transtyp\" value=\"Inköp\">\n")
	fmt.Fprintf(w, "  <label for=\"fromacc\">Från:</label>")
	fmt.Fprintf(w, "  <select name=\"fromacc\" id=\"fromacc\">")
	for _, s := range kontonamn {
		fmt.Fprintf(w, "    <option value=\"%s\">%s</option>", s, s)
	}

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
	fmt.Fprintf(w, "<form method=\"POST\" action=\"/newtrans\">\n")
	fmt.Fprintf(w, "<input type=\"hidden\" id=\"transtyp\" name=\"transtyp\" value=\"Insättning\">\n")
	fmt.Fprintf(w, "  <label for=\"toacc\">Till:</label>")
	fmt.Fprintf(w, "  <select name=\"toacc\" id=\"fromacc\">")
	for _, s := range kontonamn {
		fmt.Fprintf(w, "    <option value=\"%s\">%s</option>", s, s)
	}

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
	fmt.Fprintf(w, "<form method=\"POST\" action=\"/newtrans\">\n")
	fmt.Fprintf(w, "<input type=\"hidden\" id=\"transtyp\" name=\"transtyp\" value=\"Uttag\">\n")
	fmt.Fprintf(w, "  <label for=\"fromacc\">Från:</label>")
	fmt.Fprintf(w, "  <select name=\"fromacc\" id=\"fromacc\">")
	for _, s := range kontonamn {
		fmt.Fprintf(w, "    <option value=\"%s\">%s</option>", s, s)
	}

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
	fmt.Fprintf(w, "<form method=\"POST\" action=\"/newtrans\">\n")
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

	fmt.Fprintf(w, "<a href=\"summary\">Översikt</a>\n")
	fmt.Fprintf(w, "</body>\n")
	fmt.Fprintf(w, "</html>\n")
}

func addtransaction(w http.ResponseWriter, req *http.Request) {
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

		fmt.Fprintf(w, "Registrerar Inköp...<br> ")

		sqlStatement := `
INSERT INTO Transaktioner (FrånKonto,TillKonto,Typ,Datum,Vad,Vem,Belopp,"Text")
VALUES (?,?,?,?,?,?,?,?)`
		_, err = db.Exec(sqlStatement, fromacc, place, transtyp, date, what, who, strings.ReplaceAll(amount, ".", ","), text)
		if err != nil {
			panic(err)
		}

		fmt.Fprintf(w, "<table style=\"width:100%%\"><tr><th>Frånkonto</th><th>Plats</th><th>Typ</th><th>Vad</th><th>Datum</th><th>Vem</th><th>Belopp</th><th>Text</th>\n")
		sqlStmt := "<tr>"
		sqlStmt += "<td>" + fromacc + "</td>"
		sqlStmt += "<td>" + place + "</td>"
		sqlStmt += "<td>" + transtyp + "</td>"
		sqlStmt += "<td>" + what + "</td>"
		sqlStmt += "<td>" + date + "</td>"
		sqlStmt += "<td>" + who + "</td>"
		sqlStmt += "<td>" + amount + "</td>"
		sqlStmt += "<td>" + html.EscapeString(text) + "</td>\n"
		sqlStmt += "</tr>"
		fmt.Fprintf(w, "%s", sqlStmt)
		fmt.Fprintf(w, "</table>\n")
	}
	if transtyp == "Insättning" {
		toacc := req.FormValue("toacc")
		what := req.FormValue("what")
		fmt.Println("Val: ", toacc)
		fmt.Println("Val: ", what)

		fmt.Fprintf(w, "Registrerar Insättning...<br> ")

		sqlStatement := `
INSERT INTO Transaktioner (FrånKonto,TillKonto,Typ,Datum,Vad,Vem,Belopp,"Text")
VALUES (?,?,?,?,?,?,?,?)`
		_, err = db.Exec(sqlStatement, "---", toacc, transtyp, date, what, who, strings.ReplaceAll(amount, ".", ","), text)
		if err != nil {
			panic(err)
		}

		fmt.Fprintf(w, "<table style=\"width:100%%\"><tr><th>Konto</th><th>Typ</th><th>Vad</th><th>Datum</th><th>Vem</th><th>Belopp</th><th>Text</th>\n")
		sqlStmt := "<tr>"
		sqlStmt += "<td>" + toacc + "</td>"
		sqlStmt += "<td>" + transtyp + "</td>"
		sqlStmt += "<td>" + what + "</td>"
		sqlStmt += "<td>" + date + "</td>"
		sqlStmt += "<td>" + who + "</td>"
		sqlStmt += "<td>" + amount + "</td>"
		sqlStmt += "<td>" + html.EscapeString(text) + "</td>\n"
		sqlStmt += "</tr>"
		fmt.Fprintf(w, "%s", sqlStmt)
		fmt.Fprintf(w, "</table>\n")
	}
	if transtyp == "Uttag" {
		fromacc := req.FormValue("fromacc")
		what := req.FormValue("what")
		fmt.Println("Val: ", fromacc)
		fmt.Println("Val: ", what)

		fmt.Fprintf(w, "Registrerar Uttag...<br> ")

		sqlStatement := `
INSERT INTO Transaktioner (FrånKonto,TillKonto,Typ,Datum,Vad,Vem,Belopp,"Text")
VALUES (?,?,?,?,?,?,?,?)`
		_, err = db.Exec(sqlStatement, fromacc, "Plånboken", transtyp, date, "---", who, strings.ReplaceAll(amount, ".", ","), text)
		if err != nil {
			panic(err)
		}

		fmt.Fprintf(w, "<table style=\"width:100%%\"><tr><th>Frånkonto</th><th>Typ</th><th>Datum</th><th>Vem</th><th>Belopp</th><th>Text</th>\n")
		sqlStmt := "<tr>"
		sqlStmt += "<td>" + fromacc + "</td>"
		sqlStmt += "<td>" + transtyp + "</td>"
		sqlStmt += "<td>" + date + "</td>"
		sqlStmt += "<td>" + who + "</td>"
		sqlStmt += "<td>" + amount + "</td>"
		sqlStmt += "<td>" + html.EscapeString(text) + "</td>\n"
		sqlStmt += "</tr>"
		fmt.Fprintf(w, "%s", sqlStmt)
		fmt.Fprintf(w, "</table>\n")
	}
	if transtyp == "Överföring" {
		fromacc := req.FormValue("fromacc")
		toacc := req.FormValue("toacc")
		fmt.Println("Val: ", fromacc)
		fmt.Println("Val: ", toacc)

		fmt.Fprintf(w, "Registrerar Överföring...<br> ")

		sqlStatement := `
INSERT INTO Transaktioner (FrånKonto,TillKonto,Typ,Datum,Vad,Vem,Belopp,"Text")
VALUES (?,?,?,?,?,?,?,?)`
		_, err = db.Exec(sqlStatement, fromacc, toacc, transtyp, date, "---", who, strings.ReplaceAll(amount, ".", ","), text)
		if err != nil {
			panic(err)
		}

		fmt.Fprintf(w, "<table style=\"width:100%%\"><tr><th>Frånkonto</th><th>Tillkonto</th><th>Typ</th><th>Datum</th><th>Vem</th><th>Belopp</th><th>Text</th>\n")
		sqlStmt := "<tr>"
		sqlStmt += "<td>" + fromacc + "</td>"
		sqlStmt += "<td>" + toacc + "</td>"
		sqlStmt += "<td>" + transtyp + "</td>"
		sqlStmt += "<td>" + date + "</td>"
		sqlStmt += "<td>" + who + "</td>"
		sqlStmt += "<td>" + amount + "</td>"
		sqlStmt += "<td>" + html.EscapeString(text) + "</td>\n"
		sqlStmt += "</tr>"
		fmt.Fprintf(w, "%s", sqlStmt)
		fmt.Fprintf(w, "</table>\n")
	}
	if transtyp == "FastTrans" {
		transid := req.FormValue("transid")
		transidnum, _ := strconv.Atoi(transid)
		registreraFastTransaktion(w, transidnum)
		fmt.Fprintf(w, "<p>\n")
	}
}

func editformTransaction(w http.ResponseWriter, lopnr int, db *sql.DB) {
	fmt.Println("editformTransaktion lopnr: ", lopnr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	res1 := db.QueryRowContext(ctx,
		`SELECT FrånKonto,TillKonto,Typ,Datum,Vad,Vem,Belopp,Saldo,Fastöverföring,Text FROM transaktioner WHERE (Löpnr=?)`, lopnr)

	var fromAcc []byte // size 40
	var toAcc []byte   // size 40
	var tType []byte   // size 40
	var date []byte    // size 10
	var what []byte    // size 40
	var who []byte     // size 50
	var amount []byte  // BCD / Decimal Precision 19
	var saldo []byte   // BCD / Decimal Precision 19
	var fixed bool     // Boolean
	var comment []byte // size 60

	err := res1.Scan(&fromAcc, &toAcc, &tType, &date, &what, &who, &amount, &saldo, &fixed, &comment)
	if err != nil {
		log.Fatal(err)
		os.Exit(2)
	}

	fmt.Fprintf(w, "Redigera transaktion<br>")
	fmt.Fprintf(w, "<form method=\"POST\" action=\"/transactions\">")

	PrintEditCellText(w, "fromAcc", "Från konto", toUtf8(fromAcc))
	PrintEditCellText(w, "toAcc", "Till konto", toUtf8(toAcc))
	PrintEditCellText(w, "tType", "Typ", toUtf8(tType))
	PrintEditCellText(w, "date", "Datum", toUtf8(date))
	PrintEditCellText(w, "what", "Vad", toUtf8(what))
	PrintEditCellText(w, "who", "Vem", toUtf8(who))
	PrintEditCellText(w, "amount", "Summa", toUtf8(amount))
	PrintEditCellText(w, "fixed", "Fast transaktion", strconv.FormatBool(fixed))
	PrintEditCellText(w, "comment", "Text", toUtf8(comment))

	fmt.Fprintf(w, "<input type=\"hidden\" id=\"lopnr\" name=\"lopnr\" value=\"%d\">", lopnr)
	fmt.Fprintf(w, "<input type=\"hidden\" id=\"action\" name=\"action\" value=\"update\">")
	fmt.Fprintf(w, "<input type=\"submit\" value=\"Uppdatera\">")
	fmt.Fprintf(w, "</form>\n")
	fmt.Fprintf(w, "<p>\n")
}


func updateTransaction(w http.ResponseWriter, lopnr int, req *http.Request, db *sql.DB) {
	fmt.Println("updateTransaktion lopnr: ", lopnr)

	var fromAcc string = ""
	if len(req.FormValue("fromAcc")) > 0 {
		fromAcc = req.FormValue("fromAcc")
	}
	var toAcc string = ""
	if len(req.FormValue("toAcc")) > 0 {
		toAcc = req.FormValue("toAcc")
	}
	var tType string = ""
	if len(req.FormValue("tType")) > 0 {
		tType = req.FormValue("tType")
	}
	var date string = ""
	if len(req.FormValue("date")) > 0 {
		date = req.FormValue("date")
	}
	var what string = ""
	if len(req.FormValue("what")) > 0 {
		what = req.FormValue("what")
	}
	var who string = ""
	if len(req.FormValue("who")) > 0 {
		who = req.FormValue("who")
	}
	var amount string = ""
	if len(req.FormValue("amount")) > 0 {
		amount = req.FormValue("amount")
	}
	var fixed bool = false
	if len(req.FormValue("fixed")) > 0 {
		var fixedString string = ""
		fixedString = req.FormValue("fixed")
		fixed,_ = strconv.ParseBool(fixedString)
	}
	
	var comment string = ""
	if len(req.FormValue("comment")) > 0 {
		comment = req.FormValue("comment")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := db.ExecContext(ctx,
		`UPDATE transaktioner SET FrånKonto = ?, TillKonto = ?, Typ = ?, Datum = ?, Vad = ?, Vem = ?, Belopp = ?, Fastöverföring = ?, "Text" = ? WHERE (Löpnr=?)`,
		fromAcc,
		toAcc,
		tType,
		date,
		what,
		who,
		strings.ReplaceAll(amount, ".", ","),
		fixed,
		comment,
		lopnr)

	if err != nil {
		log.Fatal(err)
		os.Exit(2)
	}
	
	fmt.Fprintf(w, "Transaktion %d uppdaterad.<br>", lopnr)
}

func CurrDate() string {
	now := time.Now()
	currDate := now.Format("2006-01-02")
	return currDate
}

func IncrDate(datum string, veckor int, månader int) string {
	fmt.Println("IncrDate datum:", datum)
	fmt.Println("IncrDate veckor:", veckor)
	fmt.Println("IncrDate månader:", månader)

	year, _ := strconv.Atoi(datum[0:4])
	var month time.Month
	monthval, _ := strconv.Atoi(datum[5:7])
	switch monthval {
	case 1: month = time.January
	case 2: month = time.February
	case 3: month = time.March
	case 4: month = time.April
	case 5: month = time.May
	case 6: month = time.June
	case 7: month = time.July
	case 8: month = time.August
	case 9: month = time.September
	case 10: month = time.October
	case 11: month = time.November
	case 12: month = time.December
	}
	day, _ := strconv.Atoi(datum[8:10])
	t := time.Date(year, month, day, 12, 0, 0, 0, time.Local) // Note: should be CET
	fmt.Println("IncrDate t.year:", t.Year())
	fmt.Println("IncrDate t.month:", t.Month())
	fmt.Println("IncrDate t.day:", t.Day())
	nytt := t.AddDate(0, månader, veckor*7)
	fmt.Println("IncrDate nytt datum:", nytt.Format("2006-01-02"))
	return nytt.Format("2006-01-02")
}

func showFastaTransaktioner(w http.ResponseWriter, req *http.Request) {
	now := time.Now()
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()
	
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)
	currDate := lastOfMonth.Format("2006-01-02")
	antal := GetCountPendingÖverföringar(db, currDate)
	if antal > 0 {
		fmt.Fprintf(w, "<p>%d fasta transaktioner till hela denna månaden väntar på att hanteras:<br>\n", antal)
		res, err := db.Query("SELECT FrånKonto,TillKonto,Belopp,Datum,HurOfta,Vad,Vem,Löpnr,Kontrollnr,TillDatum,Rakning FROM Överföringar WHERE Datum <= ?", currDate)
		
		if err != nil {
			log.Fatal(err)
			os.Exit(2)
		}
		
		var FrånKonto []byte  // size 40
		var TillKonto []byte  // size 40
		var Belopp []byte  // BCD / Decimal Precision 19
		var Datum []byte  // size 10
		var HurOfta []byte  // size 15
		var Vad []byte  // size 40
		var Vem []byte  // size 40
		var Löpnr []byte  // Autoinc Primary Key, index
		var Kontrollnr int  // Integer
		var TillDatum []byte  // size 10
		var Rakning []byte  // size 1
		
		fmt.Fprintf(w, "<table style=\"width:100%%\"><tr><th>Löpnr</th><th>Frånkonto</th><th>Tillkonto/Plats</th><th>Belopp</th><th>Datum</th><th>Hur Ofta</th><th>Vad</th><th>Vem</th><th>Kontrollnr</th><th>Till datum</th><th>Räkning</th><th>Agera</th>\n")
		for res.Next() {
			err = res.Scan(&FrånKonto, &TillKonto, &Belopp, &Datum, &HurOfta, &Vad, &Vem, &Löpnr, &Kontrollnr, &TillDatum, &Rakning)
			
			sqlStmt := ""
			sqlStmt += "<tr><td>" + toUtf8(Löpnr) + "</td>"
			sqlStmt += "<td>" + toUtf8(FrånKonto) + "</td>"
			sqlStmt += "<td>" + toUtf8(TillKonto) + "</td>"
			sqlStmt += "<td>" + toUtf8(Belopp) + "</td>"
			sqlStmt += "<td>" + toUtf8(Datum) + "</td>"
			sqlStmt += "<td>" + toUtf8(HurOfta) + "</td>"
			sqlStmt += "<td>" + toUtf8(Vad) + "</td>"
			sqlStmt += "<td>" + toUtf8(Vem) + "</td>"
			sqlStmt += "<td>" + strconv.Itoa(Kontrollnr) + "</td>"
			sqlStmt += "<td>" + toUtf8(TillDatum) + "</td>"
			sqlStmt += "<td>" + toUtf8(Rakning) + "</td>"
			sqlStmt += "<td>"
			sqlStmt += "<form method=\"POST\" action=\"/newtrans\">\n"
			sqlStmt += "<input type=\"hidden\" id=\"transtyp\" name=\"transtyp\" value=\"FastTrans\">\n"
			sqlStmt += "<input type=\"hidden\" id=\"transid\" name=\"transid\" value=\""+toUtf8(Löpnr)+"\">\n"
			sqlStmt += "<input type=\"submit\" value=\"Registrera\"></form>\n"
			sqlStmt += "</td>"

			sqlStmt += "</tr>\n"
			fmt.Fprintf(w, "%s", sqlStmt)
		}
		fmt.Fprintf(w, "</table>\n")
	}
}

func registreraFastTransaktion(w http.ResponseWriter, transid int) {
	fmt.Fprintf(w, "Registrerar transaktion#"+strconv.Itoa(transid))
	if db == nil {
		fmt.Fprintf(w, "registreraFastTransaktion: No database open<p>\n")
		return
	}
	fmt.Fprintf(w, "<p>\n")
	// Retrieve repeating transaction
	var res *sql.Rows
	res, err := db.Query("SELECT FrånKonto,TillKonto,Belopp,Datum,HurOfta,Vad,Vem,Löpnr,Kontrollnr,TillDatum,Rakning FROM Överföringar WHERE Löpnr = ?", transid)
	if err != nil {
		log.Println("registreraFastTransaktion: SELECT ERROR")
		log.Println(err)
		return
	}
	
	var FrånKonto []byte  // size 40
	var TillKonto []byte  // size 40
	var Belopp []byte  // BCD / Decimal Precision 19
	var Datum []byte  // size 10
	var HurOfta []byte  // size 15
	var Vad []byte  // size 40
	var Vem []byte  // size 40
	var Löpnr []byte  // Autoinc Primary Key, index
	var Kontrollnr []byte  // Integer
	var TillDatum []byte  // size 10
	var Rakning []byte  // size 1
	
	fmt.Fprintf(w, "<table style=\"width:100%%\"><tr><th>Löpnr</th><th>Frånkonto</th><th>Tillkonto/Plats</th><th>Belopp</th><th>Datum</th><th>Hur Ofta</th><th>Vad</th><th>Vem</th><th>Kontrollnr</th><th>Till datum</th><th>Räkning</th><th>Agera</th>\n")
	res.Next()
	err = res.Scan(&FrånKonto, &TillKonto, &Belopp, &Datum, &HurOfta, &Vad, &Vem, &Löpnr, &Kontrollnr, &TillDatum, &Rakning)
	if err != nil {
		log.Println("registreraFastTransaktion: SCAN ERROR")
		log.Println(err)
		log.Println("registreraFastTransaktion: Bail out")
		fmt.Fprintf(w, "<tr>Bail out</tr>\n")
		return
	}

	sqlStmt := ""
	sqlStmt += "<tr><td>" + toUtf8(Löpnr) + "</td>"
	sqlStmt += "<td>" + toUtf8(FrånKonto) + "</td>"
	sqlStmt += "<td>" + toUtf8(TillKonto) + "</td>"
	sqlStmt += "<td>" + toUtf8(Belopp) + "</td>"
	sqlStmt += "<td>" + toUtf8(Datum) + "</td>"
	sqlStmt += "<td>" + toUtf8(HurOfta) + "</td>"
	sqlStmt += "<td>" + toUtf8(Vad) + "</td>"
	sqlStmt += "<td>" + toUtf8(Vem) + "</td>"
	sqlStmt += "<td>" + toUtf8(Kontrollnr) + "</td>"
	sqlStmt += "<td>" + toUtf8(TillDatum) + "</td>"
	sqlStmt += "<td>" + toUtf8(Rakning) + "</td>"
	sqlStmt += "</tr>\n"
	fmt.Fprintf(w, "%s", sqlStmt)
	fmt.Fprintf(w, "</table>\n")

	// Register transaction
	if toUtf8(Vad) == "---" {
		// Fasta överföringar
		fmt.Fprintf(w, "Registrerar Överföring...<br> ")
		
		sqlStatement := `
INSERT INTO Transaktioner (FrånKonto,TillKonto,Typ,Datum,Vad,Vem,Belopp,"Text")
VALUES (?,?,?,?,?,?,?,?)`
		_, err = db.Exec(sqlStatement, toUtf8(FrånKonto), toUtf8(TillKonto), "Överföring", toUtf8(Datum), "---", toUtf8(Vem), strings.ReplaceAll(toUtf8(Belopp), ".", ","), "Fast transaktion wHHEK")
		if err != nil {
			panic(err)
		}
		
		fmt.Fprintf(w, "<table style=\"width:100%%\"><tr><th>Frånkonto</th><th>Tillkonto</th><th>Typ</th><th>Datum</th><th>Vem</th><th>Belopp</th><th>Text</th>\n")
		sqlStmt := "<tr>"
		sqlStmt += "<td>" + toUtf8(FrånKonto) + "</td>"
		sqlStmt += "<td>" + toUtf8(TillKonto) + "</td>"
		sqlStmt += "<td>" + "Överföring" + "</td>"
		sqlStmt += "<td>" + toUtf8(Datum) + "</td>"
		sqlStmt += "<td>" + toUtf8(Vem) + "</td>"
		sqlStmt += "<td>" + toUtf8(Belopp) + "</td>"
		sqlStmt += "</tr>"
		fmt.Fprintf(w, "%s", sqlStmt)
		fmt.Fprintf(w, "</table>\n")
	} else if toUtf8(FrånKonto) == "---" {
		// Fasta inkomster
		fmt.Fprintf(w, "Registrerar Insättning...<br> ")
		
		sqlStatement := `
INSERT INTO Transaktioner (FrånKonto,TillKonto,Typ,Datum,Vad,Vem,Belopp,"Text")
VALUES (?,?,?,?,?,?,?,?)`
		_, err = db.Exec(sqlStatement, "---", toUtf8(TillKonto), "Insättning", toUtf8(Datum), toUtf8(Vad), toUtf8(Vem), strings.ReplaceAll(toUtf8(Belopp), ".", ","), "Fast transaktion wHHEK")
		if err != nil {
			panic(err)
		}
		
		fmt.Fprintf(w, "<table style=\"width:100%%\"><tr><th>Konto</th><th>Typ</th><th>Vad</th><th>Datum</th><th>Vem</th><th>Belopp</th><th>Text</th>\n")
		sqlStmt := "<tr>"
		sqlStmt += "<td>" + toUtf8(TillKonto) + "</td>"
		sqlStmt += "<td>" + "Insättning" + "</td>"
		sqlStmt += "<td>" + toUtf8(Vad) + "</td>"
		sqlStmt += "<td>" + toUtf8(Datum) + "</td>"
		sqlStmt += "<td>" + toUtf8(Vem) + "</td>"
		sqlStmt += "<td>" + toUtf8(Belopp) + "</td>"
		sqlStmt += "</tr>"
		fmt.Fprintf(w, "%s", sqlStmt)
		fmt.Fprintf(w, "</table>\n")
	} else {
		// Fasta utgifter
		fmt.Fprintf(w, "Registrerar Fast Utgift...<br> ")
		
		sqlStatement := `
INSERT INTO Transaktioner (FrånKonto,TillKonto,Typ,Datum,Vad,Vem,Belopp,"Text")
VALUES (?,?,?,?,?,?,?,?)`
		_, err = db.Exec(sqlStatement, toUtf8(FrånKonto), toUtf8(TillKonto), "Fast Utgift", toUtf8(Datum), toUtf8(Vad), toUtf8(Vem), strings.ReplaceAll(toUtf8(Belopp), ".", ","), "Fast transaktion wHHEK")
		if err != nil {
			panic(err)
		}
		
		fmt.Fprintf(w, "<table style=\"width:100%%\"><tr><th>Frånkonto</th><th>Plats</th><th>Typ</th><th>Vad</th><th>Datum</th><th>Vem</th><th>Belopp</th><th>Text</th>\n")
		sqlStmt := "<tr>"
		sqlStmt += "<td>" + toUtf8(FrånKonto) + "</td>"
		sqlStmt += "<td>" + toUtf8(TillKonto) + "</td>"
		sqlStmt += "<td>" + "Fast Utgift" + "</td>"
		sqlStmt += "<td>" + toUtf8(Vad) + "</td>"
		sqlStmt += "<td>" + toUtf8(Datum) + "</td>"
		sqlStmt += "<td>" + toUtf8(Vem) + "</td>"
		sqlStmt += "<td>" + toUtf8(Belopp) + "</td>"
		sqlStmt += "</tr>"
		fmt.Fprintf(w, "%s", sqlStmt)
		fmt.Fprintf(w, "</table>\n")
	}

	// Update repeating transaction
	var newDatum string;
	switch toUtf8(HurOfta) {
        case "Varannan vecka":
		newDatum = IncrDate(toUtf8(Datum), 2, 0)
	case "1":
		fallthrough
	case "Varje månad":
		newDatum = IncrDate(toUtf8(Datum), 0, 1)
	case "2":
		fallthrough
	case "Varannan månad":
		newDatum = IncrDate(toUtf8(Datum), 0, 2)
	case "3":
		fallthrough
        case "Varje kvartal":
		newDatum = IncrDate(toUtf8(Datum), 0, 3)
	case "6":
		fallthrough
	case "Varje halvår":
		newDatum = IncrDate(toUtf8(Datum), 0, 6)
	case "12":
		fallthrough
	case "Varje år":
		newDatum = IncrDate(toUtf8(Datum), 0, 12)
	default:
		log.Fatal("Okänd periodicitet: "+toUtf8(HurOfta))
		os.Exit(2)

	}
	sqlStatement := `UPDATE Överföringar SET Datum = ? WHERE Löpnr = ?`
	_, err = db.Exec(sqlStatement, newDatum, transid)
	if err != nil {
		panic(err)
	}

	fmt.Fprintf(w, "<p>\n")
}
