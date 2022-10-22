//-*- coding: utf-8 -*-

package main

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"html"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/shopspring/decimal" // MIT License
)

type transaction struct {
	lopnr   int
	fromAcc string
	toAcc   string
	tType   string
	what    string
	date    time.Time
	who     string
	amount  decimal.Decimal
	comment string
	fixed   bool
}

func getTransactionsInDateRange(db *sql.DB, kontonamn string, startDate string, endDate string) []transaction {
	//fmt.Println("printTransactions startDate:", startDate)
	//fmt.Println("printTransactions endDate:", endDate)
	//fmt.Println("printTransactions kontonamn:", kontonamn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var err error
	var res *sql.Rows

	res, err = db.QueryContext(ctx,
		`SELECT FrånKonto,TillKonto,Typ,Datum,Vad,Vem,Belopp,Löpnr,Saldo,Fastöverföring,[Text] from transaktioner
  where (datum < ?) and (datum >= ?) and ((FrånKonto = ?) or (TillKonto = ?))
order by datum,löpnr`, endDate, startDate, kontonamn, kontonamn)
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

	var result []transaction

	for res.Next() {
		var record transaction
		err = res.Scan(&fromAcc, &toAcc, &tType, &date, &what, &who, &amount, &nummer, &saldo, &fixed, &comment)

		record.lopnr = nummer
		record.fromAcc = toUtf8(fromAcc)
		record.toAcc = toUtf8(toAcc)
		record.tType = toUtf8(tType)
		record.what = toUtf8(what)
		record.date, err = time.Parse("2006-01-02", toUtf8(date))
		record.who = toUtf8(who)
		record.amount, err = decimal.NewFromString(toUtf8(amount))
		record.comment = toUtf8(comment)
		record.fixed = fixed

		fmt.Println("date:", record.date)
		fmt.Println("text:", record.comment)

		result = append(result, record)
	}
	return result
}

func printTransactions(w http.ResponseWriter, db *sql.DB, startDate string, endDate string, limitcomment string) {
	fmt.Println("printTransactions startDate:", startDate)
	fmt.Println("printTransactions endDate:", endDate)
	fmt.Println("printTransactions comment:", limitcomment, len(limitcomment))

	_, _ = fmt.Fprintf(w, "<h1>%s</h1>\n", currentDatabase)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var err error
	var res *sql.Rows

	if len(limitcomment) > 0 {
		fmt.Println("Query with text limit")

		res, err = db.QueryContext(ctx,
			`SELECT FrånKonto,TillKonto,Typ,Datum,Vad,Vem,Belopp,Löpnr,Saldo,Fastöverföring,[Text] from transaktioner
  where (datum < ?) and (datum >= ?) and (Text like ?)
order by datum,löpnr`, endDate, startDate, limitcomment)
	} else {
		fmt.Println("Query without text limit")

		res, err = db.QueryContext(ctx,
			`SELECT FrånKonto,TillKonto,Typ,Datum,Vad,Vem,Belopp,Löpnr,Saldo,Fastöverföring,[Text] from transaktioner
  where (datum < ?) and (datum >= ?)
order by datum,löpnr`, endDate, startDate)
	}
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

	_, _ = fmt.Fprintf(w, "<table style=\"width:100%%\"><tr><th>Löpnr</th><th>Frånkonto</th><th>Tillkonto/Plats</th><th>Typ</th><th>Vad</th><th>Datum</th><th>Vem</th><th>Belopp</th><th>Text</th><th>Fast överföring</th>\n")
	_, _ = fmt.Fprintf(w, "<th>Redigera</th><th>Radera</th>\n")
	for res.Next() {
		err = res.Scan(&fromAcc, &toAcc, &tType, &date, &what, &who, &amount, &nummer, &saldo, &fixed, &comment)
		fmt.Println("date:", toUtf8(date))
		fmt.Println("text:", comment)
		fmt.Println("text:", toUtf8(comment))

		_, _ = fmt.Fprintf(w, "<tr><td>"+strconv.Itoa(nummer)+"</td>")
		_, _ = fmt.Fprintf(w, "<td>"+toUtf8(fromAcc)+"</td>")
		_, _ = fmt.Fprintf(w, "<td>"+toUtf8(toAcc)+"</td>")
		_, _ = fmt.Fprintf(w, "<td>"+toUtf8(tType)+"</td>")
		_, _ = fmt.Fprintf(w, "<td>"+toUtf8(what)+"</td>")
		_, _ = fmt.Fprintf(w, "<td>"+toUtf8(date)+"</td>")
		_, _ = fmt.Fprintf(w, "<td>"+toUtf8(who)+"</td>")
		_, _ = fmt.Fprintf(w, "<td>"+toUtf8(amount)+"</td>")
		_, _ = fmt.Fprintf(w, "<td>"+html.EscapeString(toUtf8(comment))+"</td>\n")
		_, _ = fmt.Fprintf(w, "<td>"+strconv.FormatBool(fixed)+"</td>")
		_, _ = fmt.Fprintf(w, "<td><form method=\"POST\" action=\"/transactions\"><input type=\"hidden\" id=\"lopnr\" name=\"lopnr\" value=\"%d\"><input type=\"hidden\" id=\"action\" name=\"action\" value=\"editform\"><input type=\"submit\" value=\"Redigera\"></form></td>\n", nummer)
		_, _ = fmt.Fprintf(w, "<td><form method=\"POST\" action=\"/transactions\"><input type=\"hidden\" id=\"lopnr\" name=\"lopnr\" value=\"%d\"><input type=\"hidden\" id=\"action\" name=\"action\" value=\"radera\"><input type=\"checkbox\" id=\"OK\" name=\"OK\" required><label for=\"OK\">OK</label><input type=\"submit\" value=\"Radera\"></form></td></tr>\n", nummer)
	}
	_, _ = fmt.Fprintf(w, "</table>\n")
}

