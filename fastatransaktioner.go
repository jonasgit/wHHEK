//-*- coding: utf-8 -*-

package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/shopspring/decimal" // MIT License
)

type fixedtransaction struct {
	lopnr   int
	vernum  int
	fromAcc string
	toAcc   string
	what    string
	date    time.Time
	todate  time.Time
	who     string
	amount  decimal.Decimal
	HurOfta string
	comment string
	rakning bool
}

func CurrDate() string {
	now := time.Now()
	currDate := now.Format("2006-01-02")
	return currDate
}

func IncrDate(datum string, veckor int, månader int) string {
	//log.Println("IncrDate start ")
	year, _ := strconv.Atoi(datum[0:4])
	var month time.Month
	monthval, _ := strconv.Atoi(datum[5:7])
	switch monthval {
	case 1:
		month = time.January
	case 2:
		month = time.February
	case 3:
		month = time.March
	case 4:
		month = time.April
	case 5:
		month = time.May
	case 6:
		month = time.June
	case 7:
		month = time.July
	case 8:
		month = time.August
	case 9:
		month = time.September
	case 10:
		month = time.October
	case 11:
		month = time.November
	case 12:
		month = time.December
	default:
		log.Fatal("Okänd månad: ", monthval)
	}
	day, _ := strconv.Atoi(datum[8:10])
	location := time.FixedZone("CET", 0)
	t := time.Date(year, month, day, 12, 0, 0, 0, location)
	nytt := t.AddDate(0, månader, veckor*7)
	//fix date at end of month spilling over to next month
	if månader != 0 {
		if veckor != 0 {
			log.Fatal("Inte tillåtet med både veckor och månader")
		}
		if nytt.Day() != day {
			nytt = BeginningOfMonth(nytt)
			nytt = nytt.AddDate(0, 0, -1)
		}
	}

	//log.Println("IncrDate slut ", nytt)
	return nytt.Format("2006-01-02")
}

func BeginningOfMonth(date time.Time) time.Time {
	return date.AddDate(0, 0, -date.Day()+1)
}

func EndOfMonth(date time.Time) time.Time {
	return date.AddDate(0, 1, -date.Day())
}

func showFastaTransaktioner(w http.ResponseWriter, db *sql.DB) {
	now := time.Now()
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()

	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)
	currDate := lastOfMonth.Format("2006-01-02")
	antal := GetCountPendingÖverföringar(db, currDate)
	if antal > 0 {
		_, _ = fmt.Fprintf(w, "<p>%d fasta transaktioner till hela denna månaden väntar på att hanteras:<br>\n", antal)
		res, err := db.Query("SELECT FrånKonto,TillKonto,Belopp,Datum,HurOfta,Vad,Vem,Löpnr,Kontrollnr,TillDatum,Rakning FROM Överföringar WHERE Datum <= ?", currDate)

		if err != nil {
			log.Fatal(err)
		}

		var FrånKonto []byte // size 40
		var TillKonto []byte // size 40
		var Belopp []byte    // BCD / Decimal Precision 19
		var Datum []byte     // size 10
		var HurOfta []byte   // size 15
		var Vad []byte       // size 40
		var Vem []byte       // size 40
		var Löpnr []byte     // Autoinc Primary Key, index
		var Kontrollnr int   // Integer
		var TillDatum []byte // size 10
		var Rakning []byte   // size 1

		_, _ = fmt.Fprintf(w, "<table style=\"width:100%%\"><tr><th>Löpnr</th><th>Frånkonto</th><th>Tillkonto/Plats</th><th>Belopp</th><th>Datum</th><th>Hur Ofta</th><th>Vad</th><th>Vem</th><th>Kontrollnr</th><th>Till datum</th><th>Räkning</th><th>Agera</th>\n")
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
			sqlStmt += "<form method=\"POST\" action=\"/fixedtrans\">\n"
			sqlStmt += "<input type=\"hidden\" id=\"transtyp\" name=\"transtyp\" value=\"FastTrans\">\n"
			sqlStmt += "<input type=\"hidden\" id=\"transid\" name=\"transid\" value=\"" + toUtf8(Löpnr) + "\">\n"
			sqlStmt += "<input type=\"submit\" value=\"Registrera\"></form>\n"
			sqlStmt += "</td>"

			sqlStmt += "</tr>\n"
			_, _ = fmt.Fprintf(w, "%s", sqlStmt)
		}
		_, _ = fmt.Fprintf(w, "</table>\n")
	}
}

