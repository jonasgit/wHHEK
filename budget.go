//-*- coding: utf-8 -*-

package main

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/shopspring/decimal" // MIT License
)

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

// BudgetEditForm holds one row for the budget edit form (template html/budget_edit_form.html).
type BudgetEditForm struct {
	Lopnr          int
	Typ            string
	Inkomst        string
	HurOfta        string
	StartManad     string
	Jan            string
	Feb            string
	Mar            string
	Apr            string
	Maj            string
	Jun            string
	Jul            string
	Aug            string
	Sep            string
	Okt            string
	Nov            string
	Dec            string
	Kontrollnr     string
}

// BudgetPageView is passed to template "budget_page".
type BudgetPageView struct {
	Database      string
	UpdateMessage string
	Edit          *BudgetEditForm
	BudgetData    BudgetData
}

func loadBudgetData(db *sql.DB) BudgetData {
	res, err := db.Query("SELECT Löpnr,Typ,Inkomst,HurOfta,StartMånad,Jan,Feb,Mar,Apr,Maj,Jun,Jul,Aug,Sep,Okt,Nov,Dec,Kontrollnr FROM Budget ORDER BY Inkomst ASC, Typ ASC")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Close()

	var budgetData BudgetData

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

		budgetData.BudgetItems = append(budgetData.BudgetItems, item)
	}

	return budgetData
}

func loadBudgetEditForm(lopnr int, db *sql.DB) *BudgetEditForm {
	fmt.Println("editformBudget lopnr: ", lopnr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
	var Löpnr int

	err := db.QueryRowContext(ctx,
		`SELECT Löpnr,Typ,Inkomst,HurOfta,StartMånad,Jan,Feb,Mar,Apr,Maj,Jun,Jul,Aug,Sep,Okt,Nov,Dec,Kontrollnr FROM Budget WHERE (Löpnr=?)`, lopnr).Scan(&Löpnr, &Typ, &Inkomst, &HurOfta, &StartMånad, &Jan, &Feb, &Mar, &Apr, &Maj, &Jun, &Jul, &Aug, &Sep, &Okt, &Nov, &Dec, &Kontrollnr)
	if err != nil {
		log.Fatal(err)
	}

	knr := "null"
	if Kontrollnr != nil {
		knr = toUtf8(Kontrollnr)
	}

	return &BudgetEditForm{
		Lopnr:      Löpnr,
		Typ:        toUtf8(Typ),
		Inkomst:    toUtf8(Inkomst),
		HurOfta:    strconv.Itoa(int(HurOfta)),
		StartManad: toUtf8(StartMånad),
		Jan:        toUtf8(Jan),
		Feb:        toUtf8(Feb),
		Mar:        toUtf8(Mar),
		Apr:        toUtf8(Apr),
		Maj:        toUtf8(Maj),
		Jun:        toUtf8(Jun),
		Jul:        toUtf8(Jul),
		Aug:        toUtf8(Aug),
		Sep:        toUtf8(Sep),
		Okt:        toUtf8(Okt),
		Nov:        toUtf8(Nov),
		Dec:        toUtf8(Dec),
		Kontrollnr: knr,
	}
}

func updateBudget(lopnr int, req *http.Request, db *sql.DB) string {
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
	return fmt.Sprintf("Budgetpost %s uppdaterad.", Typ)
}

func hanteraBudget(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	formaction := req.FormValue("action")
	var lopnr = -1
	if len(req.FormValue("lopnr")) > 0 {
		lopnr, _ = strconv.Atoi(req.FormValue("lopnr"))
	}

	var updateMsg string
	var editForm *BudgetEditForm

	switch formaction {
	case "editform":
		editForm = loadBudgetEditForm(lopnr, db)
	case "update":
		updateMsg = updateBudget(lopnr, req, db)
	default:
		if formaction != "" {
			fmt.Println("Okänd action: ", formaction)
		}
	}

	page := BudgetPageView{
		Database:      currentDatabase,
		UpdateMessage: updateMsg,
		Edit:          editForm,
		BudgetData:    loadBudgetData(db),
	}

	tmpl, err := template.ParseFS(htmlTemplates,
		"html/budget_page.html",
		"html/budget_edit_form.html",
		"html/budget_table.html",
		"html/budget_footer.html",
	)
	if err != nil {
		log.Fatal(err)
	}
	err = tmpl.ExecuteTemplate(w, "budget_page", page)
	if err != nil {
		log.Fatal(err)
	}
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
