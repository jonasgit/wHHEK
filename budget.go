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
)

func printBudget(w http.ResponseWriter, db *sql.DB) {
	res, err := db.Query("SELECT Löpnr,Typ,Inkomst,HurOfta,StartMånad,Jan,Feb,Mar,Apr,Maj,Jun,Jul,Aug,Sep,Okt,Nov,Dec,Kontrollnr FROM Budget ORDER BY Inkomst ASC, Typ ASC")

	if err != nil {
		log.Fatal(err)
	}

	var Typ []byte        // size 40
	var Inkomst []byte    // size 1
	var HurOfta int16     // SmallInt
	var StartMånad []byte // size 10
	var Jan []byte        // BCD / Decimal Precision 19
	var Feb []byte        // BCD / Decimal Precision 19
	var Mar []byte        // BCD / Decimal Precision 19
	var Apr []byte        // BCD / Decimal Precision 19
	var Maj []byte        // BCD / Decimal Precision 19
	var Jun []byte        // BCD / Decimal Precision 19
	var Jul []byte        // BCD / Decimal Precision 19
	var Aug []byte        // BCD / Decimal Precision 19
	var Sep []byte        // BCD / Decimal Precision 19
	var Okt []byte        // BCD / Decimal Precision 19
	var Nov []byte        // BCD / Decimal Precision 19
	var Dec []byte        // BCD / Decimal Precision 19
	var Kontrollnr []byte //int32 // Integer
	var Löpnr int         // autoinc Primary Key, index

	_, _ = fmt.Fprintf(w, "<table style=\"width:100%%\"><tr>")
	_, _ = fmt.Fprintf(w, "<th>Löpnr</th>")
	_, _ = fmt.Fprintf(w, "<th>Typ</th>")
	_, _ = fmt.Fprintf(w, "<th>Inkomst</th>")
	_, _ = fmt.Fprintf(w, "<th>HurOfta</th>")
	_, _ = fmt.Fprintf(w, "<th>StartMånad</th>")
	_, _ = fmt.Fprintf(w, "<th>Jan</th>")
	_, _ = fmt.Fprintf(w, "<th>Feb</th>")
	_, _ = fmt.Fprintf(w, "<th>Mar</th>")
	_, _ = fmt.Fprintf(w, "<th>Apr</th>")
	_, _ = fmt.Fprintf(w, "<th>Maj</th>")
	_, _ = fmt.Fprintf(w, "<th>Jun</th>")
	_, _ = fmt.Fprintf(w, "<th>Jul</th>")
	_, _ = fmt.Fprintf(w, "<th>Aug</th>")
	_, _ = fmt.Fprintf(w, "<th>Sep</th>")
	_, _ = fmt.Fprintf(w, "<th>Okt</th>")
	_, _ = fmt.Fprintf(w, "<th>Nov</th>")
	_, _ = fmt.Fprintf(w, "<th>Dec</th>")
	_, _ = fmt.Fprintf(w, "<th>Kontrollnr</th>")
	_, _ = fmt.Fprintf(w, "<th>Redigera</th><th>Radera</th>\n")
	for res.Next() {
		err = res.Scan(&Löpnr, &Typ, &Inkomst, &HurOfta, &StartMånad, &Jan, &Feb, &Mar, &Apr, &Maj, &Jun, &Jul, &Aug, &Sep, &Okt, &Nov, &Dec, &Kontrollnr)

		_, _ = fmt.Fprintf(w, "<tr>")
		_, _ = fmt.Fprintf(w, "<td>%d</td>", Löpnr)
		_, _ = fmt.Fprintf(w, "<td>%s</td>", toUtf8(Typ))
		_, _ = fmt.Fprintf(w, "<td>%s</td>", toUtf8(Inkomst))
		_, _ = fmt.Fprintf(w, "<td>%s</td>", strconv.Itoa(int(HurOfta)))
		_, _ = fmt.Fprintf(w, "<td>%s</td>", toUtf8(StartMånad))
		_, _ = fmt.Fprintf(w, "<td>%s</td>", toUtf8(Jan))
		_, _ = fmt.Fprintf(w, "<td>%s</td>", toUtf8(Feb))
		_, _ = fmt.Fprintf(w, "<td>%s</td>", toUtf8(Mar))
		_, _ = fmt.Fprintf(w, "<td>%s</td>", toUtf8(Apr))
		_, _ = fmt.Fprintf(w, "<td>%s</td>", toUtf8(Maj))
		_, _ = fmt.Fprintf(w, "<td>%s</td>", toUtf8(Jun))
		_, _ = fmt.Fprintf(w, "<td>%s</td>", toUtf8(Jul))
		_, _ = fmt.Fprintf(w, "<td>%s</td>", toUtf8(Aug))
		_, _ = fmt.Fprintf(w, "<td>%s</td>", toUtf8(Sep))
		_, _ = fmt.Fprintf(w, "<td>%s</td>", toUtf8(Okt))
		_, _ = fmt.Fprintf(w, "<td>%s</td>", toUtf8(Nov))
		_, _ = fmt.Fprintf(w, "<td>%s</td>", toUtf8(Dec))
		if Kontrollnr != nil {
			_, _ = fmt.Fprintf(w, "<td>%s</td>", toUtf8(Kontrollnr))
		} else {
			_, _ = fmt.Fprintf(w, "<td>%s</td>", "null")
		}

		_, _ = fmt.Fprintf(w, "<td><form method=\"POST\" action=\"/budget\"><input type=\"hidden\" id=\"lopnr\" name=\"lopnr\" value=\"%d\"><input type=\"hidden\" id=\"action\" name=\"action\" value=\"editform\"><input type=\"submit\" value=\"Redigera\"></form></td>\n", Löpnr)
		_, _ = fmt.Fprintf(w, "<td><form method=\"POST\" action=\"/budget\"><input type=\"hidden\" id=\"lopnr\" name=\"lopnr\" value=\"%d\"><input type=\"hidden\" id=\"action\" name=\"action\" value=\"radera\"><input type=\"checkbox\" id=\"OK\" name=\"OK\" required><label for=\"OK\">OK</label><input type=\"submit\" value=\"Radera\"></form></td></tr>\n", Löpnr)
	}
	_, _ = fmt.Fprintf(w, "</table>\n")

	_, _ = fmt.Fprintf(w, "<form method=\"POST\" action=\"/budget\"><input type=\"hidden\" id=\"action\" name=\"action\" value=\"addform\"><input type=\"submit\" value=\"Ny budgetpost\"></form>\n")
}