func addfixedtransaction(w http.ResponseWriter, req *http.Request, db *sql.DB) {
	log.Println("addfixedtransaction start")
	err := req.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	transtyp := req.FormValue("transtyp")
	date := req.FormValue("date")
	who := req.FormValue("who")
	amount := req.FormValue("amount")
	text := req.FormValue("text")
	log.Println("Val: ", transtyp)
	log.Println("Val: ", date)
	log.Println("Val: ", who)
	log.Println("Val: ", amount)
	log.Println("Val: ", text)

	if transtyp == "FastTrans" {
		transid := req.FormValue("transid")
		transidnum, _ := strconv.Atoi(transid)
		registreraFastTransaktionHTML(w, transidnum, db)
		_, _ = fmt.Fprintf(w, "<p>\n")
	}
	log.Println("addfixedtransaction slut")
}

func registreraFastTransaktion(db *sql.DB, transid int) {
	if db == nil {
		log.Println("registreraFastTransaktion: No database open")
		return
	}

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
	var Belopp []byte     // BCD / Decimal Precision 19
	var Datum []byte      // size 10
	var HurOfta []byte    // size 15
	var Vad []byte        // size 40
	var Vem []byte        // size 40
	var Löpnr []byte      // Autoinc Primary Key, index
	var Kontrollnr []byte // Integer
	var TillDatum []byte  // size 10
	var Rakning []byte    // size 1

	res.Next()
	err = res.Scan(&FrånKonto, &TillKonto, &Belopp, &Datum, &HurOfta, &Vad, &Vem, &Löpnr, &Kontrollnr, &TillDatum, &Rakning)
	if err != nil {
		log.Println("registreraFastTransaktion: SCAN ERROR")
		log.Println(err)
		log.Println("registreraFastTransaktion: Bail out")
		return
	}

	amountstr := SanitizeAmountb(Belopp)
	amount, err := decimal.NewFromString(amountstr)
	if err != nil {
		log.Println("OK: registreraFastTransaktion, trasig/saknar amount ", amountstr, err)
	}
	
	sqlStmt := ""
	sqlStmt += "<tr><td>" + toUtf8(Löpnr) + "</td>"
	sqlStmt += "<td>" + toUtf8(FrånKonto) + "</td>"
	sqlStmt += "<td>" + toUtf8(TillKonto) + "</td>"
	sqlStmt += "<td>" + amount.String() + "</td>"
	sqlStmt += "<td>" + toUtf8(Datum) + "</td>"
	sqlStmt += "<td>" + toUtf8(HurOfta) + "</td>"
	sqlStmt += "<td>" + toUtf8(Vad) + "</td>"
	sqlStmt += "<td>" + toUtf8(Vem) + "</td>"
	sqlStmt += "<td>" + toUtf8(Kontrollnr) + "</td>"
	sqlStmt += "<td>" + toUtf8(TillDatum) + "</td>"
	sqlStmt += "<td>" + toUtf8(Rakning) + "</td>"
	sqlStmt += "</tr>\n"

	_ = res.Close()
	// Register transaction
	if toUtf8(Vad) == "---" {
		// Fasta överföringar
		addTransaktionÖverföring(toUtf8(FrånKonto),
			toUtf8(TillKonto),
			toUtf8(Datum),
			toUtf8(Vem),
			amount,
			"Fast transaktion wHHEK")
	} else if toUtf8(FrånKonto) == "---" {
		// Fasta inkomster
		addTransaktionInsättning(toUtf8(TillKonto),
			toUtf8(Datum),
			toUtf8(Vad),
			toUtf8(Vem),
			amount,
			"Fast transaktion wHHEK")
	} else {
		// Fasta utgifter
		addTransaktionInköp(toUtf8(FrånKonto),
			toUtf8(TillKonto),
			toUtf8(Datum),
			toUtf8(Vad),
			toUtf8(Vem),
			amount,
			"Fast transaktion wHHEK",
			true)
	}

	// Update repeating transaction
	var newDatum string
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
		log.Fatal("Okänd periodicitet: " + toUtf8(HurOfta))
	}
	sqlStatement := `UPDATE Överföringar SET Datum = ? WHERE Löpnr = ?`
	_, err = db.Exec(sqlStatement, newDatum, transid)
	if err != nil {
		panic(err)
	}
}

