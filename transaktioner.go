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

		//fmt.Println("date:", record.date)
		//fmt.Println("text:", record.comment)

		result = append(result, record)
	}
	res.Close()
	return result
}

//go:embed html/transakt4.html
var htmltrans4 string
type Trans4Data struct {
	DBName string
	Transaktioner []TransactionType
}

func printTransactions(w http.ResponseWriter, db *sql.DB, startDate string, endDate string, limitcomment string, fromacc string, toacc string, place string, vad string, minbelopp string, maxbelopp string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var err error
	var res *sql.Rows

	var querystring string
	var queryargs []string

	querystring1 := `SELECT FrånKonto,TillKonto,Typ,Datum,Vad,Vem,Belopp,Löpnr,Saldo,Fastöverföring,[Text] from transaktioner
`
	querystring2 := `
 where (datum < ?) and (datum >= ?)
`
	querystring3 := `
order by datum,löpnr
`
	queryargs = append(queryargs, endDate)
	queryargs = append(queryargs, startDate)
	
	if len(limitcomment) > 0 {
		querystring2 = querystring2 + ` and (Text like ?) `
		queryargs = append(queryargs, limitcomment)
	}
	if len(fromacc) > 0 {
		querystring2 = querystring2 + ` and (FrånKonto = ?) `
		queryargs = append(queryargs, fromacc)
	}
	if len(toacc) > 0 {
		querystring2 = querystring2 + ` and (TillKonto = ?) `
		queryargs = append(queryargs, toacc)
	}
	if len(place) > 0 {
		querystring2 = querystring2 + ` and (TillKonto = ?) `
		queryargs = append(queryargs, place)
	}
	if len(vad) > 0 {
		querystring2 = querystring2 + ` and (Vad = ?) `
		queryargs = append(queryargs, vad)
	}
	if len(minbelopp) > 0 {
		querystring2 = querystring2 + ` and (Belopp >= ?) `
		queryargs = append(queryargs, minbelopp)
	}
	if len(maxbelopp) > 0 {
		querystring2 = querystring2 + ` and (Belopp <= ?) `
		queryargs = append(queryargs, maxbelopp)
	}

	querystring = querystring1 + querystring2 + querystring3
	b := make([]interface{},0,len(queryargs))
	for _,i:= range queryargs {
		b = append(b,i)
	}
	res, err = db.QueryContext(ctx, querystring, b...)
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
	
	var transactions []TransactionType

	for res.Next() {
		err = res.Scan(&fromAcc, &toAcc, &tType, &date, &what, &who, &amount, &nummer, &saldo, &fixed, &comment)
		
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

		transaction.Text = toUtf8(comment)
		transaction.Fixed = strconv.FormatBool(fixed)
		transactions = append(transactions, transaction)
	}
	res.Close()
	
	t := template.New("Transaktion4")
	t, _ = t.Parse(htmltrans4)
	data := Trans4Data{
		DBName: currentDatabase,
		Transaktioner: transactions,
	}
	err = t.Execute(w, data)
	if err != nil {
		log.Println("While serving HTTP trans4: ", err)
	}
}

func isobytetodate(rawdate []byte) (time.Time, error) {
	return time.Parse("2006-01-02", toUtf8(rawdate))
}

//go:embed html/transakt3.html
var htmltrans3 string
type Trans3Data struct {
	DBFirstDay string
	DBLastDay string
	FormStartDay string
	FormLastDay string
	FormComment string
	Kontonamn []string
	Platser []string
	Vad     []string
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
		var date []byte // size 10
		err = db.QueryRow("SELECT MIN(Datum) FROM Transaktioner").Scan(&date)
		kontostartdatum, err := isobytetodate(date)
		if err != nil {
			log.Print(err)
		}

		err = db.QueryRow("SELECT MAX(Datum) FROM Transaktioner").Scan(&date)
		kontoslutdatum, err := isobytetodate(date)
		if err != nil {
			log.Print(err)
		}

		printTransactions(w, db, startDate, endDate, req.FormValue("comment"), req.FormValue("fromacc"), req.FormValue("toacc"), req.FormValue("place"), req.FormValue("vad"), req.FormValue("minamount"), req.FormValue("maxamount"))

		trtypes := []string{""}
		trtypes = append(trtypes, getTypeInNames()...)
		trtypes = append(trtypes, getTypeOutNames()...)
		t := template.New("Transaktion3")
		t, _ = t.Parse(htmltrans3)
		data := Trans3Data{
			DBFirstDay: kontostartdatum.Format("2006-01-02"),
			DBLastDay: kontoslutdatum.Format("2006-01-02"),
			FormStartDay: startDate,
			FormLastDay: endDate,
			FormComment: req.FormValue("comment"),
			Kontonamn: append([]string{""}, getAccNames()...),
			Platser: append([]string{""}, getPlaceNames()...),
			Vad: trtypes,
		}
		err = t.Execute(w, data)
		if err != nil {
			log.Println("While serving HTTP trans3: ", err)
		}
	}
}

//go:embed html/transakt1.html
var htmltrans1 string
//go:embed html/transakt2.html
var htmltrans2 string

