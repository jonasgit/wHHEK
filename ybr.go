//-*- coding: utf-8 -*-

package main

import (
	_ "embed"
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
	Inkomster     []YBRkategoriType
	SumIn         string
	Utgifter      []YBRkategoriType
	SumUt         string
}

func hanteraYBR(w http.ResponseWriter, req *http.Request) {
	log.Println("Func hanteraYBR")

	err := req.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	formYear := 0
	selectYear := time.Now().Year()
	if len(req.FormValue("formYear")) > 3 {
		formYear, err = strconv.Atoi(req.FormValue("formYear"))
		if formYear > 1900 && formYear < 2200 {
			selectYear = formYear
		}
	}

	showNulls := false
	sumin := decimal.NewFromInt32(0)
	sumut := decimal.NewFromInt32(0)
	sumutplats := decimal.NewFromInt32(0)

	var inkomster []YBRkategoriType
	var utgifter []YBRkategoriType

	decZero := decimal.NewFromInt(0)
	katin := getTypeInNames()
	for _, kat := range katin {
		belopp := sumKatYear(kat, selectYear, true)
		sum := Dec2Str(belopp)
		decimal.DivisionPrecision = 2
		if showNulls || (!belopp.Equal(decZero)) {
			inkomster = append(inkomster, YBRkategoriType{kat, sum, "TBD", "TBD", "TBD"})
		}

		sumin = sumin.Add(belopp)
	}
	katut := getTypeOutNames()
	for _, kat := range katut {
		belopp := sumKatYear(kat, selectYear, false)
		sum := Dec2Str(belopp)
		if showNulls || (!belopp.Equal(decZero)) {
			utgifter = append(utgifter, YBRkategoriType{kat, sum, "TBD", "TBD", "TBD"})
		}

		sumut = sumut.Add(belopp)
	}

	var date []byte // size 10
	err = db.QueryRow("SELECT MIN(Datum) FROM Transaktioner").Scan(&date)
	firstYear, err := strconv.Atoi(toUtf8(date)[0:4])
	err = db.QueryRow("SELECT MAX(Datum) FROM Transaktioner").Scan(&date)
	lastYear, err := strconv.Atoi(toUtf8(date)[0:4])

	var years []string
	for i := firstYear; i <= lastYear; i++ {
		years = append(years, strconv.Itoa(i))
	}

	log.Println("Func hanteraYBR year:", strconv.Itoa(selectYear))
	log.Println("Func hanteraYBR sumin:", sumin.String())
	log.Println("Func hanteraYBR sumut:", sumut.String())
	log.Println("Func hanteraYBR sumutplats:", sumutplats.String())

	tmpl1 := template.New("wHHEK Årsstatus")
	tmpl1, _ = tmpl1.Parse(htmlybr)
	data := YBRData{
		CurrentYear:   "TBD",
		CurrentDay:    "TBD",
		CurrentDayPercent: "TBD",
		Inkomster:     inkomster,
		SumIn:         Dec2Str(sumin),
		Utgifter:      utgifter,
		SumUt:         Dec2Str(sumut),
	}
	_ = tmpl1.Execute(w, data)
}
