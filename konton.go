//-*- coding: utf-8 -*-

package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/alexbrainman/odbc" // BSD-3-Clause License
	_ "github.com/mattn/go-sqlite3"  // MIT License
)

func printKonton(w http.ResponseWriter, db *sql.DB) {
	res, err := db.Query("SELECT KontoNummer,Benämning,Saldo,StartSaldo,StartManad,Löpnr,SaldoArsskifte,ArsskifteManad FROM Konton")

	if err != nil {
		log.Fatal(err)
		os.Exit(2)
	}

	var KontoNummer []byte    // size 20
	var Benämning []byte      // size 40, index
	var Saldo []byte          // BCD / Decimal Precision 19
	var StartSaldo []byte     // BCD / Decimal Precision 19
	var StartManad []byte     // size 10
	var Löpnr int             // autoinc Primary Key
	var SaldoArsskifte []byte // BCD / Decimal Precision 19
	var ArsskifteManad []byte // size 10

	fmt.Fprintf(w, "<table style=\"width:100%%\"><tr>")
	fmt.Fprintf(w, "<th>Kontonummer</th>")
	fmt.Fprintf(w, "<th>Benämning</th>")
	fmt.Fprintf(w, "<th>Saldo</th>")
	fmt.Fprintf(w, "<th>Startsaldo</th>")
	fmt.Fprintf(w, "<th>Startmånad</th>")
	fmt.Fprintf(w, "<th>Saldo årsskifte</th>")
	fmt.Fprintf(w, "<th>Årsskiftesmånad</th>")
	fmt.Fprintf(w, "<th>Redigera</th><th>Radera</th>\n")
	for res.Next() {
		err = res.Scan(&KontoNummer, &Benämning, &Saldo, &StartSaldo, &StartManad, &Löpnr, &SaldoArsskifte, &ArsskifteManad)

		fmt.Fprintf(w, "<tr>")
		fmt.Fprintf(w, "<td>%s</td>", toUtf8(KontoNummer))
		fmt.Fprintf(w, "<td>%s</td>", toUtf8(Benämning))
		fmt.Fprintf(w, "<td>%s</td>", toUtf8(Saldo))
		fmt.Fprintf(w, "<td>%s</td>", toUtf8(StartSaldo))
		fmt.Fprintf(w, "<td>%s</td>", toUtf8(StartManad))
		fmt.Fprintf(w, "<td>%s</td>", toUtf8(SaldoArsskifte))
		fmt.Fprintf(w, "<td>%s</td>", toUtf8(ArsskifteManad))

		fmt.Fprintf(w, "<td><form method=\"POST\" action=\"/konton\"><input type=\"hidden\" id=\"lopnr\" name=\"lopnr\" value=\"%d\"><input type=\"hidden\" id=\"action\" name=\"action\" value=\"editform\"><input type=\"submit\" value=\"Redigera\"></form></td>\n", Löpnr)
		fmt.Fprintf(w, "<td><form method=\"POST\" action=\"/konton\"><input type=\"hidden\" id=\"lopnr\" name=\"lopnr\" value=\"%d\"><input type=\"hidden\" id=\"action\" name=\"action\" value=\"radera\"><input type=\"checkbox\" id=\"OK\" name=\"OK\" required><label for=\"OK\">OK</label><input type=\"submit\" value=\"Radera\"></form></td></tr>\n", Löpnr)
	}
	fmt.Fprintf(w, "</table>\n")

	fmt.Fprintf(w, "<form method=\"POST\" action=\"/konton\"><input type=\"hidden\" id=\"action\" name=\"action\" value=\"addform\"><input type=\"submit\" value=\"Nytt konto\"></form>\n")
}

func printKontonFooter(w http.ResponseWriter, db *sql.DB) {
	fmt.Fprintf(w, "<a href=\"summary\">Översikt</a>\n")
	fmt.Fprintf(w, "</body>\n")
	fmt.Fprintf(w, "</html>\n")
}