func registreraFastTransaktionHTML(w http.ResponseWriter, transid int, db *sql.DB) {
	log.Println("registreraFastTransaktionHTML start")
	_, _ = fmt.Fprintf(w, "Läser ut fast transaktion#"+strconv.Itoa(transid))
	if db == nil {
		_, _ = fmt.Fprintf(w, "registreraFastTransaktion: No database open<p>\n")
		return
	}
	_, _ = fmt.Fprintf(w, "<p>\n")
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
	var Belopp []byte     // BCD / Decimal Precision 19
	var Datum []byte      // size 10
	var HurOfta []byte    // size 15
	var Vad []byte        // size 40
	var Vem []byte        // size 40
	var Löpnr []byte      // Autoinc Primary Key, index
	var Kontrollnr []byte // Integer
	var TillDatum []byte  // size 10
	var Rakning []byte    // size 1

	_, _ = fmt.Fprintf(w, "<table style=\"width:100%%\"><tr><th>Löpnr</th><th>Frånkonto</th><th>Tillkonto/Plats</th><th>Belopp</th><th>Datum</th><th>Hur Ofta</th><th>Vad</th><th>Vem</th><th>Kontrollnr</th><th>Till datum</th><th>Räkning</th><th>Agera</th>\n")
	res.Next()
	err = res.Scan(&FrånKonto, &TillKonto, &Belopp, &Datum, &HurOfta, &Vad, &Vem, &Löpnr, &Kontrollnr, &TillDatum, &Rakning)
	if err != nil {
		log.Println("registreraFastTransaktion: SCAN ERROR")
		log.Println(err)
		log.Println("registreraFastTransaktion: Bail out")
		_, _ = fmt.Fprintf(w, "<tr>Bail out</tr>\n")
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
	_, _ = fmt.Fprintf(w, "%s", sqlStmt)
	_, _ = fmt.Fprintf(w, "</table>\n")

	amountstr := SanitizeAmountb(Belopp)
	amount, err := decimal.NewFromString(amountstr)
	if err != nil {
		log.Println("OK: registreraFastTransaktionHTML, trasig/saknar amount ", amountstr, err)
	}

	// Register transaction
	if toUtf8(Vad) == "---" {
		// Fasta överföringar
		_, _ = fmt.Fprintf(w, "Registrerar Överföring...<br> ")
		log.Println("registreraFastTransaktionHTML Registrerar Överföring...")

		addTransaktionÖverföring(toUtf8(FrånKonto), toUtf8(TillKonto), toUtf8(Datum), toUtf8(Vem), amount, "Fast transaktion wHHEK")

		_, _ = fmt.Fprintf(w, "<table style=\"width:100%%\"><tr><th>Frånkonto</th><th>Tillkonto</th><th>Typ</th><th>Datum</th><th>Vem</th><th>Belopp</th><th>Text</th>\n")
		sqlStmt := "<tr>"
		sqlStmt += "<td>" + toUtf8(FrånKonto) + "</td>"
		sqlStmt += "<td>" + toUtf8(TillKonto) + "</td>"
		sqlStmt += "<td>" + "Överföring" + "</td>"
		sqlStmt += "<td>" + toUtf8(Datum) + "</td>"
		sqlStmt += "<td>" + toUtf8(Vem) + "</td>"
		sqlStmt += "<td>" + toUtf8(Belopp) + "</td>"
		sqlStmt += "</tr>"
		_, _ = fmt.Fprintf(w, "%s", sqlStmt)
		_, _ = fmt.Fprintf(w, "</table>\n")
	} else if toUtf8(FrånKonto) == "---" {
		// Fasta inkomster
		_, _ = fmt.Fprintf(w, "Registrerar Insättning...<br> ")
		log.Println("registreraFastTransaktionHTML Insättning...")

		addTransaktionInsättning(toUtf8(TillKonto), toUtf8(Datum), toUtf8(Vad), toUtf8(Vem), amount, "Fast transaktion wHHEK")

		_, _ = fmt.Fprintf(w, "<table style=\"width:100%%\"><tr><th>Konto</th><th>Typ</th><th>Vad</th><th>Datum</th><th>Vem</th><th>Belopp</th><th>Text</th>\n")
		sqlStmt := "<tr>"
		sqlStmt += "<td>" + toUtf8(TillKonto) + "</td>"
		sqlStmt += "<td>" + "Insättning" + "</td>"
		sqlStmt += "<td>" + toUtf8(Vad) + "</td>"
		sqlStmt += "<td>" + toUtf8(Datum) + "</td>"
		sqlStmt += "<td>" + toUtf8(Vem) + "</td>"
		sqlStmt += "<td>" + toUtf8(Belopp) + "</td>"
		sqlStmt += "</tr>"
		_, _ = fmt.Fprintf(w, "%s", sqlStmt)
		_, _ = fmt.Fprintf(w, "</table>\n")
	} else {
		// Fasta utgifter
		_, _ = fmt.Fprintf(w, "Registrerar Fast Utgift...<br> ")
		log.Println("registreraFastTransaktionHTML Fast Utgift...")

		addTransaktionInköp(toUtf8(FrånKonto), toUtf8(TillKonto), toUtf8(Datum), toUtf8(Vad), toUtf8(Vem), amount, "Fast transaktion wHHEK", true)

		_, _ = fmt.Fprintf(w, "<table style=\"width:100%%\"><tr><th>Frånkonto</th><th>Plats</th><th>Typ</th><th>Vad</th><th>Datum</th><th>Vem</th><th>Belopp</th><th>Text</th>\n")
		sqlStmt := "<tr>"
		sqlStmt += "<td>" + toUtf8(FrånKonto) + "</td>"
		sqlStmt += "<td>" + toUtf8(TillKonto) + "</td>"
		sqlStmt += "<td>" + "Fast Utgift" + "</td>"
		sqlStmt += "<td>" + toUtf8(Vad) + "</td>"
		sqlStmt += "<td>" + toUtf8(Datum) + "</td>"
		sqlStmt += "<td>" + toUtf8(Vem) + "</td>"
		sqlStmt += "<td>" + toUtf8(Belopp) + "</td>"
		sqlStmt += "</tr>"
		_, _ = fmt.Fprintf(w, "%s", sqlStmt)
		_, _ = fmt.Fprintf(w, "</table>\n")
	}

	log.Println("registreraFastTransaktionHTML Update date...")
	log.Println("registreraFastTransaktionHTML switch " + toUtf8(HurOfta))
	// Update repeating transaction
	var newDatum string
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
		log.Fatal("Okänd periodicitet: " + toUtf8(HurOfta))

	}
	log.Println("registreraFastTransaktionHTML nytt datum " + newDatum)
	sqlStatement := `UPDATE Överföringar SET Datum = ? WHERE Löpnr = ?`
	_, err = db.Exec(sqlStatement, newDatum, transid)
	if err != nil {
		log.Println("registreraFastTransaktionHTML Update error: ", err)
		panic(err)
	}

	_, _ = fmt.Fprintf(w, "<p>\n")
	log.Println("registreraFastTransaktionHTML slut")
}

