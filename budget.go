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
)

func printBudget(w http.ResponseWriter, db *sql.DB) {
	res, err := db.Query("SELECT Löpnr,Typ,Inkomst,HurOfta,StartMånad,Jan,Feb,Mar,Apr,Maj,Jun,Jul,Aug,Sep,Okt,Nov,Dec,Kontrollnr FROM Budget ORDER BY Inkomst ASC, Typ ASC")

	if err != nil {
		log.Fatal(err)
		os.Exit(2)
	}

	var Typ []byte  // size 40
	var Inkomst []byte  // size 1
	var HurOfta int16 // SmallInt
	var StartMånad []byte  // size 10
	var Jan []byte  // BCD / Decimal Precision 19
	var Feb []byte  // BCD / Decimal Precision 19
	var Mar []byte  // BCD / Decimal Precision 19
	var Apr []byte  // BCD / Decimal Precision 19
	var Maj []byte  // BCD / Decimal Precision 19
	var Jun []byte  // BCD / Decimal Precision 19
	var Jul []byte  // BCD / Decimal Precision 19
	var Aug []byte  // BCD / Decimal Precision 19
	var Sep []byte  // BCD / Decimal Precision 19
	var Okt []byte  // BCD / Decimal Precision 19
	var Nov []byte  // BCD / Decimal Precision 19
	var Dec []byte  // BCD / Decimal Precision 19
	var Kontrollnr []byte  //int32 // Integer
	var Löpnr int  // autoinc Primary Key, index

	fmt.Fprintf(w, "<table style=\"width:100%%\"><tr>")
	fmt.Fprintf(w, "<th>Löpnr</th>")
	fmt.Fprintf(w, "<th>Typ</th>")
	fmt.Fprintf(w, "<th>Inkomst</th>")
	fmt.Fprintf(w, "<th>HurOfta</th>")
	fmt.Fprintf(w, "<th>StartMånad</th>")
	fmt.Fprintf(w, "<th>Jan</th>")
	fmt.Fprintf(w, "<th>Feb</th>")
	fmt.Fprintf(w, "<th>Mar</th>")
	fmt.Fprintf(w, "<th>Apr</th>")
	fmt.Fprintf(w, "<th>Maj</th>")
	fmt.Fprintf(w, "<th>Jun</th>")
	fmt.Fprintf(w, "<th>Jul</th>")
	fmt.Fprintf(w, "<th>Aug</th>")
	fmt.Fprintf(w, "<th>Sep</th>")
	fmt.Fprintf(w, "<th>Okt</th>")
	fmt.Fprintf(w, "<th>Nov</th>")
	fmt.Fprintf(w, "<th>Dec</th>")
	fmt.Fprintf(w, "<th>Kontrollnr</th>")
	fmt.Fprintf(w, "<th>Redigera</th><th>Radera</th>\n")
	for res.Next() {
		err = res.Scan(&Löpnr,&Typ,&Inkomst,&HurOfta,&StartMånad,&Jan,&Feb,&Mar,&Apr,&Maj,&Jun,&Jul,&Aug,&Sep,&Okt,&Nov,&Dec,&Kontrollnr)

		fmt.Fprintf(w, "<tr>")
		fmt.Fprintf(w, "<td>%d</td>", Löpnr)
		fmt.Fprintf(w, "<td>%s</td>", toUtf8(Typ))
		fmt.Fprintf(w, "<td>%s</td>", toUtf8(Inkomst))
		fmt.Fprintf(w, "<td>%s</td>", strconv.Itoa(int(HurOfta)))
		fmt.Fprintf(w, "<td>%s</td>", toUtf8(StartMånad))
		fmt.Fprintf(w, "<td>%s</td>", toUtf8(Jan))
		fmt.Fprintf(w, "<td>%s</td>", toUtf8(Feb))
		fmt.Fprintf(w, "<td>%s</td>", toUtf8(Mar))
		fmt.Fprintf(w, "<td>%s</td>", toUtf8(Apr))
		fmt.Fprintf(w, "<td>%s</td>", toUtf8(Maj))
		fmt.Fprintf(w, "<td>%s</td>", toUtf8(Jun))
		fmt.Fprintf(w, "<td>%s</td>", toUtf8(Jul))
		fmt.Fprintf(w, "<td>%s</td>", toUtf8(Aug))
		fmt.Fprintf(w, "<td>%s</td>", toUtf8(Sep))
		fmt.Fprintf(w, "<td>%s</td>", toUtf8(Okt))
		fmt.Fprintf(w, "<td>%s</td>", toUtf8(Nov))
		fmt.Fprintf(w, "<td>%s</td>", toUtf8(Dec))
		if Kontrollnr != nil {
			fmt.Fprintf(w, "<td>%s</td>", toUtf8(Kontrollnr))
		} else {
			fmt.Fprintf(w, "<td>%s</td>", "null")
		}

		fmt.Fprintf(w, "<td><form method=\"POST\" action=\"/budget\"><input type=\"hidden\" id=\"lopnr\" name=\"lopnr\" value=\"%d\"><input type=\"hidden\" id=\"action\" name=\"action\" value=\"editform\"><input type=\"submit\" value=\"Redigera\"></form></td>\n", Löpnr)
		fmt.Fprintf(w, "<td><form method=\"POST\" action=\"/budget\"><input type=\"hidden\" id=\"lopnr\" name=\"lopnr\" value=\"%d\"><input type=\"hidden\" id=\"action\" name=\"action\" value=\"radera\"><input type=\"checkbox\" id=\"OK\" name=\"OK\" required><label for=\"OK\">OK</label><input type=\"submit\" value=\"Radera\"></form></td></tr>\n", Löpnr)
	}
	fmt.Fprintf(w, "</table>\n")

	fmt.Fprintf(w, "<form method=\"POST\" action=\"/budget\"><input type=\"hidden\" id=\"action\" name=\"action\" value=\"addform\"><input type=\"submit\" value=\"Ny budgetpost\"></form>\n")
}