func raderaKonto(w http.ResponseWriter, lopnr int, db *sql.DB) {
	fmt.Println("raderaKonto lopnr: ", lopnr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := db.ExecContext(ctx,
		`DELETE FROM Konton WHERE (Löpnr=?)`, lopnr)

	if err != nil {
		log.Fatal(err)
		os.Exit(2)
	}
	fmt.Fprintf(w, "Konto med löpnr %d raderat.<br>", lopnr)
}

func editformKonto(w http.ResponseWriter, lopnr int, db *sql.DB) {
	fmt.Println("editformKonto lopnr: ", lopnr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	res1 := db.QueryRowContext(ctx,
		`SELECT KontoNummer,Benämning,Saldo,StartSaldo,StartManad,SaldoArsskifte,ArsskifteManad FROM Konton WHERE (Löpnr=?)`, lopnr)

	var KontoNummer []byte    // size 20
	var Benämning []byte      // size 40, index
	var Saldo []byte          // BCD / Decimal Precision 19
	var StartSaldo []byte     // BCD / Decimal Precision 19
	var StartManad []byte     // size 10
	var SaldoArsskifte []byte // BCD / Decimal Precision 19
	var ArsskifteManad []byte // size 10

	err := res1.Scan(&KontoNummer, &Benämning, &Saldo, &StartSaldo, &StartManad, &SaldoArsskifte, &ArsskifteManad)
	if err != nil {
		log.Fatal(err)
		os.Exit(2)
	}

	fmt.Fprintf(w, "Redigera konto<br>")
	fmt.Fprintf(w, "<form method=\"POST\" action=\"/konton\">")

	fmt.Fprintf(w, "<label for=\"Benamning\">Benämning:</label>")
	fmt.Fprintf(w, "<input type=\"text\" id=\"Benamning\" name=\"Benamning\" value=\"%s\">", toUtf8(Benämning))
	fmt.Fprintf(w, "<label for=\"Saldo\" hidden>Saldo:</label>")
	fmt.Fprintf(w, "<input type=\"text\" id=\"Saldo\" name=\"Saldo\" value=\"%s\" hidden>", Saldo)
	fmt.Fprintf(w, "<label for=\"StartSaldo\">StartSaldo:</label>")
	fmt.Fprintf(w, "<input type=\"text\" id=\"StartSaldo\" name=\"StartSaldo\" value=\"%s\">", StartSaldo)
	fmt.Fprintf(w, "<label for=\"StartManad\">StartMånad:</label>")
	fmt.Fprintf(w, "<input type=\"text\" id=\"StartManad\" name=\"StartManad\" value=\"%s\">", StartManad)
	fmt.Fprintf(w, "<label for=\"SaldoArsskifte\" hidden>Saldo Årsskifte:</label>")
	fmt.Fprintf(w, "<input type=\"text\" id=\"SaldoArsskifte\" name=\"SaldoArsskifte\" value=\"%s\" hidden>", SaldoArsskifte)
	fmt.Fprintf(w, "<label for=\"ArsskifteManad\" hidden>Årsskiftesmanad:</label>")
	fmt.Fprintf(w, "<input type=\"text\" id=\"ArsskifteManad\" name=\"ArsskifteManad\" value=\"%s\" hidden>", ArsskifteManad)

	fmt.Fprintf(w, "<input type=\"hidden\" id=\"lopnr\" name=\"lopnr\" value=\"%d\">", lopnr)
	fmt.Fprintf(w, "<input type=\"hidden\" id=\"action\" name=\"action\" value=\"update\">")
	fmt.Fprintf(w, "<input type=\"submit\" value=\"Uppdatera\">")
	fmt.Fprintf(w, "</form>\n")
	fmt.Fprintf(w, "<p>\n")
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

func addformKonto(w http.ResponseWriter, db *sql.DB) {
	fmt.Println("addformKonto ")

	currentTime := time.Now()
	currentMonth := currentTime.Month()
	currentMonthInt := month2Int(currentMonth)

	fmt.Fprintf(w, "Lägg till konto<br>")
	fmt.Fprintf(w, "<form method=\"POST\" action=\"/konton\">")

	fmt.Fprintf(w, "<label for=\"Benamning\">Benämning:</label>")
	fmt.Fprintf(w, "<input type=\"text\" id=\"Benamning\" name=\"Benamning\" value=\"%s\">", "")
	fmt.Fprintf(w, "<label for=\"StartSaldo\">Startsaldo:</label>")
	fmt.Fprintf(w, "<input type=\"text\" id=\"StartSaldo\" name=\"StartSaldo\" value=\"%s\">", "")
	fmt.Fprintf(w, "<label for=\"StartManad\">Startmånad:</label>")
	fmt.Fprintf(w, "<select id=\"StartManad\" name=\"StartManad\" required>\n")
	for month := 1; month < 13; month++ {
		fmt.Fprintf(w, "<option value=\"%d\"", month)
		if month == currentMonthInt {
			fmt.Fprintf(w, " selected ")
		}
		fmt.Fprintf(w, ">%d</option>\n", month)
	}
	fmt.Fprintf(w, "</select>\n")

	fmt.Fprintf(w, "<input type=\"hidden\" id=\"action\" name=\"action\" value=\"add\">")
	fmt.Fprintf(w, "<input type=\"submit\" value=\"Nytt konto\">")
	fmt.Fprintf(w, "</form>\n")
	fmt.Fprintf(w, "<p>\n")
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
	return "???"
}

func addKonto(w http.ResponseWriter, Benamning string, StartSaldo string, StartManad string, db *sql.DB) {
	fmt.Println("addKonto namn: ", Benamning)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	StartMan, _ := strconv.Atoi(StartManad)
	_, err := db.ExecContext(ctx,
		`INSERT INTO Konton(KontoNummer,Benämning,Saldo,StartSaldo,SaldoArsskifte,StartManad,ArsskifteManad) VALUES (?, ?, ?, ?, ?, ?, ?)`, 0, Benamning, strings.ReplaceAll(StartSaldo, ".", ","), strings.ReplaceAll(StartSaldo, ".", ","), strings.ReplaceAll(StartSaldo, ".", ","), int2man(StartMan), int2man(1))

	if err != nil {
		log.Fatal(err)
		os.Exit(2)
	}
	fmt.Fprintf(w, "Konto %s tillagd.<br>", Benamning)
}

func updateKonto(w http.ResponseWriter, lopnr int, Benamning string, StartSaldo string, StartManad string, db *sql.DB) {
	fmt.Println("updateKonto lopnr: ", lopnr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := db.ExecContext(ctx,
		`UPDATE Konton SET Benämning = ?, StartSaldo = ?, StartManad = ? WHERE (Löpnr=?)`, Benamning, strings.ReplaceAll(StartSaldo, ".", ","), StartManad, lopnr)

	if err != nil {
		log.Fatal(err)
		os.Exit(2)
	}
	fmt.Fprintf(w, "Konto %s uppdaterad.<br>", Benamning)
}

func hanterakonton(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "<html>\n")
	fmt.Fprintf(w, "<head>\n")
	fmt.Fprintf(w, "<style>\n")
	fmt.Fprintf(w, "table,th,td { border: 1px solid black }\n")
	fmt.Fprintf(w, "</style>\n")
	fmt.Fprintf(w, "</head>\n")
	fmt.Fprintf(w, "<body>\n")

	fmt.Fprintf(w, "<h1>%s</h1>\n", currentDatabase)
	fmt.Fprintf(w, "<h2>Konton</h2>\n")

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
		raderaKonto(w, lopnr, db)
	case "addform":
		addformKonto(w, db)
	case "add":
		var Benamning string = ""
		if len(req.FormValue("Benamning")) > 0 {
			Benamning = req.FormValue("Benamning")
		}
		var StartSaldo string = ""
		if len(req.FormValue("StartSaldo")) > 0 {
			StartSaldo = req.FormValue("StartSaldo")
		}
		var StartManad string = ""
		if len(req.FormValue("StartManad")) > 0 {
			StartManad = req.FormValue("StartManad")
		}
		addKonto(w, Benamning, StartSaldo, StartManad, db)
	case "editform":
		editformKonto(w, lopnr, db)
	case "update":
		var Benamning string = ""
		if len(req.FormValue("Benamning")) > 0 {
			Benamning = req.FormValue("Benamning")
		}
		var StartSaldo string = ""
		if len(req.FormValue("StartSaldo")) > 0 {
			StartSaldo = req.FormValue("StartSaldo")
		}
		var StartManad string = ""
		if len(req.FormValue("StartManad")) > 0 {
			StartManad = req.FormValue("StartManad")
		}
		updateKonto(w, lopnr, Benamning, StartSaldo, StartManad, db)
	default:
		fmt.Println("Okänd action: %s\n", formaction)
	}
	printKonton(w, db)
	printKontonFooter(w, db)
}

func getAccNames() []string {
	names := make([]string, 0)

	res, err := db.Query("SELECT Benämning FROM Konton ORDER BY Benämning")

	if err != nil {
		log.Fatal(err)
		os.Exit(2)
	}

	var Benämning []byte // size 40, index
	for res.Next() {
		err = res.Scan(&Benämning)
		names = append(names, toUtf8(Benämning))
	}
	return names
}
