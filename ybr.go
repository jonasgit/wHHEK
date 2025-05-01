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
	"time"

	"github.com/shopspring/decimal" // MIT License
)

type YBRkategoriType struct {
	Name    string
	Beloppy string
	BeloppB string
	RegP    string
	ProgP   string
}

//go:embed html/ybr.html
var htmlybr string

type YBRData struct {
	CurrentYear       string
	CurrentDay        string
	CurrentDayPercent string
	Inkomster         []YBRkategoriType
	SumIn             string
	Utgifter          []YBRkategoriType
	SumUt             string
}

func sumKatToday(kat string, selectYear int, intyp bool) decimal.Decimal {
	result := decimal.NewFromInt32(0)

	year := decimal.NewFromInt(int64(selectYear)).String()
	startstring := year + "-01-01"
	now := time.Now()
	day := now.Day()
	month := now.Month()
	endstring := fmt.Sprintf("%s-%02d-%02d", year, month, day)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var err error
	var res *sql.Rows

	if intyp {
		res, err = db.QueryContext(ctx,
			`SELECT Belopp from transaktioner
  where (Typ = ? or Typ = ?) and Vad = ? and Datum >= ? and Datum <= ?`, "Insättning", "Fast Inkomst", kat, startstring, endstring)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		res, err = db.QueryContext(ctx,
			`SELECT Belopp from transaktioner
  where (Typ = ? or Typ = ?) and Vad = ? and Datum >= ? and Datum <= ?`, "Inköp", "Fast Utgift", kat, startstring, endstring)
		if err != nil {
			log.Fatal(err)
		}
	}

	var amount []byte // BCD / Decimal Precision 19

	for res.Next() {
		err = res.Scan(&amount)

		decamount, err := decimal.NewFromString(toUtf8(amount))
		if err != nil {
			log.Fatal(err)
		}

		result = result.Add(decamount)
	}
	res.Close()
	return result
}

func hanteraYBR(w http.ResponseWriter, req *http.Request) {
	log.Println("Func hanteraYBR")

	err := req.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	now := time.Now()
	selectYear := now.Year()
	yrday := now.YearDay()
	yrdayp := int((float64(yrday) / 365.0) * 100) // Bryr mig inte om skottår nu

	showNulls := false

	var inkomster []YBRkategoriType
	var utgifter []YBRkategoriType

	decimal.DivisionPrecision = 2

	decZero := decimal.NewFromInt(0)
	dec100 := decimal.NewFromInt(100)
	dec365 := decimal.NewFromInt(365)
	decDay := decimal.NewFromInt(int64(yrday))
	katin := getTypeInNames()
	for _, kat := range katin {
		belopp := sumKatToday(kat, selectYear, true)
		sum := Dec2Str(belopp)
		budgetsum := getKatYearSum(db, kat)
		var budgetproc string
		var prognos string
		if decZero.Equal(budgetsum) {
			budgetproc = "∞"
			prognos = "∞"
		} else {
			budgetproc = belopp.Div(budgetsum).Mul(dec100).String()
			beloppdag := belopp.Div(decDay).Mul(dec365).Div(budgetsum).Mul(dec100)
			prognos = beloppdag.String()

		}
		if showNulls || (!belopp.Equal(decZero)) {
			inkomster = append(inkomster, YBRkategoriType{kat, sum, Dec2Str(budgetsum), budgetproc, prognos})
		}
	}
	katut := getTypeOutNames()
	for _, kat := range katut {
		belopp := sumKatToday(kat, selectYear, false)
		sum := Dec2Str(belopp)
		budgetsum := getKatYearSum(db, kat)
		var budgetproc string
		var prognos string
		if decZero.Equal(budgetsum) {
			budgetproc = "∞"
			prognos = "∞"
		} else {
			budgetproc = belopp.Div(budgetsum).Mul(dec100).String()
			beloppdag := belopp.Div(decDay).Mul(dec365).Div(budgetsum).Mul(dec100)
			prognos = beloppdag.String()
		}
		if showNulls || (!belopp.Equal(decZero)) {
			utgifter = append(utgifter, YBRkategoriType{kat, sum, Dec2Str(budgetsum), budgetproc, prognos})
		}
	}

	log.Println("Func hanteraYBR year:", strconv.Itoa(selectYear))

	tmpl1 := template.New("wHHEK Årsstatus")
	tmpl1, _ = tmpl1.Parse(htmlybr)
	data := YBRData{
		CurrentYear:       strconv.Itoa(selectYear),
		CurrentDay:        strconv.Itoa(yrday),
		CurrentDayPercent: strconv.Itoa(yrdayp),
		Inkomster:         inkomster,
		Utgifter:          utgifter,
	}
	_ = tmpl1.Execute(w, data)
}