func fixedtransactionHTML(w http.ResponseWriter, req *http.Request) {
	log.Println("fixedtransactionHTML start")
	_, _ = fmt.Fprintf(w, "<html>\n")
	_, _ = fmt.Fprintf(w, "<head>\n")
	_, _ = fmt.Fprintf(w, "<style>\n")
	_, _ = fmt.Fprintf(w, "table,th,td { border: 1px solid black }\n")
	_, _ = fmt.Fprintf(w, "</style>\n")
	_, _ = fmt.Fprintf(w, "</head>\n")
	_, _ = fmt.Fprintf(w, "<body>\n")
	_, _ = fmt.Fprintf(w, "<h1>%s</h1>\n", currentDatabase)

	addfixedtransaction(w, req, db)

	showFastaTransaktioner(w, db)

	_, _ = fmt.Fprintf(w, "<a href=\"summary\">Översikt</a>\n")
	_, _ = fmt.Fprintf(w, "</body>\n")
	_, _ = fmt.Fprintf(w, "</html>\n")
	log.Println("fixedtransactionHTML slut")
}

func antalFastaTransaktioner(db *sql.DB) int {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	res1 := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM Överföringar`)

	var antal int

	err := res1.Scan(&antal)
	if err != nil {
		log.Fatal(err)
	}

	return antal
}

func skapaFastUtgift(db *sql.DB, vad string, konto string, vem string, plats string, summa decimal.Decimal, datum string, hurofta string) error {
	if db == nil {
		log.Fatal("skapaFastUtgift anropad med db=nil")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := db.ExecContext(ctx,
		`INSERT INTO Överföringar (FrånKonto,TillKonto,Belopp,Datum,HurOfta, Vad, Vem, Kontrollnr, TillDatum, Rakning) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, konto, plats, summa, datum, hurofta, vad, vem, 1, "---", false)

	if err != nil {
		log.Fatal(err)
	}

	return err
}

