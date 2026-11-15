//-*- coding: utf-8 -*-

package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

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
	Lopnr      int
	Typ        string
	Inkomst    string
	HurOfta    string
	StartManad string
	Jan        string
	Feb        string
	Mar        string
	Apr        string
	Maj        string
	Jun        string
	Jul        string
	Aug        string
	Sep        string
	Okt        string
	Nov        string
	Dec        string
	Kontrollnr string
}

// BudgetPageView is passed to template "budget_page".
type BudgetPageView struct {
	Database      string
	UpdateMessage string
	ErrorMessage  string
	Edit          *BudgetEditForm
	BudgetData    BudgetData
}

func loadBudgetData(db *sql.DB) (BudgetData, error) {
	if db == nil {
		return BudgetData{}, fmt.Errorf("ingen databas är öppen")
	}

	res, err := db.Query("SELECT Löpnr,Typ,Inkomst,HurOfta,StartMånad,Jan,Feb,Mar,Apr,Maj,Jun,Jul,Aug,Sep,Okt,Nov,Dec,Kontrollnr FROM Budget ORDER BY Inkomst ASC, Typ ASC")
	if err != nil {
		return BudgetData{}, fmt.Errorf("kunde inte läsa budgettabellen: %w", err)
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
			res.Close()
			return BudgetData{}, fmt.Errorf("kunde inte läsa budgetrad: %w", err)
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

	return budgetData, nil
}

func loadBudgetEditForm(lopnr int, db *sql.DB) (*BudgetEditForm, error) {
	log.Printf("budget editform: lopnr=%d", lopnr)

	if db == nil {
		return nil, fmt.Errorf("ingen databas är öppen")
	}
	if lopnr < 0 {
		return nil, fmt.Errorf("ogiltigt löpnr")
	}

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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("budgetpost med löpnr %d hittades inte", lopnr)
		}
		return nil, fmt.Errorf("kunde inte läsa budgetpost: %w", err)
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
	}, nil
}

func updateBudget(lopnr int, req *http.Request, db *sql.DB) (string, error) {
	log.Printf("budget update: lopnr=%d", lopnr)

	if db == nil {
		return "", fmt.Errorf("ingen databas är öppen")
	}
	if lopnr < 0 {
		return "", fmt.Errorf("ogiltigt löpnr")
	}

	c := newFormCheck(req)
	Typ := c.RequiredText("Typ", "Typ", maxLenTyp)
	Inkomst := c.OneOf("Inkomst", "Inkomst", "J", "N")
	HurOfta := c.Int("HurOfta", "Hur ofta", 0, 12, true, 0)
	StartMånad := c.Month("StartMånad", "Startmånad")
	var belopp [12]string
	for i, m := range monthNames {
		amount, _ := c.Amount(m, m, false)
		belopp[i] = AmountDec2DBStr(amount)
	}
	if err := c.Err(); err != nil {
		return "", fmt.Errorf("budgetposten uppdaterades inte: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := db.ExecContext(ctx,
		`UPDATE Budget SET Typ = ?, Inkomst = ?, HurOfta = ?, StartMånad = ?, Jan = ?, Feb = ?, Mar = ?, Apr = ?, Maj = ?, Jun = ?, Jul = ?, Aug = ?, Sep = ?, Okt = ?, Nov = ?, Dec = ? WHERE (Löpnr=?)`,
		Typ,
		Inkomst,
		HurOfta,
		StartMånad,
		belopp[0],
		belopp[1],
		belopp[2],
		belopp[3],
		belopp[4],
		belopp[5],
		belopp[6],
		belopp[7],
		belopp[8],
		belopp[9],
		belopp[10],
		belopp[11],
		lopnr)

	if err != nil {
		return "", fmt.Errorf("kunde inte uppdatera budgetpost: %w", err)
	}
	return fmt.Sprintf("Budgetpost %s uppdaterad.", Typ), nil
}

func renderBudgetPage(w http.ResponseWriter, page BudgetPageView) {
	tmpl, err := template.ParseFS(htmlTemplates,
		"html/budget_page.html",
		"html/budget_edit_form.html",
		"html/budget_table.html",
		"html/budget_footer.html",
	)
	if err != nil {
		log.Printf("budget: parse templates: %v", err)
		http.Error(w, "Kunde inte ladda budgetsidan.", http.StatusInternalServerError)
		return
	}
	if err := tmpl.ExecuteTemplate(w, "budget_page", page); err != nil {
		log.Printf("budget: render page: %v", err)
		http.Error(w, "Kunde inte visa budgetsidan.", http.StatusInternalServerError)
	}
}

func hanteraBudget(w http.ResponseWriter, req *http.Request) {
	log.Printf("budget: %s action=%q lopnr=%q", req.Method, req.FormValue("action"), req.FormValue("lopnr"))

	page := BudgetPageView{
		Database: currentDatabase,
	}

	if err := req.ParseForm(); err != nil {
		log.Printf("budget: parse form: %v", err)
		page.ErrorMessage = "Kunde inte läsa formulärdata."
		renderBudgetPage(w, page)
		return
	}

	formaction := req.FormValue("action")
	lopnr := -1
	if lopnrStr := req.FormValue("lopnr"); lopnrStr != "" {
		var err error
		lopnr, err = strconv.Atoi(lopnrStr)
		if err != nil {
			log.Printf("budget: ogiltigt lopnr %q: %v", lopnrStr, err)
			page.ErrorMessage = fmt.Sprintf("Ogiltigt löpnr: %s", lopnrStr)
		}
	}

	var editForm *BudgetEditForm

	switch formaction {
	case "editform":
		if page.ErrorMessage == "" {
			var err error
			editForm, err = loadBudgetEditForm(lopnr, db)
			if err != nil {
				log.Printf("budget: editform lopnr=%d: %v", lopnr, err)
				page.ErrorMessage = err.Error()
			}
		}
	case "update":
		if page.ErrorMessage == "" {
			msg, err := updateBudget(lopnr, req, db)
			if err != nil {
				log.Printf("budget: update lopnr=%d: %v", lopnr, err)
				page.ErrorMessage = err.Error()
			} else {
				page.UpdateMessage = msg
			}
		}
	default:
		if formaction != "" {
			log.Printf("budget: okänd action %q", formaction)
			page.ErrorMessage = fmt.Sprintf("Okänd åtgärd: %s", formaction)
		}
	}

	page.Edit = editForm

	budgetData, err := loadBudgetData(db)
	if err != nil {
		log.Printf("budget: load data: %v", err)
		if page.ErrorMessage == "" {
			page.ErrorMessage = err.Error()
		}
	} else {
		page.BudgetData = budgetData
	}

	renderBudgetPage(w, page)
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

// addDecStr adderar ett belopp från databasen. Både "100.50" och "100,50"
// accepteras (äldre SQLite-databaser kan innehålla kommatecken). Ett belopp som
// inte går att tolka loggas och räknas som 0 i stället för att krascha sidan.
func addDecStr(v1 decimal.Decimal, v2 string) decimal.Decimal {
	tot, err := decimal.NewFromString(SanitizeAmount(v2))
	if err != nil {
		log.Println("addDecStr trasig decimal sträng: ", v2, err)
		return v1
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
