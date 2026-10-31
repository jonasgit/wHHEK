//-*- coding: utf-8 -*-

package main

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
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
	rakning bool
}

func CurrDate() string {
	now := time.Now()
	currDate := now.Format("2006-01-02")
	return currDate
}

// parseFixedDate tolkar ett datum som "2021-01-31" eller "2021-1-31".
func parseFixedDate(datum string) (time.Time, error) {
	datum = strings.TrimSpace(datum)
	for _, layout := range []string{"2006-01-02", "2006-1-2"} {
		if t, err := time.Parse(layout, datum); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("ogiltigt datum '%s' (ska vara ÅÅÅÅ-MM-DD)", datum)
}

// IncrDate räknar fram datumet veckor/månader efter datum. Hamnar en månadsökning
// på en dag som inte finns (t.ex. 31 feb) blir det sista dagen i månaden.
func IncrDate(datum string, veckor int, månader int) (string, error) {
	if månader != 0 && veckor != 0 {
		return "", fmt.Errorf("inte tillåtet med både veckor och månader")
	}
	t, err := parseFixedDate(datum)
	if err != nil {
		return "", err
	}
	day := t.Day()
	t = time.Date(t.Year(), t.Month(), day, 12, 0, 0, 0, time.FixedZone("CET", 0))
	nytt := t.AddDate(0, månader, veckor*7)
	//fix date at end of month spilling over to next month
	if månader != 0 && nytt.Day() != day {
		nytt = BeginningOfMonth(nytt)
		nytt = nytt.AddDate(0, 0, -1)
	}
	return nytt.Format("2006-01-02"), nil
}

// fixedPeriods är de periodiciteter som finns för fasta transaktioner, som
// veckor och månader. Både text och Hogias sifferkoder förekommer i databasen.
var fixedPeriods = []struct {
	names           []string
	veckor, månader int
}{
	{[]string{"Varannan vecka"}, 2, 0},
	{[]string{"Varje månad", "1"}, 0, 1},
	{[]string{"Varannan månad", "2"}, 0, 2},
	{[]string{"Varje kvartal", "3"}, 0, 3},
	{[]string{"Varje halvår", "6"}, 0, 6},
	{[]string{"Varje år", "12"}, 0, 12},
}

// normalizePeriod känner igen en periodicitet oavsett skiftläge och mellanslag
// runt omkring och returnerar den som den ska sparas, t.ex. "varje månad" ->
// "Varje månad". Sifferkoderna behålls som de är.
func normalizePeriod(hurofta string) (string, int, int, error) {
	s := strings.TrimSpace(hurofta)
	for _, p := range fixedPeriods {
		for _, name := range p.names {
			if strings.EqualFold(s, name) {
				return name, p.veckor, p.månader, nil
			}
		}
	}
	return "", 0, 0, fmt.Errorf("okänd periodicitet '%s' (giltiga: Varannan vecka, Varje månad, "+
		"Varannan månad, Varje kvartal, Varje halvår, Varje år)", hurofta)
}

// nextFixedDate räknar fram nästa datum för en fast transaktion.
func nextFixedDate(datum string, hurofta string) (string, error) {
	_, veckor, månader, err := normalizePeriod(hurofta)
	if err != nil {
		return "", err
	}
	return IncrDate(datum, veckor, månader)
}

// normalizeFixed kontrollerar och snyggar till datum och periodicitet innan en
// fast transaktion sparas, så att den går att registrera senare.
func normalizeFixed(datum string, hurofta string) (string, string, error) {
	t, err := parseFixedDate(datum)
	if err != nil {
		return "", "", err
	}
	period, _, _, err := normalizePeriod(hurofta)
	if err != nil {
		return "", "", err
	}
	return t.Format("2006-01-02"), period, nil
}

func BeginningOfMonth(date time.Time) time.Time {
	return date.AddDate(0, 0, -date.Day()+1)
}

func EndOfMonth(date time.Time) time.Time {
	return date.AddDate(0, 1, -date.Day())
}

type transType struct {
	Löpnr      string
	FranKonto  string
	TillKonto  string
	Belopp     string
	Datum      string
	HurOfta    string
	Vad        string
	Vem        string
	Kontrollnr string
	TillDatum  string
	Rakning    string
}

//go:embed html/fasta1.html
var htmlfasta1 string

type Fasta1Data struct {
	Antal         string
	TypeEdit      bool
	Transaktioner []transType
}

func showFastaTransaktioner(w http.ResponseWriter, db *sql.DB, showedit bool) {
	now := time.Now()
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()

	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)
	currDate := lastOfMonth.Format("2006-01-02")
	antal := GetCountPendingÖverföringar(db, currDate)
	if antal > 0 || showedit {
		var res *sql.Rows
		var err error
		if showedit {
			res, err = db.Query("SELECT FrånKonto,TillKonto,Belopp,Datum,HurOfta,Vad,Vem,Löpnr,Kontrollnr,TillDatum,Rakning FROM Överföringar ")
		} else {
			res, err = db.Query("SELECT FrånKonto,TillKonto,Belopp,Datum,HurOfta,Vad,Vem,Löpnr,Kontrollnr,TillDatum,Rakning FROM Överföringar WHERE Datum <= ?", currDate)
		}

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

		var transaktioner []transType
		var transaktion transType

		for res.Next() {
			_ = res.Scan(&FrånKonto, &TillKonto, &Belopp, &Datum, &HurOfta, &Vad, &Vem, &Löpnr, &Kontrollnr, &TillDatum, &Rakning)

			transaktion.Löpnr = toUtf8(Löpnr)
			transaktion.FranKonto = toUtf8(FrånKonto)
			transaktion.TillKonto = toUtf8(TillKonto)
			transaktion.Belopp = toUtf8(Belopp)
			transaktion.Datum = toUtf8(Datum)
			transaktion.HurOfta = toUtf8(HurOfta)
			transaktion.Vad = toUtf8(Vad)
			transaktion.Vem = toUtf8(Vem)
			transaktion.Kontrollnr = strconv.Itoa(Kontrollnr)
			transaktion.TillDatum = toUtf8(TillDatum)
			transaktion.Rakning = toUtf8(Rakning)
			transaktioner = append(transaktioner, transaktion)
		}

		res.Close()

		tmpl1 := template.New("wHHEK Fasta")
		tmpl1, err = tmpl1.Parse(htmlfasta1)
		if err != nil {
			log.Fatal(err)
		}
		data := Fasta1Data{
			Antal:         strconv.Itoa(antal),
			TypeEdit:      showedit,
			Transaktioner: transaktioner,
		}
		_ = tmpl1.Execute(w, data)
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
	log.Println("Val: ", transtyp, getCurrentFuncName())
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

// registreraFastTransaktion registrerar en fast transaktion och flyttar fram dess
// datum. Är datum eller periodicitet ogiltiga registreras ingenting.
func registreraFastTransaktion(db *sql.DB, transid int) error {
	if db == nil {
		log.Println("registreraFastTransaktion: No database open")
		return fmt.Errorf("ingen databas öppen")
	}

	// Retrieve repeating transaction
	var res *sql.Rows
	res, err := db.Query("SELECT FrånKonto,TillKonto,Belopp,Datum,HurOfta,Vad,Vem,Löpnr,Kontrollnr,TillDatum,Rakning FROM Överföringar WHERE Löpnr = ?", transid)
	if err != nil {
		log.Println("registreraFastTransaktion: SELECT ERROR")
		log.Println(err)
		return err
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
		_ = res.Close()
		return err
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

	// Next date first: if date or period is invalid nothing is registered.
	newDatum, err := nextFixedDate(toUtf8(Datum), toUtf8(HurOfta))
	if err != nil {
		log.Println("registreraFastTransaktion:", err)
		return err
	}

	// Register with the date in standard form (ÅÅÅÅ-MM-DD), also for old rows
	// saved as e.g. 2021-1-31, so the transactions sort correctly.
	regDate, _ := parseFixedDate(toUtf8(Datum))
	regDatum := regDate.Format("2006-01-02")

	// Register transaction
	if toUtf8(Vad) == "---" {
		// Fasta överföringar
		addTransaktionÖverföring(db, toUtf8(FrånKonto),
			toUtf8(TillKonto),
			regDatum,
			toUtf8(Vem),
			amount,
			"Fast transaktion wHHEK")
	} else if toUtf8(FrånKonto) == "---" {
		// Fasta inkomster
		addTransaktionInsättning(db, toUtf8(TillKonto),
			regDatum,
			toUtf8(Vad),
			toUtf8(Vem),
			amount,
			"Fast transaktion wHHEK")
	} else {
		// Fasta utgifter
		addTransaktionInköp(db, toUtf8(FrånKonto),
			toUtf8(TillKonto),
			regDatum,
			toUtf8(Vad),
			toUtf8(Vem),
			amount,
			"Fast transaktion wHHEK",
			true)
	}

	// Update repeating transaction
	sqlStatement := `UPDATE Överföringar SET Datum = ? WHERE Löpnr = ?`
	_, err = db.Exec(sqlStatement, newDatum, transid)
	return err
}

func registreraFastTransaktionHTML(w http.ResponseWriter, transid int, db *sql.DB) {
	log.Println("registreraFastTransaktionHTML start")
	_, _ = fmt.Fprintf(w, "%s", "Läser ut fast transaktion#"+strconv.Itoa(transid))
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

	res.Close()

	// Next date first: if date or period is invalid nothing is registered.
	newDatum, err := nextFixedDate(toUtf8(Datum), toUtf8(HurOfta))
	if err != nil {
		log.Println("registreraFastTransaktionHTML:", err)
		_, _ = fmt.Fprintf(w, "Kan inte registrera: %s. Rätta den fasta transaktionen först.<p>\n",
			template.HTMLEscapeString(err.Error()))
		return
	}

	// Register with the date in standard form (ÅÅÅÅ-MM-DD), also for old rows
	// saved as e.g. 2021-1-31, so the transactions sort correctly.
	regDate, _ := parseFixedDate(toUtf8(Datum))
	regDatum := regDate.Format("2006-01-02")

	// Register transaction
	if toUtf8(Vad) == "---" {
		// Fasta överföringar
		_, _ = fmt.Fprintf(w, "Registrerar Överföring...<br> ")
		log.Println("registreraFastTransaktionHTML Registrerar Överföring...")

		addTransaktionÖverföring(db, toUtf8(FrånKonto), toUtf8(TillKonto), regDatum, toUtf8(Vem), amount, "Fast transaktion wHHEK")

		_, _ = fmt.Fprintf(w, "<table style=\"width:100%%\"><tr><th>Frånkonto</th><th>Tillkonto</th><th>Typ</th><th>Datum</th><th>Vem</th><th>Belopp</th><th>Text</th>\n")
		sqlStmt := "<tr>"
		sqlStmt += "<td>" + toUtf8(FrånKonto) + "</td>"
		sqlStmt += "<td>" + toUtf8(TillKonto) + "</td>"
		sqlStmt += "<td>" + "Överföring" + "</td>"
		sqlStmt += "<td>" + regDatum + "</td>"
		sqlStmt += "<td>" + toUtf8(Vem) + "</td>"
		sqlStmt += "<td>" + toUtf8(Belopp) + "</td>"
		sqlStmt += "</tr>"
		_, _ = fmt.Fprintf(w, "%s", sqlStmt)
		_, _ = fmt.Fprintf(w, "</table>\n")
	} else if toUtf8(FrånKonto) == "---" {
		// Fasta inkomster
		_, _ = fmt.Fprintf(w, "Registrerar Insättning...<br> ")
		log.Println("registreraFastTransaktionHTML Insättning...")

		addTransaktionInsättning(db, toUtf8(TillKonto), regDatum, toUtf8(Vad), toUtf8(Vem), amount, "Fast transaktion wHHEK")

		_, _ = fmt.Fprintf(w, "<table style=\"width:100%%\"><tr><th>Konto</th><th>Typ</th><th>Vad</th><th>Datum</th><th>Vem</th><th>Belopp</th><th>Text</th>\n")
		sqlStmt := "<tr>"
		sqlStmt += "<td>" + toUtf8(TillKonto) + "</td>"
		sqlStmt += "<td>" + "Insättning" + "</td>"
		sqlStmt += "<td>" + toUtf8(Vad) + "</td>"
		sqlStmt += "<td>" + regDatum + "</td>"
		sqlStmt += "<td>" + toUtf8(Vem) + "</td>"
		sqlStmt += "<td>" + toUtf8(Belopp) + "</td>"
		sqlStmt += "</tr>"
		_, _ = fmt.Fprintf(w, "%s", sqlStmt)
		_, _ = fmt.Fprintf(w, "</table>\n")
	} else {
		// Fasta utgifter
		_, _ = fmt.Fprintf(w, "Registrerar Fast Utgift...<br> ")
		log.Println("registreraFastTransaktionHTML Fast Utgift...")

		addTransaktionInköp(db, toUtf8(FrånKonto), toUtf8(TillKonto), regDatum, toUtf8(Vad), toUtf8(Vem), amount, "Fast transaktion wHHEK", true)

		_, _ = fmt.Fprintf(w, "<table style=\"width:100%%\"><tr><th>Frånkonto</th><th>Plats</th><th>Typ</th><th>Vad</th><th>Datum</th><th>Vem</th><th>Belopp</th><th>Text</th>\n")
		sqlStmt := "<tr>"
		sqlStmt += "<td>" + toUtf8(FrånKonto) + "</td>"
		sqlStmt += "<td>" + toUtf8(TillKonto) + "</td>"
		sqlStmt += "<td>" + "Fast Utgift" + "</td>"
		sqlStmt += "<td>" + toUtf8(Vad) + "</td>"
		sqlStmt += "<td>" + regDatum + "</td>"
		sqlStmt += "<td>" + toUtf8(Vem) + "</td>"
		sqlStmt += "<td>" + toUtf8(Belopp) + "</td>"
		sqlStmt += "</tr>"
		_, _ = fmt.Fprintf(w, "%s", sqlStmt)
		_, _ = fmt.Fprintf(w, "</table>\n")
	}

	log.Println("registreraFastTransaktionHTML nytt datum " + newDatum)
	sqlStatement := `UPDATE Överföringar SET Datum = ? WHERE Löpnr = ?`
	_, err = db.Exec(sqlStatement, newDatum, transid)
	if err != nil {
		log.Println("registreraFastTransaktionHTML Update error: ", err)
		_, _ = fmt.Fprintf(w, "Kunde inte flytta fram datumet: %s<p>\n", template.HTMLEscapeString(err.Error()))
		return
	}

	_, _ = fmt.Fprintf(w, "<p>\n")
	log.Println("registreraFastTransaktionHTML slut")
}

//go:embed html/fasta4.html
var htmlfasta4 string

type Fasta4Data struct {
	CurrDBName string
	CurrDate   string
}

func fixedtransactionHTML(w http.ResponseWriter, req *http.Request) {
	log.Println("fixedtransactionHTML start")
	currentTime := time.Now()
	currDate := currentTime.Format("2006-01-02")

	t := template.New("Fasta4")
	t, _ = t.Parse(htmlfasta4)
	data := Fasta4Data{
		CurrDBName: currentDatabase,
		CurrDate:   currDate,
	}
	_ = t.Execute(w, data)

	addfixedtransaction(w, req, db)

	showFastaTransaktioner(w, db, false)

	_, _ = fmt.Fprintf(w, "<a href=\"summary\">Översikt</a>\n")
	_, _ = fmt.Fprintf(w, "</body>\n")
	_, _ = fmt.Fprintf(w, "</html>\n")
	log.Println("fixedtransactionHTML slut")
}

func antalFastaTransaktioner(db *sql.DB) int {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var antal int

	err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM Överföringar`).Scan(&antal)
	if err != nil {
		log.Fatal(err)
	}

	return antal
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
		_ = res.Scan(&FrånKonto, &TillKonto, &Belopp, &Datum, &HurOfta, &Vad, &Vem, &Löpnr, &Kontrollnr, &TillDatum, &Rakning)

		record.lopnr, _ = strconv.Atoi(toUtf8(Löpnr))
		record.vernum = Kontrollnr
		record.fromAcc = toUtf8(FrånKonto)
		record.toAcc = toUtf8(TillKonto)
		record.what = toUtf8(Vad)
		record.date, _ = time.Parse("2006-01-02", toUtf8(Datum))
		record.who = toUtf8(Vem)
		record.amount, _ = decimal.NewFromString(toUtf8(Belopp))
		record.HurOfta = toUtf8(HurOfta)
		record.rakning, _ = strconv.ParseBool(toUtf8(Rakning))
		if toUtf8(TillDatum) == "---" {
			//record.todate = nil
		} else {
			record.todate, _ = time.Parse("2006-01-02", toUtf8(TillDatum))
		}
		result = record
	}
	return result
}

//go:embed html/fasta3.html
var htmlfasta3 string

type Fasta3Data struct {
	CurrDBName string
	CurrDate   string
}

func editfixedtransactionHTML(w http.ResponseWriter, req *http.Request) {
	log.Println("editfixedtransactionHTML start")
	currentTime := time.Now()
	currDate := currentTime.Format("2006-01-02")

	t := template.New("Fasta3")
	t, _ = t.Parse(htmlfasta3)
	data := Fasta3Data{
		CurrDBName: currentDatabase,
		CurrDate:   currDate,
	}
	_ = t.Execute(w, data)

	editfixedtransaction(w, req, db)

	showFastaTransaktioner(w, db, true)

	_, _ = fmt.Fprintf(w, "<a href=\"summary\">Översikt</a>\n")
	_, _ = fmt.Fprintf(w, "</body>\n")
	_, _ = fmt.Fprintf(w, "</html>\n")
	log.Println("fixedtransactionHTML slut")
}

//go:embed html/fasta2.html
var htmlfasta2 string

type Fasta2Data struct {
	Transaktion     transType
	KontonamnLista  []string
	PlatserLista    []string
	PersonerLista   []string
	VadInkomstLista []string
	VadUtgiftLista  []string
}

func editfixedtransaction(w http.ResponseWriter, req *http.Request, db *sql.DB) {
	log.Println("editfixedtransaction start")
	err := req.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	transtyp := req.FormValue("transtyp")
	date := req.FormValue("date")
	who := req.FormValue("who")
	amount := req.FormValue("amount")
	text := req.FormValue("text")
	log.Println("Val: ", transtyp, getCurrentFuncName())
	log.Println("Val: ", date)
	log.Println("Val: ", who)
	log.Println("Val: ", amount)
	log.Println("Val: ", text)

	if transtyp == "FastTrans" {
		transid := req.FormValue("transid")
		transidnum, _ := strconv.Atoi(transid)

		var err error

		var FrånKonto []byte  // size 40
		var TillKonto []byte  // size 40
		var Belopp []byte     // BCD / Decimal Precision 19
		var Datum []byte      // size 10
		var HurOfta []byte    // size 15
		var Vad []byte        // size 40
		var Vem []byte        // size 40
		var Löpnr []byte      // Autoinc Primary Key, index
		var Kontrollnr []byte // Borde vara?Integer
		var TillDatum []byte  // size 10
		var Rakning []byte    // size 1

		var transaktion transType

		err = db.QueryRow("SELECT FrånKonto,TillKonto,Belopp,Datum,HurOfta,Vad,Vem,Löpnr,Kontrollnr,TillDatum,Rakning FROM Överföringar WHERE Löpnr = ?", transidnum).Scan(&FrånKonto, &TillKonto, &Belopp, &Datum, &HurOfta, &Vad, &Vem, &Löpnr, &Kontrollnr, &TillDatum, &Rakning)

		if err != nil {
			log.Fatal(err)
		}

		transaktion.Löpnr = toUtf8(Löpnr)
		transaktion.FranKonto = toUtf8(FrånKonto)
		transaktion.TillKonto = toUtf8(TillKonto)
		transaktion.Belopp = toUtf8(Belopp)
		transaktion.Datum = toUtf8(Datum)
		transaktion.HurOfta = toUtf8(HurOfta)
		transaktion.Vad = toUtf8(Vad)
		transaktion.Vem = toUtf8(Vem)
		transaktion.Kontrollnr = toUtf8(Kontrollnr)
		transaktion.TillDatum = toUtf8(TillDatum)
		transaktion.Rakning = toUtf8(Rakning)

		kontonamnlista := getAccNames()
		platserlista := getPlaceNames()
		personerlista := getPersonNames()
		vadinkomstlista := getTypeInNames()
		vadutgiftlista := getTypeOutNames()

		t := template.New("EditFixed2")
		t, _ = t.Parse(htmlfasta2)
		data := Fasta2Data{
			Transaktion:     transaktion,
			KontonamnLista:  kontonamnlista,
			PlatserLista:    platserlista,
			PersonerLista:   personerlista,
			VadInkomstLista: vadinkomstlista,
			VadUtgiftLista:  vadutgiftlista,
		}
		err = t.Execute(w, data)
		if err != nil {
			return
		}
	}
	log.Println("editfixedtransaction slut")
}
