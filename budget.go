//-*- coding: utf-8 -*-

package main

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/shopspring/decimal" // MIT License
)

//go:embed html/budget_table.html
var budgetTableTemplate embed.FS

type BudgetItem struct {
	Lopnr      int    // autoinc Primary Key, index
	Typ        string // size 40
	Inkomst    string // size 1
	HurOfta    string // SmallInt
	StartManad string // size 10
	Jan        string // BCD / Decimal Precision 19
	Feb        string // BCD / Decimal Precision 19
	Mar        string // BCD / Decimal Precision 19
	Apr        string // BCD / Decimal Precision 19
	Maj        string // BCD / Decimal Precision 19
	Jun        string // BCD / Decimal Precision 19
	Jul        string // BCD / Decimal Precision 19
	Aug        string // BCD / Decimal Precision 19
	Sep        string // BCD / Decimal Precision 19
	Okt        string // BCD / Decimal Precision 19
	Nov        string // BCD / Decimal Precision 19
	Dec        string // BCD / Decimal Precision 19
	Kontrollnr string // int32 // Integer
}

type BudgetData struct {
	BudgetItems []BudgetItem
}

func printBudget(w http.ResponseWriter, db *sql.DB) {
	res, err := db.Query("SELECT Löpnr,Typ,Inkomst,HurOfta,StartMånad,Jan,Feb,Mar,Apr,Maj,Jun,Jul,Aug,Sep,Okt,Nov,Dec,Kontrollnr FROM Budget ORDER BY Inkomst ASC, Typ ASC")

	if err != nil {
		log.Fatal(err)
	}

	var budgetData BudgetData
	var budgetItems []BudgetItem

	for res.Next() {
		var item BudgetItem
		var Typ []byte
		var Inkomst []byte
		var HurOfta int16
		var StartMånad []byte
		var Jan []byte
		var Feb []byte
		var Mar []byte
		var Apr []byte
		var Maj []byte
		var Jun []byte
		var Jul []byte
		var Aug []byte
		var Sep []byte
		var Okt []byte
		var Nov []byte
		var Dec []byte
		var Kontrollnr []byte

		err = res.Scan(&item.Lopnr, &Typ, &Inkomst, &HurOfta, &StartMånad, &Jan, &Feb, &Mar, &Apr, &Maj, &Jun, &Jul, &Aug, &Sep, &Okt, &Nov, &Dec, &Kontrollnr)
		if err != nil {
			log.Fatal(err)
		}

		item.Typ = toUtf8(Typ)
		item.Inkomst = toUtf8(Inkomst)
		item.HurOfta = strconv.Itoa(int(HurOfta))
		item.StartManad = toUtf8(StartMånad)
		item.Jan = toUtf8(Jan)
		item.Feb = toUtf8(Feb)
		item.Mar = toUtf8(Mar)
		item.Apr = toUtf8(Apr)
		item.Maj = toUtf8(Maj)
		item.Jun = toUtf8(Jun)
		item.Jul = toUtf8(Jul)
		item.Aug = toUtf8(Aug)
		item.Sep = toUtf8(Sep)
		item.Okt = toUtf8(Okt)
		item.Nov = toUtf8(Nov)
		item.Dec = toUtf8(Dec)
		if Kontrollnr != nil {
			item.Kontrollnr = toUtf8(Kontrollnr)
		} else {
			item.Kontrollnr = "null"
		}

		budgetItems = append(budgetItems, item)
	}

	budgetData.BudgetItems = budgetItems

	tmplContent, err := budgetTableTemplate.ReadFile("html/budget_table.html")
	if err != nil {
		log.Fatal(err)
	}

	tmpl, err := template.New("budget_table").Parse(string(tmplContent))
	if err != nil {
		log.Fatal(err)
	}

	err = tmpl.ExecuteTemplate(w, "budget_table", budgetData)
	if err != nil {
		log.Fatal(err)
	}

	res.Close()
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

	err := db.QueryRowContext(ctx,
		`SELECT Löpnr,Typ,Inkomst,HurOfta,StartMånad,Jan,Feb,Mar,Apr,Maj,Jun,Jul,Aug,Sep,Okt,Nov,Dec,Kontrollnr FROM Budget WHERE (Löpnr=?)`, lopnr).Scan(&Löpnr, &Typ, &Inkomst, &HurOfta, &StartMånad, &Jan, &Feb, &Mar, &Apr, &Maj, &Jun, &Jul, &Aug, &Sep, &Okt, &Nov, &Dec, &Kontrollnr)

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
		lopnr, _ = strconv.Atoi(req.FormValue("lopnr"))
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

	var antal int

	err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM Budget`).Scan(&antal)

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

		_ = res.Scan(&Typ, &Inkomst)

		record[0] = toUtf8(Typ)
		record[1] = toUtf8(Inkomst)
		result = append(result, record)
	}
	res.Close()
	return result
}

func addDecStr(v1 decimal.Decimal, v2 string) decimal.Decimal {

	tot, err := decimal.NewFromString(v2)
	if err != nil {
		log.Println("addDecStr strasig decimal sträng: ", v2)
		panic(err)
	}
	return v1.Add(tot)
}

// Returnera årssumman för en specifik kategori
func getKatYearSum(db *sql.DB, kategori string) decimal.Decimal {
	var ret decimal.Decimal

	if db == nil {
		log.Println("getKatYearSum db=nil")
		return ret
	}

	var Jan []byte // BCD / Decimal Precision 19
	var Feb []byte // BCD / Decimal Precision 19
	var Mar []byte // BCD / Decimal Precision 19
	var Apr []byte // BCD / Decimal Precision 19
	var Maj []byte // BCD / Decimal Precision 19
	var Jun []byte // BCD / Decimal Precision 19
	var Jul []byte // BCD / Decimal Precision 19
	var Aug []byte // BCD / Decimal Precision 19
	var Sep []byte // BCD / Decimal Precision 19
	var Okt []byte // BCD / Decimal Precision 19
	var Nov []byte // BCD / Decimal Precision 19
	var Dec []byte // BCD / Decimal Precision 19

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := db.QueryRowContext(ctx,
		`SELECT Jan,Feb,Mar,Apr,Maj,Jun,Jul,Aug,Sep,Okt,Nov,Dec FROM Budget WHERE (Typ=?)`, kategori).Scan(&Jan, &Feb, &Mar, &Apr, &Maj, &Jun, &Jul, &Aug, &Sep, &Okt, &Nov, &Dec)

	if err != nil {
		log.Fatal(err)
	}

	ret = addDecStr(ret, toUtf8(Jan))
	ret = addDecStr(ret, toUtf8(Feb))
	ret = addDecStr(ret, toUtf8(Mar))
	ret = addDecStr(ret, toUtf8(Apr))
	ret = addDecStr(ret, toUtf8(Maj))
	ret = addDecStr(ret, toUtf8(Jun))
	ret = addDecStr(ret, toUtf8(Jul))
	ret = addDecStr(ret, toUtf8(Aug))
	ret = addDecStr(ret, toUtf8(Sep))
	ret = addDecStr(ret, toUtf8(Okt))
	ret = addDecStr(ret, toUtf8(Nov))
	ret = addDecStr(ret, toUtf8(Dec))

	return ret
}
