//-*- coding: utf-8 -*-

package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal" // MIT License
)

type konto struct {
	KontoNummer    string          // size 20
	Benämning      string          // size 40, index
	Saldo          decimal.Decimal // BCD / Decimal Precision 19
	StartSaldo     decimal.Decimal // BCD / Decimal Precision 19
	StartManad     string          // size 10
	SaldoArsskifte string          // BCD / Decimal Precision 19
	ArsskifteManad string          // size 10
}

func printKonton(w http.ResponseWriter, db *sql.DB) {
	res, err := db.Query("SELECT KontoNummer,Benämning,Saldo,StartSaldo,StartManad,Löpnr,SaldoArsskifte,ArsskifteManad FROM Konton")

	if err != nil {
		log.Fatal(err)
	}

	var KontoNummer []byte    // size 20
	var Benämning []byte      // size 40, index
	var Saldo []byte          // BCD / Decimal Precision 19
	var StartSaldo []byte     // BCD / Decimal Precision 19
	var StartManad []byte     // size 10
	var Löpnr int             // autoinc Primary Key
	var SaldoArsskifte []byte // BCD / Decimal Precision 19
	var ArsskifteManad []byte // size 10

	_, _ = fmt.Fprintf(w, "<table style=\"width:100%%\"><tr>")
	_, _ = fmt.Fprintf(w, "<th>Kontonummer</th>")
	_, _ = fmt.Fprintf(w, "<th>Benämning</th>")
	_, _ = fmt.Fprintf(w, "<th>Saldo</th>")
	_, _ = fmt.Fprintf(w, "<th>Startsaldo</th>")
	_, _ = fmt.Fprintf(w, "<th>Startmånad</th>")
	_, _ = fmt.Fprintf(w, "<th>Saldo årsskifte</th>")
	_, _ = fmt.Fprintf(w, "<th>Årsskiftesmånad</th>")
	_, _ = fmt.Fprintf(w, "<th>Redigera</th><th>Radera</th>\n")
	for res.Next() {
		_ = res.Scan(&KontoNummer, &Benämning, &Saldo, &StartSaldo, &StartManad, &Löpnr, &SaldoArsskifte, &ArsskifteManad)

		_, _ = fmt.Fprintf(w, "<tr>")
		_, _ = fmt.Fprintf(w, "<td>%s</td>", toUtf8(KontoNummer))
		_, _ = fmt.Fprintf(w, "<td>%s</td>", toUtf8(Benämning))
		_, _ = fmt.Fprintf(w, "<td>%s</td>", toUtf8(Saldo))
		_, _ = fmt.Fprintf(w, "<td>%s</td>", toUtf8(StartSaldo))
		_, _ = fmt.Fprintf(w, "<td>%s</td>", toUtf8(StartManad))
		_, _ = fmt.Fprintf(w, "<td>%s</td>", toUtf8(SaldoArsskifte))
		_, _ = fmt.Fprintf(w, "<td>%s</td>", toUtf8(ArsskifteManad))

		_, _ = fmt.Fprintf(w, "<td><form method=\"POST\" action=\"/konton\"><input type=\"hidden\" id=\"lopnr\" name=\"lopnr\" value=\"%d\"><input type=\"hidden\" id=\"action\" name=\"action\" value=\"editform\"><input type=\"submit\" value=\"Redigera\"></form></td>\n", Löpnr)
		_, _ = fmt.Fprintf(w, "<td><form method=\"POST\" action=\"/konton\"><input type=\"hidden\" id=\"lopnr\" name=\"lopnr\" value=\"%d\"><input type=\"hidden\" id=\"action\" name=\"action\" value=\"radera\"><input type=\"checkbox\" id=\"OK\" name=\"OK\" required><label for=\"OK\">OK</label><input type=\"submit\" value=\"Radera\"></form></td></tr>\n", Löpnr)
	}
	_, _ = fmt.Fprintf(w, "</table>\n")

	res.Close()

	_, _ = fmt.Fprintf(w, "<form method=\"POST\" action=\"/konton\"><input type=\"hidden\" id=\"action\" name=\"action\" value=\"addform\"><input type=\"submit\" value=\"Nytt konto\"></form>\n")
}