func printBudgetFooter(w http.ResponseWriter) {
	_, _ = fmt.Fprintf(w, "<a href=\"summary\">Översikt</a>\n")
	_, _ = fmt.Fprintf(w, "</body>\n")
	_, _ = fmt.Fprintf(w, "</html>\n")
}

func PrintEditCellText(w http.ResponseWriter, label string, title string, value string) {
	_, _ = fmt.Fprintf(w, "<label for=\"%s\">%s:</label>", label, title)
	_, _ = fmt.Fprintf(w, "<input type=\"text\" id=\"%s\" name=\"%s\" value=\"%s\" >", label, label, value)
}

func editformBudget(w http.ResponseWriter, lopnr int, db *sql.DB) {
	fmt.Println("editformBudget lopnr: ", lopnr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	res1 := db.QueryRowContext(ctx,
		`SELECT Löpnr,Typ,Inkomst,HurOfta,StartMånad,Jan,Feb,Mar,Apr,Maj,Jun,Jul,Aug,Sep,Okt,Nov,Dec,Kontrollnr FROM Budget WHERE (Löpnr=?)`, lopnr)

	var Typ []byte        // size 40
	var Inkomst []byte    // size 1
	var HurOfta int16     // SmallInt
	var StartMånad []byte // size 10
	var Jan []byte        // BCD / Decimal Precision 19
	var Feb []byte        // BCD / Decimal Precision 19
	var Mar []byte        // BCD / Decimal Precision 19
	var Apr []byte        // BCD / Decimal Precision 19
	var Maj []byte        // BCD / Decimal Precision 19
	var Jun []byte        // BCD / Decimal Precision 19
	var Jul []byte        // BCD / Decimal Precision 19
	var Aug []byte        // BCD / Decimal Precision 19
	var Sep []byte        // BCD / Decimal Precision 19
	var Okt []byte        // BCD / Decimal Precision 19
	var Nov []byte        // BCD / Decimal Precision 19
	var Dec []byte        // BCD / Decimal Precision 19
	var Kontrollnr []byte //int32 // Integer
	var Löpnr int         // autoinc Primary Key, index

	err := res1.Scan(&Löpnr, &Typ, &Inkomst, &HurOfta, &StartMånad, &Jan, &Feb, &Mar, &Apr, &Maj, &Jun, &Jul, &Aug, &Sep, &Okt, &Nov, &Dec, &Kontrollnr)
	if err != nil {
		log.Fatal(err)
	}

	_, _ = fmt.Fprintf(w, "Redigera budgetpost<br>")
	_, _ = fmt.Fprintf(w, "<form method=\"POST\" action=\"/budget\">")

	PrintEditCellText(w, "Typ", "Typ", toUtf8(Typ))
	PrintEditCellText(w, "Inkomst", "Inkomst", toUtf8(Inkomst))
	PrintEditCellText(w, "HurOfta", "HurOfta", strconv.Itoa(int(HurOfta)))
	PrintEditCellText(w, "StartMånad", "StartMånad", toUtf8(StartMånad))
	PrintEditCellText(w, "Jan", "Jan", toUtf8(Jan))
	PrintEditCellText(w, "Feb", "Feb", toUtf8(Feb))
	PrintEditCellText(w, "Mar", "Mar", toUtf8(Mar))
	PrintEditCellText(w, "Apr", "Apr", toUtf8(Apr))
	PrintEditCellText(w, "Maj", "Maj", toUtf8(Maj))
	PrintEditCellText(w, "Jun", "Jun", toUtf8(Jun))
	PrintEditCellText(w, "Jul", "Jul", toUtf8(Jul))
	PrintEditCellText(w, "Aug", "Aug", toUtf8(Aug))
	PrintEditCellText(w, "Sep", "Sep", toUtf8(Sep))
	PrintEditCellText(w, "Okt", "Okt", toUtf8(Okt))
	PrintEditCellText(w, "Nov", "Nov", toUtf8(Nov))
	PrintEditCellText(w, "Dec", "Dec", toUtf8(Dec))
	if Kontrollnr != nil {
		PrintEditCellText(w, "Kontrollnr", "Kontrollnr", toUtf8(Kontrollnr))
	} else {
		PrintEditCellText(w, "Kontrollnr", "Kontrollnr", "null")
	}

	_, _ = fmt.Fprintf(w, "<input type=\"hidden\" id=\"lopnr\" name=\"lopnr\" value=\"%d\">", Löpnr)
	_, _ = fmt.Fprintf(w, "<input type=\"hidden\" id=\"action\" name=\"action\" value=\"update\">")
	_, _ = fmt.Fprintf(w, "<input type=\"submit\" value=\"Uppdatera\">")
	_, _ = fmt.Fprintf(w, "</form>\n")
	_, _ = fmt.Fprintf(w, "<p>\n")
}

func updateBudget(w http.ResponseWriter, lopnr int, req *http.Request, db *sql.DB) {
	fmt.Println("updateBudget lopnr: ", lopnr)

	var Typ = ""
	if len(req.FormValue("Typ")) > 0 {
		Typ = req.FormValue("Typ")
	}
	var Inkomst = ""
	if len(req.FormValue("Inkomst")) > 0 {
		Inkomst = req.FormValue("Inkomst")
	}
	var HurOfta = ""
	if len(req.FormValue("HurOfta")) > 0 {
		HurOfta = req.FormValue("HurOfta")
	}
	var StartMånad = ""
	if len(req.FormValue("StartMånad")) > 0 {
		StartMånad = req.FormValue("StartMånad")
	}
	var Jan = ""
	if len(req.FormValue("Jan")) > 0 {
		Jan = req.FormValue("Jan")
	}
	var Feb = ""
	if len(req.FormValue("Feb")) > 0 {
		Feb = req.FormValue("Feb")
	}
	var Mar = ""
	if len(req.FormValue("Mar")) > 0 {
		Mar = req.FormValue("Mar")
	}
	var Apr = ""
	if len(req.FormValue("Apr")) > 0 {
		Apr = req.FormValue("Apr")
	}
	var Maj = ""
	if len(req.FormValue("Maj")) > 0 {
		Maj = req.FormValue("Maj")
	}
	var Jun = ""
	if len(req.FormValue("Jun")) > 0 {
		Jun = req.FormValue("Jun")
	}
	var Jul = ""
	if len(req.FormValue("Jul")) > 0 {
		Jul = req.FormValue("Jul")
	}
	var Aug = ""
	if len(req.FormValue("Aug")) > 0 {
		Aug = req.FormValue("Aug")
	}
	var Sep = ""
	if len(req.FormValue("Sep")) > 0 {
		Sep = req.FormValue("Sep")
	}
	var Okt = ""
	if len(req.FormValue("Okt")) > 0 {
		Okt = req.FormValue("Okt")
	}
	var Nov = ""
	if len(req.FormValue("Nov")) > 0 {
		Nov = req.FormValue("Nov")
	}
	var Dec = ""
	if len(req.FormValue("Dec")) > 0 {
		Dec = req.FormValue("Dec")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := db.ExecContext(ctx,
		`UPDATE Budget SET Typ = ?, Inkomst = ?, HurOfta = ?, StartMånad = ?, Jan = ?, Feb = ?, Mar = ?, Apr = ?, Maj = ?, Jun = ?, Jul = ?, Aug = ?, Sep = ?, Okt = ?, Nov = ?, Dec = ? WHERE (Löpnr=?)`,
		Typ,
		Inkomst,
		HurOfta,
		StartMånad,
		strings.ReplaceAll(Jan, ".", ","),
		strings.ReplaceAll(Feb, ".", ","),
		strings.ReplaceAll(Mar, ".", ","),
		strings.ReplaceAll(Apr, ".", ","),
		strings.ReplaceAll(Maj, ".", ","),
		strings.ReplaceAll(Jun, ".", ","),
		strings.ReplaceAll(Jul, ".", ","),
		strings.ReplaceAll(Aug, ".", ","),
		strings.ReplaceAll(Sep, ".", ","),
		strings.ReplaceAll(Okt, ".", ","),
		strings.ReplaceAll(Nov, ".", ","),
		strings.ReplaceAll(Dec, ".", ","),
		lopnr)

	if err != nil {
		log.Fatal(err)
	}
	_, _ = fmt.Fprintf(w, "Budgetpost %s uppdaterad.<br>", Typ)
}

func hanteraBudget(w http.ResponseWriter, req *http.Request) {
	_, _ = fmt.Fprintf(w, "<html>\n")
	_, _ = fmt.Fprintf(w, "<head>\n")
	_, _ = fmt.Fprintf(w, "<style>\n")
	_, _ = fmt.Fprintf(w, "table,th,td { border: 1px solid black }\n")
	_, _ = fmt.Fprintf(w, "</style>\n")
	_, _ = fmt.Fprintf(w, "</head>\n")
	_, _ = fmt.Fprintf(w, "<body>\n")

	_, _ = fmt.Fprintf(w, "<h1>%s</h1>\n", currentDatabase)
	_, _ = fmt.Fprintf(w, "<h2>Budget</h2>\n")

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
	case "editform":
		editformBudget(w, lopnr, db)
	case "update":
		updateBudget(w, lopnr, req, db)
	default:
		fmt.Println("Okänd action: ", formaction)
	}
	printBudget(w, db)
	printBudgetFooter(w)
}

func antalBudgetposter(db *sql.DB) int {
	if db == nil {
		log.Println("antalBudgetposter db=nil")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	res1 := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM Budget`)

	var antal int

	err := res1.Scan(&antal)
	if err != nil {
		log.Fatal(err)
	}

	return antal
}

func getAllBudgetposter(db *sql.DB) [][2]string {
	if db == nil {
		log.Println("getAllBudgetposter db=nil")
	}

	res, err := db.Query("SELECT Typ,Inkomst FROM Budget ORDER BY Inkomst ASC, Typ ASC")

	if err != nil {
		log.Fatal(err)
	}

	var Typ []byte     // size 40
	var Inkomst []byte // size 1

	var result [][2]string
	for res.Next() {
		var record [2]string

		err = res.Scan(&Typ, &Inkomst)

		record[0] = toUtf8(Typ)
		record[1] = toUtf8(Inkomst)
		result = append(result, record)
	}
	return result
}