func isobytetodate(rawdate []byte) (time.Time, error) {
	return time.Parse("2006-01-02", toUtf8(rawdate))
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
		_, _ = fmt.Fprintf(w, "Transactions: No database open<p>\n")
	} else {
		res1 := db.QueryRow("SELECT MIN(Datum) FROM Transaktioner")
		var date []byte // size 10
		err = res1.Scan(&date)
		kontostartdatum, err := isobytetodate(date)
		if err != nil {
			log.Print(err)
		}

		res1 = db.QueryRow("SELECT MAX(Datum) FROM Transaktioner")
		err = res1.Scan(&date)
		kontoslutdatum, err := isobytetodate(date)
		if err != nil {
			log.Print(err)
		}

		printTransactions(w, db, startDate, endDate, req.FormValue("comment"))
		_, _ = fmt.Fprintf(w, "Kontots första transaktion %s<br>\n", kontostartdatum.Format("2006-01-02"))
		_, _ = fmt.Fprintf(w, "Kontots sista transaktion %s<p>\n", kontoslutdatum.Format("2006-01-02"))

		_, _ = fmt.Fprintf(w, "<form method=\"POST\" action=\"/transactions\">\n")
		_, _ = fmt.Fprintf(w, "<label for=\"startdate\">Startdatum:</label>")
		_, _ = fmt.Fprintf(w, "	<input type=\"date\" id=\"startdate\" name=\"startdate\" value=\"%s\" title=\"Inklusive\">", startDate)
		_, _ = fmt.Fprintf(w, "<label for=\"enddate\">Slutdatum:</label>")
		_, _ = fmt.Fprintf(w, "	<input type=\"date\" id=\"enddate\" name=\"enddate\" value=\"%s\" title=\"Exclusive\">", endDate)
		_, _ = fmt.Fprintf(w, "<label for=\"comment\">Kommentar:</label>")
		_, _ = fmt.Fprintf(w, "	<input id=\"comment\" name=\"comment\" value=\"%s\" placeholder=\"wildcards %%_\" title=\"Söktext\n%% är noll, ett eller många tecken.\n_ är ett tecken.\nTomt fält betyder inget filtreras.\">", req.FormValue("comment"))

		_, _ = fmt.Fprintf(w, "<input type=\"submit\" value=\"Visa\"></form>\n")

		_, _ = fmt.Fprintf(w, "<form method=\"POST\" action=\"/transactions\">\n")
	}
}