func printKontonFooter(w http.ResponseWriter) {
	_, _ = fmt.Fprintf(w, "<a href=\"summary\">Översikt</a>\n")
	_, _ = fmt.Fprintf(w, "</body>\n")
	_, _ = fmt.Fprintf(w, "</html>\n")
}

func raderaKonto(w http.ResponseWriter, lopnr int, db *sql.DB) {
	fmt.Println("raderaKonto lopnr: ", lopnr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := db.ExecContext(ctx,
		`DELETE FROM Konton WHERE (Löpnr=?)`, lopnr)

	if err != nil {
		log.Fatal(err)
	}
	_, _ = fmt.Fprintf(w, "Konto med löpnr %d raderat.<br>", lopnr)
}

func editformKonto(w http.ResponseWriter, lopnr int, db *sql.DB) {
	fmt.Println("editformKonto lopnr: ", lopnr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var KontoNummer []byte    // size 20
	var Benämning []byte      // size 40, index
	var Saldo []byte          // BCD / Decimal Precision 19
	var StartSaldo []byte     // BCD / Decimal Precision 19
	var StartManad []byte     // size 10
	var SaldoArsskifte []byte // BCD / Decimal Precision 19
	var ArsskifteManad []byte // size 10

	err := db.QueryRowContext(ctx,
		`SELECT KontoNummer,Benämning,Saldo,StartSaldo,StartManad,SaldoArsskifte,ArsskifteManad FROM Konton WHERE (Löpnr=?)`, lopnr).Scan(&KontoNummer, &Benämning, &Saldo, &StartSaldo, &StartManad, &SaldoArsskifte, &ArsskifteManad)
	if err != nil {
		log.Fatal(err)
	}

	_, _ = fmt.Fprintf(w, "Redigera konto<br>")
	_, _ = fmt.Fprintf(w, "<form method=\"POST\" action=\"/konton\">")

	_, _ = fmt.Fprintf(w, "<label for=\"Benamning\">Benämning:</label>")
	_, _ = fmt.Fprintf(w, "<input type=\"text\" id=\"Benamning\" name=\"Benamning\" value=\"%s\">", toUtf8(Benämning))
	_, _ = fmt.Fprintf(w, "<label for=\"Saldo\" hidden>Saldo:</label>")
	_, _ = fmt.Fprintf(w, "<input type=\"text\" id=\"Saldo\" name=\"Saldo\" value=\"%s\" hidden>", Saldo)
	_, _ = fmt.Fprintf(w, "<label for=\"StartSaldo\">StartSaldo:</label>")
	_, _ = fmt.Fprintf(w, "<input type=\"text\" id=\"StartSaldo\" name=\"StartSaldo\" value=\"%s\">", StartSaldo)
	_, _ = fmt.Fprintf(w, "<label for=\"StartManad\">StartMånad:</label>")
	_, _ = fmt.Fprintf(w, "<input type=\"text\" id=\"StartManad\" name=\"StartManad\" value=\"%s\">", StartManad)
	_, _ = fmt.Fprintf(w, "<label for=\"SaldoArsskifte\" hidden>Saldo Årsskifte:</label>")
	_, _ = fmt.Fprintf(w, "<input type=\"text\" id=\"SaldoArsskifte\" name=\"SaldoArsskifte\" value=\"%s\" hidden>", SaldoArsskifte)
	_, _ = fmt.Fprintf(w, "<label for=\"ArsskifteManad\" hidden>Årsskiftesmanad:</label>")
	_, _ = fmt.Fprintf(w, "<input type=\"text\" id=\"ArsskifteManad\" name=\"ArsskifteManad\" value=\"%s\" hidden>", ArsskifteManad)

	_, _ = fmt.Fprintf(w, "<input type=\"hidden\" id=\"lopnr\" name=\"lopnr\" value=\"%d\">", lopnr)
	_, _ = fmt.Fprintf(w, "<input type=\"hidden\" id=\"action\" name=\"action\" value=\"update\">")
	_, _ = fmt.Fprintf(w, "<input type=\"submit\" value=\"Uppdatera\">")
	_, _ = fmt.Fprintf(w, "</form>\n")
	_, _ = fmt.Fprintf(w, "<p>\n")
}

func month2Int(month time.Month) int {
	switch month {
	case time.January:
		return 1
	case time.February:
		return 2
	case time.March:
		return 3
	case time.April:
		return 4
	case time.May:
		return 5
	case time.June:
		return 6
	case time.July:
		return 7
	case time.August:
		return 8
	case time.September:
		return 9
	case time.October:
		return 10
	case time.November:
		return 11
	case time.December:
		return 12
	}
	return -1
}

func addformKonto(w http.ResponseWriter) {
	fmt.Println("addformKonto ")

	currentTime := time.Now()
	currentMonth := currentTime.Month()
	currentMonthInt := month2Int(currentMonth)

	_, _ = fmt.Fprintf(w, "Lägg till konto<br>")
	_, _ = fmt.Fprintf(w, "<form method=\"POST\" action=\"/konton\">")

	_, _ = fmt.Fprintf(w, "<label for=\"Benamning\">Benämning:</label>")
	_, _ = fmt.Fprintf(w, "<input type=\"text\" id=\"Benamning\" name=\"Benamning\" value=\"%s\">", "")
	_, _ = fmt.Fprintf(w, "<label for=\"StartSaldo\">Startsaldo:</label>")
	_, _ = fmt.Fprintf(w, "<input type=\"text\" id=\"StartSaldo\" name=\"StartSaldo\" value=\"%s\">", "")
	_, _ = fmt.Fprintf(w, "<label for=\"StartManad\">Startmånad:</label>")
	_, _ = fmt.Fprintf(w, "<select id=\"StartManad\" name=\"StartManad\" required>\n")
	for month := 1; month < 13; month++ {
		_, _ = fmt.Fprintf(w, "<option value=\"%d\"", month)
		if month == currentMonthInt {
			_, _ = fmt.Fprintf(w, " selected ")
		}
		_, _ = fmt.Fprintf(w, ">%d</option>\n", month)
	}
	_, _ = fmt.Fprintf(w, "</select>\n")

	_, _ = fmt.Fprintf(w, "<input type=\"hidden\" id=\"action\" name=\"action\" value=\"add\">")
	_, _ = fmt.Fprintf(w, "<input type=\"submit\" value=\"Nytt konto\">")
	_, _ = fmt.Fprintf(w, "</form>\n")
	_, _ = fmt.Fprintf(w, "<p>\n")
}

func int2man(month int) string {
	switch month {
	case 1:
		return "Jan"
	case 2:
		return "Feb"
	case 3:
		return "Mar"
	case 4:
		return "Apr"
	case 5:
		return "Maj"
	case 6:
		return "Jun"
	case 7:
		return "Jul"
	case 8:
		return "Aug"
	case 9:
		return "Sep"
	case 10:
		return "Okt"
	case 11:
		return "Nov"
	case 12:
		return "Dec"
	}

	// Fail HARD!
	log.Fatal("int2man: unknown month " + strconv.Itoa(month))
	return "???"
}

func addKonto(Benamning string, StartSaldo decimal.Decimal, StartManad string, db *sql.DB) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var startSaldo = "Not Possible"
	//StartMan, _ := strconv.Atoi(StartManad)
	if JetDBSupport {
		startSaldo = strings.ReplaceAll(StartSaldo.String(), ".", ",")
	} else {
		startSaldo = StartSaldo.String()
	}

	if month, err := strconv.Atoi(StartManad); err == nil {
		StartManad = int2man(month)
	}

	_, err := db.ExecContext(ctx,
		`INSERT INTO Konton(KontoNummer,Benämning,Saldo,StartSaldo,SaldoArsskifte,StartManad,ArsskifteManad) VALUES (?, ?, ?, ?, ?, ?, ?)`, 0, Benamning, startSaldo, startSaldo, startSaldo, StartManad, StartManad)

	if err != nil {
		log.Fatal(err)
	}
}

func addKontow(w http.ResponseWriter, Benamning string, StartSaldo string, StartManad string, db *sql.DB) {
	fmt.Println("addKontow namn: ", Benamning)

	startSaldo, err := decimal.NewFromString(StartSaldo)
	if err != nil {
		log.Print(err)
		startSaldo = decimal.NewFromInt(0)
	}
	addKonto(Benamning, startSaldo, StartManad, db)

	_, _ = fmt.Fprintf(w, "Konto %s tillagd.<br>", Benamning)
}

func updateKonto(w http.ResponseWriter, lopnr int, Benamning string, StartSaldo string, StartManad string, db *sql.DB) {
	fmt.Println("updateKonto lopnr: ", lopnr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := db.ExecContext(ctx,
		`UPDATE Konton SET Benämning = ?, StartSaldo = ?, StartManad = ? WHERE (Löpnr=?)`, Benamning, strings.ReplaceAll(StartSaldo, ".", ","), StartManad, lopnr)

	if err != nil {
		log.Fatal(err)
	}
	_, _ = fmt.Fprintf(w, "Konto %s uppdaterad.<br>", Benamning)
}

func updateKontoSaldo(Benamning string, Saldo decimal.Decimal) {
	lopnr := hämtakontoID(db, Benamning)
	//fmt.Println("updateKontoSaldo lopnr: ", lopnr)
	var amount string = AmountDec2DBStr(Saldo)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := db.ExecContext(ctx,
		`UPDATE Konton SET Saldo = ? WHERE (Löpnr=?)`, amount, lopnr)

	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println("Konto uppdaterad, nytt saldo.", Benamning, amount)
}

func hanterakonton(w http.ResponseWriter, req *http.Request) {
	_, _ = fmt.Fprintf(w, "<html>\n")
	_, _ = fmt.Fprintf(w, "<head>\n")
	_, _ = fmt.Fprintf(w, "<style>\n")
	_, _ = fmt.Fprintf(w, "table,th,td { border: 1px solid black }\n")
	_, _ = fmt.Fprintf(w, "</style>\n")
	_, _ = fmt.Fprintf(w, "</head>\n")
	_, _ = fmt.Fprintf(w, "<body>\n")

	_, _ = fmt.Fprintf(w, "<h1>%s</h1>\n", currentDatabase)
	_, _ = fmt.Fprintf(w, "<h2>Konton</h2>\n")

	err := req.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	formaction := req.FormValue("action")
	var lopnr = -1
	if len(req.FormValue("lopnr")) > 0 {
		lopnr, _ = strconv.Atoi(req.FormValue("lopnr"))
	}

	switch formaction {
	case "radera":
		raderaKonto(w, lopnr, db)
	case "addform":
		addformKonto(w)
	case "add":
		var Benamning = ""
		if len(req.FormValue("Benamning")) > 0 {
			Benamning = req.FormValue("Benamning")
		}
		var StartSaldo = ""
		if len(req.FormValue("StartSaldo")) > 0 {
			StartSaldo = req.FormValue("StartSaldo")
		}
		var StartManad = ""
		if len(req.FormValue("StartManad")) > 0 {
			StartManad = req.FormValue("StartManad")
		}
		addKontow(w, Benamning, StartSaldo, StartManad, db)
	case "editform":
		editformKonto(w, lopnr, db)
	case "update":
		var Benamning = ""
		if len(req.FormValue("Benamning")) > 0 {
			Benamning = req.FormValue("Benamning")
		}
		var StartSaldo = ""
		if len(req.FormValue("StartSaldo")) > 0 {
			StartSaldo = req.FormValue("StartSaldo")
		}
		var StartManad = ""
		if len(req.FormValue("StartManad")) > 0 {
			StartManad = req.FormValue("StartManad")
		}
		updateKonto(w, lopnr, Benamning, StartSaldo, StartManad, db)
	default:
		fmt.Println("Okänd action: ", formaction)
	}
	printKonton(w, db)
	printKontonFooter(w)
}

func getAccNames() []string {
	names := make([]string, 0)

	res, err := db.Query("SELECT Benämning FROM Konton ORDER BY Benämning")

	if err != nil {
		log.Fatal(err)
	}

	var Benämning []byte // size 40, index
	for res.Next() {
		_ = res.Scan(&Benämning)
		names = append(names, toUtf8(Benämning))
	}
	res.Close()
	return names
}

func antalKonton() int {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var antal int

	err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM Konton`).Scan(&antal)
	if err != nil {
		log.Fatal(err)
	}

	return antal
}

func hämtakontoID(db *sql.DB, accName string) int {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var Löpnr int // autoinc Primary Key
	err := db.QueryRowContext(ctx,
		`select Löpnr
  from konton
  where benämning = ?`, accName).Scan(&Löpnr)
	if err != nil {
		log.Fatal(err)
	}

	return Löpnr
}

func hämtaKonto(db *sql.DB, lopnr int) konto {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var KontoNummer []byte    // size 20
	var Benämning []byte      // size 40, index
	var Saldo []byte          // BCD / Decimal Precision 19
	var StartSaldo []byte     // BCD / Decimal Precision 19
	var StartManad []byte     // size 10
	var SaldoArsskifte []byte // BCD / Decimal Precision 19
	var ArsskifteManad []byte // size 10

	err := db.QueryRowContext(ctx,
		`SELECT KontoNummer,Benämning,Saldo,StartSaldo,StartManad,SaldoArsskifte,ArsskifteManad FROM Konton WHERE (Löpnr=?)`, lopnr).Scan(&KontoNummer, &Benämning, &Saldo, &StartSaldo, &StartManad, &SaldoArsskifte, &ArsskifteManad)
	if err != nil {
		log.Fatal(err)
	}

	var retkonto konto

	retkonto.KontoNummer = toUtf8(KontoNummer)
	retkonto.Benämning = toUtf8(Benämning)
	retkonto.Saldo, _ = decimal.NewFromString(toUtf8(Saldo))
	retkonto.StartSaldo, _ = decimal.NewFromString(toUtf8(StartSaldo))
	retkonto.StartManad = toUtf8(StartManad)
	retkonto.SaldoArsskifte = toUtf8(SaldoArsskifte)
	retkonto.ArsskifteManad = toUtf8(ArsskifteManad)

	return retkonto
}

func saldoKonto(db *sql.DB, accName string, endDate string) decimal.Decimal {
	//	fmt.Println("saldoKonto: accName ", accName)
	if endDate == "" {
		endDate = "2999-12-31"
	}
	//fmt.Println("saldoKonto: endDate ", endDate)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var err error
	var res *sql.Rows

	var rawStart []byte // size 16

	_ = db.QueryRowContext(ctx,
		`select startsaldo
  from konton
  where benämning = ?`, accName).Scan(&rawStart)
	res2 := toUtf8(rawStart)
	startSaldo, _ := decimal.NewFromString(res2)
	currSaldo := startSaldo
	//fmt.Println("saldoKonto: startsaldo ", currSaldo)

	res, err = db.QueryContext(ctx,
		`SELECT FrånKonto,TillKonto,Typ,Datum,Vad,Vem,Belopp,Löpnr,Saldo,Fastöverföring,Text from transaktioner
  where (datum <= ?)
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
		_ = res.Scan(&fromAcc, &toAcc, &tType, &date, &what, &who, &amount, &nummer, &saldo, &fixed, &comment)
		decAmount, _ := decimal.NewFromString(toUtf8(amount))
		//fmt.Println("saldoKonto: decAmount ", decAmount)
		//fmt.Println("saldoKonto: toAcc ", toUtf8(toAcc))
		//fmt.Println("saldoKonto: fromAcc ", toUtf8(fromAcc))
		//fmt.Println("saldoKonto: tType ", toUtf8(tType))

		if (accName == toUtf8(toAcc)) &&
			((toUtf8(tType) == "Uttag") ||
				(toUtf8(tType) == "Fast Inkomst") ||
				(toUtf8(tType) == "Insättning") ||
				(toUtf8(tType) == "Överföring")) {
			currSaldo = currSaldo.Add(decAmount)
			//fmt.Println("saldoKonto: add")
		}
		if (accName == toUtf8(fromAcc)) &&
			((toUtf8(tType) == "Uttag") ||
				(toUtf8(tType) == "Inköp") ||
				(toUtf8(tType) == "Fast Utgift") ||
				(toUtf8(tType) == "Överföring")) {
			currSaldo = currSaldo.Sub(decAmount)
			//fmt.Println("saldoKonto: sub")
		}
		//fmt.Println("saldoKonto: new saldo ", currSaldo)
	}
	res.Close()
	return currSaldo
}

func saldonKonto(db *sql.DB, accName string, endDate string) (decimal.Decimal, decimal.Decimal) {
	//fmt.Println("saldoKonto: accName ", accName)
	//fmt.Println("saldoKonto: endDate ", endDate)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var err error
	var res *sql.Rows

	var rawStart []byte // size 16
	_ = db.QueryRowContext(ctx,
		`select startsaldo
  from konton
  where benämning = ?`, accName).Scan(&rawStart)
	res2 := toUtf8(rawStart)
	startSaldo, _ := decimal.NewFromString(res2)
	currSaldo := startSaldo
	totSaldo := currSaldo
	//fmt.Println("saldoKonto: startsaldo ", currSaldo)

	res, err = db.QueryContext(ctx,
		`SELECT FrånKonto,TillKonto,Typ,Datum,Vad,Vem,Belopp,Löpnr,Saldo,Fastöverföring,Text from transaktioner
  where 
         ((tillkonto = ?)
         or (frånkonto = ?))
order by datum,löpnr`, accName, accName)

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
		_ = res.Scan(&fromAcc, &toAcc, &tType, &date, &what, &who, &amount, &nummer, &saldo, &fixed, &comment)
		decAmount, _ := decimal.NewFromString(toUtf8(amount))
		//fmt.Println("saldoKonto: decAmount ", decAmount)
		//fmt.Println("saldoKonto: toAcc ", toUtf8(toAcc))
		//fmt.Println("saldoKonto: fromAcc ", toUtf8(fromAcc))
		//fmt.Println("saldoKonto: tType ", toUtf8(tType))

		if (accName == toUtf8(toAcc)) &&
			((toUtf8(tType) == "Uttag") ||
				(toUtf8(tType) == "Fast Inkomst") ||
				(toUtf8(tType) == "Insättning") ||
				(toUtf8(tType) == "Överföring")) {
			if toUtf8(date) <= endDate {
				currSaldo = currSaldo.Add(decAmount)
			}
			totSaldo = totSaldo.Add(decAmount)
			//fmt.Println("saldoKonto: add")
		}
		if (accName == toUtf8(fromAcc)) &&
			((toUtf8(tType) == "Uttag") ||
				(toUtf8(tType) == "Inköp") ||
				(toUtf8(tType) == "Fast Utgift") ||
				(toUtf8(tType) == "Överföring")) {
			if toUtf8(date) <= endDate {
				currSaldo = currSaldo.Sub(decAmount)
			}
			totSaldo = totSaldo.Sub(decAmount)
			//fmt.Println("saldoKonto: sub")
		}
		//fmt.Println("saldoKonto: new saldo ", currSaldo, " totSaldo ", totSaldo)
	}
	res.Close()
	return currSaldo, totSaldo
}