func printBudgetFooter(w http.ResponseWriter, db *sql.DB) {
	fmt.Fprintf(w, "<a href=\"summary\">Översikt</a>\n")
	fmt.Fprintf(w, "</body>\n")
	fmt.Fprintf(w, "</html>\n")
}

func PrintEditCellText(w http.ResponseWriter, label string, title string, value string) {
	fmt.Fprintf(w, "<label for=\"%s\">%s:</label>", label, title)
	fmt.Fprintf(w, "<input type=\"text\" id=\"%s\" name=\"%s\" value=\"%s\" >", label, label, value)
}

func editformBudget(w http.ResponseWriter, lopnr int, db *sql.DB) {
	fmt.Println("editformBudget lopnr: ", lopnr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	res1 := db.QueryRowContext(ctx,
		`SELECT Löpnr,Typ,Inkomst,HurOfta,StartMånad,Jan,Feb,Mar,Apr,Maj,Jun,Jul,Aug,Sep,Okt,Nov,Dec,Kontrollnr FROM Budget WHERE (Löpnr=?)`, lopnr)

	var Typ []byte  // size 40
	var Inkomst []byte  // size 1
	var HurOfta int16 // SmallInt
	var StartMånad []byte  // size 10
	var Jan []byte  // BCD / Decimal Precision 19
	var Feb []byte  // BCD / Decimal Precision 19
	var Mar []byte  // BCD / Decimal Precision 19
	var Apr []byte  // BCD / Decimal Precision 19
	var Maj []byte  // BCD / Decimal Precision 19
	var Jun []byte  // BCD / Decimal Precision 19
	var Jul []byte  // BCD / Decimal Precision 19
	var Aug []byte  // BCD / Decimal Precision 19
	var Sep []byte  // BCD / Decimal Precision 19
	var Okt []byte  // BCD / Decimal Precision 19
	var Nov []byte  // BCD / Decimal Precision 19
	var Dec []byte  // BCD / Decimal Precision 19
	var Kontrollnr []byte  //int32 // Integer
	var Löpnr int  // autoinc Primary Key, index

	err := res1.Scan(&Löpnr,&Typ,&Inkomst,&HurOfta,&StartMånad,&Jan,&Feb,&Mar,&Apr,&Maj,&Jun,&Jul,&Aug,&Sep,&Okt,&Nov,&Dec,&Kontrollnr)
	if err != nil {
		log.Fatal(err)
		os.Exit(2)
	}

	fmt.Fprintf(w, "Redigera budgetpost<br>")
	fmt.Fprintf(w, "<form method=\"POST\" action=\"/budget\">")

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

	fmt.Fprintf(w, "<input type=\"hidden\" id=\"lopnr\" name=\"lopnr\" value=\"%d\">", Löpnr)
	fmt.Fprintf(w, "<input type=\"hidden\" id=\"action\" name=\"action\" value=\"update\">")
	fmt.Fprintf(w, "<input type=\"submit\" value=\"Uppdatera\">")
	fmt.Fprintf(w, "</form>\n")
	fmt.Fprintf(w, "<p>\n")
}


func updateBudget(w http.ResponseWriter, lopnr int, req *http.Request, db *sql.DB) {
	fmt.Println("updateBudget lopnr: ", lopnr)

	var Typ string = ""
	if len(req.FormValue("Typ")) > 0 {
		Typ = req.FormValue("Typ")
	}
	var Inkomst string = ""
	if len(req.FormValue("Inkomst")) > 0 {
		Inkomst = req.FormValue("Inkomst")
	}
	var HurOfta string = ""
	if len(req.FormValue("HurOfta")) > 0 {
		HurOfta = req.FormValue("HurOfta")
	}
	var StartMånad string = ""
	if len(req.FormValue("StartMånad")) > 0 {
		StartMånad = req.FormValue("StartMånad")
	}
	var Jan string = ""
	if len(req.FormValue("Jan")) > 0 {
		Jan = req.FormValue("Jan")
	}
	var Feb string = ""
	if len(req.FormValue("Feb")) > 0 {
		Feb = req.FormValue("Feb")
	}
	var Mar string = ""
	if len(req.FormValue("Mar")) > 0 {
		Mar = req.FormValue("Mar")
	}
	var Apr string = ""
	if len(req.FormValue("Apr")) > 0 {
		Apr = req.FormValue("Apr")
	}
	var Maj string = ""
	if len(req.FormValue("Maj")) > 0 {
		Maj = req.FormValue("Maj")
	}
	var Jun string = ""
	if len(req.FormValue("Jun")) > 0 {
		Jun = req.FormValue("Jun")
	}
	var Jul string = ""
	if len(req.FormValue("Jul")) > 0 {
		Jul = req.FormValue("Jul")
	}
	var Aug string = ""
	if len(req.FormValue("Aug")) > 0 {
		Aug = req.FormValue("Aug")
	}
	var Sep string = ""
	if len(req.FormValue("Sep")) > 0 {
		Sep = req.FormValue("Sep")
	}
	var Okt string = ""
	if len(req.FormValue("Okt")) > 0 {
		Okt = req.FormValue("Okt")
	}
	var Nov string = ""
	if len(req.FormValue("Nov")) > 0 {
		Nov = req.FormValue("Nov")
	}
	var Dec string = ""
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
		os.Exit(2)
	}
	fmt.Fprintf(w, "Budgetpost %s uppdaterad.<br>", Typ)
}

func hanteraBudget(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "<html>\n")
	fmt.Fprintf(w, "<head>\n")
	fmt.Fprintf(w, "<style>\n")
	fmt.Fprintf(w, "table,th,td { border: 1px solid black }\n")
	fmt.Fprintf(w, "</style>\n")
	fmt.Fprintf(w, "</head>\n")
	fmt.Fprintf(w, "<body>\n")

	fmt.Fprintf(w, "<h1>%s</h1>\n", currentDatabase)
	fmt.Fprintf(w, "<h2>Budget</h2>\n")

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
	case "editform":
		editformBudget(w, lopnr, db)
	case "update":
		updateBudget(w, lopnr, req, db)
	default:
		fmt.Println("Okänd action: ", formaction)
	}
	printBudget(w, db)
	printBudgetFooter(w, db)
}