func transactions(w http.ResponseWriter, req *http.Request) {
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

	formaction := req.FormValue("action")
	var lopnr = -1
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

	_, _ = fmt.Fprintf(w, "<a href=\"summary\">Översikt</a>\n")
	_, _ = fmt.Fprintf(w, "</body>\n")
	_, _ = fmt.Fprintf(w, "</html>\n")
}

func raderaTransaction(w http.ResponseWriter, lopnr int, db *sql.DB) {
	fmt.Println("raderaTransaction lopnr: ", lopnr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := db.ExecContext(ctx,
		`DELETE FROM transaktioner WHERE (Löpnr=?)`, lopnr)

	if err != nil {
		log.Fatal(err)
	}
	_, _ = fmt.Fprintf(w, "Transaktion med löpnr %d raderad.<br>", lopnr)
}

//go:embed html/newtransaction1.html
var newtrans1 string
//go:embed html/newtransaction2.html
var newtrans2 string

type NewTrans1Data struct {
	PageName string
}
type NewTrans2Data struct {
	Kontonamn []string
	Platser []string
	Personer []string
	VadInkomst []string
	VadUtgift []string
}


func newtransaction(w http.ResponseWriter, req *http.Request) {
	// Common
	kontonamn := getAccNames()

	platser := getPlaceNames()
	personer := getPersonNames()
	vadInkomst := getTypeInNames()
	vadUtgift := getTypeOutNames()

	// del 1
	tmpl1 := template.New("wHHEK newtrans1")
	tmpl1, _ = tmpl1.Parse(newtrans1)
	data := NewTrans1Data{
		PageName: currentDatabase,
	}
	_ = tmpl1.Execute(w, data)


	// handle submitted form
	addtransaction(w, req)

	// del 2
	tmpl2 := template.New("wHHEK newtrans2")
	tmpl2, _ = tmpl2.Parse(newtrans2)
	data2 := NewTrans2Data{
		Kontonamn: kontonamn,
		Platser: platser,
		Personer: personer,
		VadInkomst: vadInkomst,
		VadUtgift: vadUtgift,
	}
	_ = tmpl2.Execute(w, data2)
}

func addTransaktionSQL(transtyp string, fromacc string, toacc string, date string, what string, who string, summa decimal.Decimal, text string) {
	var amount = "NONE"

	amount = AmountDec2DBStr(summa)
	
	sqlStatement := `
	INSERT INTO Transaktioner (FrånKonto,TillKonto,Typ,Datum,Vad,Vem,Belopp,Saldo,[Fastöverföring],[Text])
	VALUES (?,?,?,?,?,?,?,?,?,?)`
	fmt.Println("addTransaktionSQL: ", sqlStatement)
	fmt.Println("addTransaktionSQL: ", fromacc)
	fmt.Println("addTransaktionSQL: ", toacc)
	fmt.Println("addTransaktionSQL: ", transtyp)
	fmt.Println("addTransaktionSQL: ", date)
	fmt.Println("addTransaktionSQL: ", what)
	fmt.Println("addTransaktionSQL: ", who)
	fmt.Println("addTransaktionSQL: ", amount)
	fmt.Println("addTransaktionSQL: ", text)

	_, err := db.Exec(sqlStatement, fromacc, toacc, transtyp, date, what, who, amount, nil, false, text)
	if err != nil {
		log.Println("SQL err")
		log.Println("ny transaktionSQL: ", transtyp, fromacc, summa, toacc, date, what, who, amount, text)
		panic(err)
	}
}

func addTransaktionInsättning(toacc string, date string, what string, who string, summa decimal.Decimal, text string) {
	var transtyp = "Insättning"

	// TODO: Check length of "text"
	// TODO: Check date format
	// TODO: Check toacc valid
	// TODO: Check what valid
	// TODO: Check who valid

	addTransaktionSQL(transtyp, "---", toacc, date, what, who, summa, text)

	saldo := saldoKonto(db, toacc, "")
	updateKontoSaldo(toacc, saldo.String())

	saldo = saldoKonto(db, toacc, "")
	updateKontoSaldo(toacc, saldo.String())
}

func addTransaktionInköp(fromacc string, place string, date string, what string, who string, summa decimal.Decimal, text string, fixed bool) {
	var transtyp = "Inköp"
	if fixed {
		transtyp = "Fast Utgift"
	}
	// TODO: Check length of "text"
	// TODO: Check date format
	// TODO: Check toacc valid
	// TODO: Check what valid
	// TODO: Check who valid

	addTransaktionSQL(transtyp, fromacc, place, date, what, who, summa, text)

	saldo := saldoKonto(db, fromacc, "")
	updateKontoSaldo(fromacc, saldo.String())
}

func addTransaktionUttag(fromacc string, date string, what string, who string, summa decimal.Decimal, text string) {
	var transtyp = "Uttag"

	// TODO: Check length of "text"
	// TODO: Check date format
	// TODO: Check toacc valid
	// TODO: Check what valid
	// TODO: Check who valid

	addTransaktionSQL(transtyp, fromacc, "Plånboken", date, what, who, summa, text)

	saldo := saldoKonto(db, fromacc, "")
	updateKontoSaldo(fromacc, saldo.String())

	saldo = saldoKonto(db, "Plånboken", "")
	updateKontoSaldo("Plånboken", saldo.String())
}

func addTransaktionÖverföring(fromacc string, toacc string, date string, who string, summa decimal.Decimal, text string) {
	var transtyp = "Överföring"

	// TODO: Check length of "text"
	// TODO: Check date format
	// TODO: Check toacc valid
	// TODO: Check what valid
	// TODO: Check who valid

	addTransaktionSQL(transtyp, fromacc, toacc, date, "---", who, summa, text)

	saldo := saldoKonto(db, fromacc, "")
	updateKontoSaldo(fromacc, saldo.String())

	saldo = saldoKonto(db, toacc, "")
	updateKontoSaldo(toacc, saldo.String())
}

func addtransaction(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	transtyp := req.FormValue("transtyp")
	date := req.FormValue("date")
	who := req.FormValue("who")
	amountstr := req.FormValue("amount")
	amountstr = SanitizeAmount(amountstr)
	amount, err := decimal.NewFromString(amountstr)
	if err != nil {
		log.Println("OK: addtransaction, trasig/saknar amount ", amountstr, err)
	}

	text := req.FormValue("text")
	fmt.Println("Val tt: ", transtyp)
	fmt.Println("Val d: ", date)
	fmt.Println("Val w: ", who)
	fmt.Println("Val a: ", amount)
	fmt.Println("Val t: ", text)

	if transtyp == "Inköp" {
		fromacc := req.FormValue("fromacc")
		place := req.FormValue("place")
		what := req.FormValue("what")
		fmt.Println("Val: ", fromacc)
		fmt.Println("Val: ", place)
		fmt.Println("Val: ", what)

		_, _ = fmt.Fprintf(w, "Registrerar Inköp...<br> ")

		addTransaktionInköp(fromacc, place, date, what, who, amount, text, false)

		_, _ = fmt.Fprintf(w, "<table style=\"width:100%%\"><tr><th>Frånkonto</th><th>Plats</th><th>Typ</th><th>Vad</th><th>Datum</th><th>Vem</th><th>Belopp</th><th>Text</th>\n")
		sqlStmt := "<tr>"
		sqlStmt += "<td>" + fromacc + "</td>"
		sqlStmt += "<td>" + place + "</td>"
		sqlStmt += "<td>" + transtyp + "</td>"
		sqlStmt += "<td>" + what + "</td>"
		sqlStmt += "<td>" + date + "</td>"
		sqlStmt += "<td>" + who + "</td>"
		sqlStmt += "<td>" + amount.String() + "</td>"
		sqlStmt += "<td>" + html.EscapeString(text) + "</td>\n"
		sqlStmt += "</tr>"
		_, _ = fmt.Fprintf(w, "%s", sqlStmt)
		_, _ = fmt.Fprintf(w, "</table>\n")
	}
	if transtyp == "Insättning" {
		toacc := req.FormValue("toacc")
		what := req.FormValue("what")
		fmt.Println("Val: ", toacc)
		fmt.Println("Val: ", what)

		_, _ = fmt.Fprintf(w, "Registrerar Insättning...<br> ")
		addTransaktionInsättning(toacc, date, what, who, amount, text)

		_, _ = fmt.Fprintf(w, "<table style=\"width:100%%\"><tr><th>Konto</th><th>Typ</th><th>Vad</th><th>Datum</th><th>Vem</th><th>Belopp</th><th>Text</th>\n")
		sqlStmt := "<tr>"
		sqlStmt += "<td>" + toacc + "</td>"
		sqlStmt += "<td>" + transtyp + "</td>"
		sqlStmt += "<td>" + what + "</td>"
		sqlStmt += "<td>" + date + "</td>"
		sqlStmt += "<td>" + who + "</td>"
		sqlStmt += "<td>" + amount.String() + "</td>"
		sqlStmt += "<td>" + html.EscapeString(text) + "</td>\n"
		sqlStmt += "</tr>"
		_, _ = fmt.Fprintf(w, "%s", sqlStmt)
		_, _ = fmt.Fprintf(w, "</table>\n")
	}
	if transtyp == "Uttag" {
		fromacc := req.FormValue("fromacc")
		what := req.FormValue("what")
		fmt.Println("Val: ", fromacc)
		fmt.Println("Val: ", what)

		_, _ = fmt.Fprintf(w, "Registrerar Uttag...<br> ")

		addTransaktionUttag(fromacc, date, what, who, amount, text)

		_, _ = fmt.Fprintf(w, "<table style=\"width:100%%\"><tr><th>Frånkonto</th><th>Typ</th><th>Datum</th><th>Vem</th><th>Belopp</th><th>Text</th>\n")
		sqlStmt := "<tr>"
		sqlStmt += "<td>" + fromacc + "</td>"
		sqlStmt += "<td>" + transtyp + "</td>"
		sqlStmt += "<td>" + date + "</td>"
		sqlStmt += "<td>" + who + "</td>"
		sqlStmt += "<td>" + amount.String() + "</td>"
		sqlStmt += "<td>" + html.EscapeString(text) + "</td>\n"
		sqlStmt += "</tr>"
		_, _ = fmt.Fprintf(w, "%s", sqlStmt)
		_, _ = fmt.Fprintf(w, "</table>\n")
	}
	if transtyp == "Överföring" {
		fromacc := req.FormValue("fromacc")
		toacc := req.FormValue("toacc")
		fmt.Println("Val: ", fromacc)
		fmt.Println("Val: ", toacc)

		_, _ = fmt.Fprintf(w, "Registrerar Överföring...<br> ")

		addTransaktionÖverföring(fromacc, toacc, date, who, amount, text)

		_, _ = fmt.Fprintf(w, "<table style=\"width:100%%\"><tr><th>Frånkonto</th><th>Tillkonto</th><th>Typ</th><th>Datum</th><th>Vem</th><th>Belopp</th><th>Text</th>\n")
		sqlStmt := "<tr>"
		sqlStmt += "<td>" + fromacc + "</td>"
		sqlStmt += "<td>" + toacc + "</td>"
		sqlStmt += "<td>" + transtyp + "</td>"
		sqlStmt += "<td>" + date + "</td>"
		sqlStmt += "<td>" + who + "</td>"
		sqlStmt += "<td>" + amount.String() + "</td>"
		sqlStmt += "<td>" + html.EscapeString(text) + "</td>\n"
		sqlStmt += "</tr>"
		_, _ = fmt.Fprintf(w, "%s", sqlStmt)
		_, _ = fmt.Fprintf(w, "</table>\n")
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
	}

	_, _ = fmt.Fprintf(w, "Redigera transaktion<br>")
	_, _ = fmt.Fprintf(w, "<form method=\"POST\" action=\"/transactions\">")

	PrintEditCellText(w, "fromAcc", "Från konto", toUtf8(fromAcc))
	PrintEditCellText(w, "toAcc", "Till konto", toUtf8(toAcc))
	PrintEditCellText(w, "tType", "Typ", toUtf8(tType))
	PrintEditCellText(w, "date", "Datum", toUtf8(date))
	PrintEditCellText(w, "what", "Vad", toUtf8(what))
	PrintEditCellText(w, "who", "Vem", toUtf8(who))
	PrintEditCellText(w, "amount", "Summa", toUtf8(amount))
	PrintEditCellText(w, "fixed", "Fast transaktion", strconv.FormatBool(fixed))
	PrintEditCellText(w, "comment", "Text", toUtf8(comment))

	_, _ = fmt.Fprintf(w, "<input type=\"hidden\" id=\"lopnr\" name=\"lopnr\" value=\"%d\">", lopnr)
	_, _ = fmt.Fprintf(w, "<input type=\"hidden\" id=\"action\" name=\"action\" value=\"update\">")
	_, _ = fmt.Fprintf(w, "<input type=\"submit\" value=\"Uppdatera\">")
	_, _ = fmt.Fprintf(w, "</form>\n")
	_, _ = fmt.Fprintf(w, "<p>\n")
}

func updateTransaction(w http.ResponseWriter, lopnr int, req *http.Request, db *sql.DB) {
	fmt.Println("updateTransaktion lopnr: ", lopnr)

	var fromAcc = ""
	if len(req.FormValue("fromAcc")) > 0 {
		fromAcc = req.FormValue("fromAcc")
	}
	var toAcc = ""
	if len(req.FormValue("toAcc")) > 0 {
		toAcc = req.FormValue("toAcc")
	}
	var tType = ""
	if len(req.FormValue("tType")) > 0 {
		tType = req.FormValue("tType")
	}
	var date = ""
	if len(req.FormValue("date")) > 0 {
		date = req.FormValue("date")
	}
	var what = ""
	if len(req.FormValue("what")) > 0 {
		what = req.FormValue("what")
	}
	var who = ""
	if len(req.FormValue("who")) > 0 {
		who = req.FormValue("who")
	}
	var amount = ""
	if len(req.FormValue("amount")) > 0 {
		amount = SanitizeAmount(req.FormValue("amount"))
	}
	var fixed = false
	if len(req.FormValue("fixed")) > 0 {
		var fixedString = ""
		fixedString = req.FormValue("fixed")
		fixed, _ = strconv.ParseBool(fixedString)
	}

	var comment = ""
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
		AmountStr2DBStr(amount),
		fixed,
		comment,
		lopnr)

	if err != nil {
		log.Fatal(err)
	}

	_, _ = fmt.Fprintf(w, "Transaktion %d uppdaterad.<br>", lopnr)
}

func antalTransaktioner(db *sql.DB) int {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	res1 := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM Transaktioner`)

	var antal int

	err := res1.Scan(&antal)
	if err != nil {
		log.Fatal(err)
	}

	return antal
}

func hämtaTransaktion(lopnr int) (result transaction) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var err error
	var res *sql.Rows

	res, err = db.QueryContext(ctx,
		`SELECT FrånKonto,TillKonto,Typ,Datum,Vad,Vem,Belopp,Löpnr,Saldo,Fastöverföring,Text from transaktioner
  where Löpnr = ?`, lopnr)
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
		var record transaction
		err = res.Scan(&fromAcc, &toAcc, &tType, &date, &what, &who, &amount, &nummer, &saldo, &fixed, &comment)

		record.lopnr = nummer
		record.fromAcc = toUtf8(fromAcc)
		record.toAcc = toUtf8(toAcc)
		record.tType = toUtf8(tType)
		record.what = toUtf8(what)
		record.date, err = isobytetodate(date)
		record.who = toUtf8(who)
		record.amount, err = decimal.NewFromString(toUtf8(amount))
		record.comment = toUtf8(comment)
		record.fixed = fixed

		result = record
	}
	return result
}