func transactions(w http.ResponseWriter, req *http.Request) {
	t := template.New("Transaktion1")
	t, _ = t.Parse(htmltrans1)
	err := t.Execute(w, nil)
	if err != nil {
		log.Println("While serving HTTP trans1: ", err)
	}

	err = req.ParseForm()
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

	t = template.New("Transaktion2")
	t, _ = t.Parse(htmltrans2)
	err = t.Execute(w, nil)
	if err != nil {
		log.Println("While serving HTTP trans2: ", err)
	}
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
	Kontonamn  []string
	Platser    []string
	Personer   []string
	VadInkomst []string
	VadUtgift  []string
	Today      string
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


	// del 2
	currentTime := time.Now()
	currDate := currentTime.Format("2006-01-02")

	tmpl2 := template.New("wHHEK newtrans2")
	tmpl2, _ = tmpl2.Parse(newtrans2)
	data2 := NewTrans2Data{
		Kontonamn: kontonamn,
		Platser: platser,
		Personer: personer,
		VadInkomst: vadInkomst,
		VadUtgift: vadUtgift,
		Today: currDate,
	}
	_ = tmpl2.Execute(w, data2)
}

func addTransaktionSQL(transtyp string, fromacc string, toacc string, date string, what string, who string, summa decimal.Decimal, text string) {
	var amount = "NONE"

	amount = AmountDec2DBStr(summa)
	if len(text) < 1 {
		text = " "
	}
	
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
	updateKontoSaldo(toacc, saldo)
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
	updateKontoSaldo(fromacc, saldo)
}

func addTransaktionUttag(fromacc string, date string, who string, summa decimal.Decimal, text string) {
	var transtyp = "Uttag"

	// TODO: Check length of "text"
	// TODO: Check date format
	// TODO: Check toacc valid
	// TODO: Check what valid
	// TODO: Check who valid

	addTransaktionSQL(transtyp, fromacc, "Plånboken", date, "---", who, summa, text)

	saldo := saldoKonto(db, fromacc, "")
	updateKontoSaldo(fromacc, saldo)

	saldo = saldoKonto(db, "Plånboken", "")
	updateKontoSaldo("Plånboken", saldo)
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
	updateKontoSaldo(fromacc, saldo)

	saldo = saldoKonto(db, toacc, "")
	updateKontoSaldo(toacc, saldo)
}

func addtransaction(w http.ResponseWriter, req *http.Request) {
	log.Println("addtransaction: start")
	
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
		sqlStmt += "<td>" + Dec2Str(amount) + "</td>"
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
		sqlStmt += "<td>" + Dec2Str(amount) + "</td>"
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

		addTransaktionUttag(fromacc, date, who, amount, text)

		_, _ = fmt.Fprintf(w, "<table style=\"width:100%%\"><tr><th>Frånkonto</th><th>Typ</th><th>Datum</th><th>Vem</th><th>Belopp</th><th>Text</th>\n")
		sqlStmt := "<tr>"
		sqlStmt += "<td>" + fromacc + "</td>"
		sqlStmt += "<td>" + transtyp + "</td>"
		sqlStmt += "<td>" + date + "</td>"
		sqlStmt += "<td>" + who + "</td>"
		sqlStmt += "<td>" + Dec2Str(amount) + "</td>"
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
		sqlStmt += "<td>" + Dec2Str(amount) + "</td>"
		sqlStmt += "<td>" + html.EscapeString(text) + "</td>\n"
		sqlStmt += "</tr>"
		_, _ = fmt.Fprintf(w, "%s", sqlStmt)
		_, _ = fmt.Fprintf(w, "</table>\n")
	}
	log.Println("addtransaction: end")
}

//go:embed html/transakt5.html
var htmltrans5 string
type Trans5Data struct {
	FromAcc string
	Dest string
	Typ string
	Datum string
	Vad string
	Vem string
	Belopp string
	Fixed string
	Text string
	Löpnr string
}

func editformTransaction(w http.ResponseWriter, lopnr int, db *sql.DB) {
	fmt.Println("editformTransaktion lopnr: ", lopnr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	err := db.QueryRowContext(ctx,
		`SELECT FrånKonto,TillKonto,Typ,Datum,Vad,Vem,Belopp,Saldo,Fastöverföring,Text FROM transaktioner WHERE (Löpnr=?)`, lopnr).Scan(&fromAcc, &toAcc, &tType, &date, &what, &who, &amount, &saldo, &fixed, &comment)
	if err != nil {
		log.Fatal(err)
	}

	t := template.New("Transaktion5")
	t, _ = t.Parse(htmltrans5)
	data := Trans5Data{
		FromAcc: toUtf8(fromAcc),
		Dest: toUtf8(toAcc),
		Typ: toUtf8(tType),
		Datum: toUtf8(date),
		Vad: toUtf8(what),
		Vem: toUtf8(who),
		Belopp: toUtf8(amount),
		Fixed: strconv.FormatBool(fixed),
		Text: toUtf8(comment),
		Löpnr: strconv.Itoa(lopnr),
	}
	err = t.Execute(w, data)
	if err != nil {
		log.Println("While serving HTTP trans4: ", err)
	}
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

	var antal int

	err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM Transaktioner`).Scan(&antal)
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
	res.Close()
	return result
}