func hämtaFastTransaktion(db *sql.DB, lopnr int) (result fixedtransaction) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var err error
	var res *sql.Rows

	res, err = db.QueryContext(ctx, `SELECT FrånKonto,TillKonto,Belopp,Datum,HurOfta,Vad,Vem,Löpnr,Kontrollnr,TillDatum,Rakning FROM Överföringar
  where Löpnr = ?`, lopnr)
	if err != nil {
		log.Fatal(err)
	}

	var FrånKonto []byte // size 40
	var TillKonto []byte // size 40
	var Belopp []byte    // BCD / Decimal Precision 19
	var Datum []byte     // size 10
	var HurOfta []byte   // size 15
	var Vad []byte       // size 40
	var Vem []byte       // size 40
	var Löpnr []byte     // Autoinc Primary Key, index
	var Kontrollnr int   // Integer
	var TillDatum []byte // size 10
	var Rakning []byte   // size 1

	for res.Next() {
		var record fixedtransaction
		err = res.Scan(&FrånKonto, &TillKonto, &Belopp, &Datum, &HurOfta, &Vad, &Vem, &Löpnr, &Kontrollnr, &TillDatum, &Rakning)

		record.lopnr, err = strconv.Atoi(toUtf8(Löpnr))
		record.vernum = Kontrollnr
		record.fromAcc = toUtf8(FrånKonto)
		record.toAcc = toUtf8(TillKonto)
		record.what = toUtf8(Vad)
		record.date, err = time.Parse("2006-01-02", toUtf8(Datum))
		record.who = toUtf8(Vem)
		record.amount, err = decimal.NewFromString(toUtf8(Belopp))
		record.HurOfta = toUtf8(HurOfta)
		record.rakning, err = strconv.ParseBool(toUtf8(Rakning))
		if toUtf8(TillDatum) == "---" {
			//record.todate = nil
		} else {
			record.todate, err = time.Parse("2006-01-02", toUtf8(TillDatum))
		}
		result = record
	}
	return result
}
