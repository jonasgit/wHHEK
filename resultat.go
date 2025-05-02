//-*- coding: utf-8 -*-

package main

import (
	"context"
	"database/sql"
	_ "embed"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/shopspring/decimal" // MIT License
)

type kategoriType struct {
	Name    string
	Beloppy string
	Beloppm string
}

//go:embed html/resultat.html
var htmlresultat string

type ResultatData struct {
	SelectedYear  string
	Inkomster     []kategoriType
	SumIn         string
	Utgifter      []kategoriType
	SumUt         string
	UtgifterPlats []kategoriType
	SumUtPlats    string
	Years         []string
}

func hanteraYResult(w http.ResponseWriter, req *http.Request) {
	log.Println("Func hanteraYResult")

	err := req.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	formYear := 0
	selectYear := time.Now().Year()
	if len(req.FormValue("formYear")) > 3 {
		formYear, _ = strconv.Atoi(req.FormValue("formYear"))
		if formYear > 1900 && formYear < 2200 {
			selectYear = formYear
		}
	}

	showNulls := false
	sumin := decimal.NewFromInt32(0)
	sumut := decimal.NewFromInt32(0)
	sumutplats := decimal.NewFromInt32(0)

	var inkomster []kategoriType
	var utgifter []kategoriType
	var utgifterplats []kategoriType

	decZero := decimal.NewFromInt(0)
	dec12 := decimal.NewFromInt(12)
	katin := getTypeInNames()
	for _, kat := range katin {
		belopp := sumKatYear(kat, selectYear, true)
		sum := Dec2Str(belopp)
		decimal.DivisionPrecision = 2
		beloppm := belopp.Div(dec12)
		summ := Dec2Str(beloppm)
		if showNulls || (!belopp.Equal(decZero)) {
			inkomster = append(inkomster, kategoriType{kat, sum, summ})
		}

		sumin = sumin.Add(belopp)
	}
	katut := getTypeOutNames()
	for _, kat := range katut {
		belopp := sumKatYear(kat, selectYear, false)
		sum := Dec2Str(belopp)
		beloppm := belopp.Div(dec12)
		summ := Dec2Str(beloppm)
		if showNulls || (!belopp.Equal(decZero)) {
			utgifter = append(utgifter, kategoriType{kat, sum, summ})
		}

		sumut = sumut.Add(belopp)
	}
	places := getPlaceNames()
	for _, place := range places {
		belopp := sumPlaceYear(place, selectYear)
		sum := Dec2Str(belopp)
		beloppm := belopp.Div(dec12)
		summ := Dec2Str(beloppm)
		if showNulls || (!belopp.Equal(decZero)) {
			utgifterplats = append(utgifterplats, kategoriType{place, sum, summ})
		}

		sumutplats = sumutplats.Add(belopp)
	}

	var date []byte // size 10
	_ = db.QueryRow("SELECT MIN(Datum) FROM Transaktioner").Scan(&date)
	firstYear, _ := strconv.Atoi(toUtf8(date)[0:4])
	_ = db.QueryRow("SELECT MAX(Datum) FROM Transaktioner").Scan(&date)
	lastYear, _ := strconv.Atoi(toUtf8(date)[0:4])

	var years []string
	for i := firstYear; i <= lastYear; i++ {
		years = append(years, strconv.Itoa(i))
	}

	log.Println("Func hanteraYResult year:", strconv.Itoa(selectYear))
	log.Println("Func hanteraYResult sumin:", sumin.String())
	log.Println("Func hanteraYResult sumut:", sumut.String())
	log.Println("Func hanteraYResult sumutplats:", sumutplats.String())

	tmpl1 := template.New("wHHEK Årsresultat")
	tmpl1, _ = tmpl1.Parse(htmlresultat)
	data := ResultatData{
		SelectedYear:  strconv.Itoa(selectYear),
		Inkomster:     inkomster,
		SumIn:         Dec2Str(sumin),
		Utgifter:      utgifter,
		SumUt:         Dec2Str(sumut),
		UtgifterPlats: utgifterplats,
		SumUtPlats:    Dec2Str(sumutplats),
		Years:         years,
	}
	_ = tmpl1.Execute(w, data)
}

func sumKatYear(kat string, selectYear int, intyp bool) decimal.Decimal {
	result := decimal.NewFromInt32(0)

	year := decimal.NewFromInt(int64(selectYear)).String()
	startstring := year + "-01-01"
	endstring := year + "-12-31"

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
		_ = res.Scan(&amount)

		decamount, err := decimal.NewFromString(toUtf8(amount))
		if err != nil {
			log.Fatal(err)
		}

		result = result.Add(decamount)
	}
	res.Close()
	return result
}

func sumPlaceYear(kat string, selectYear int) decimal.Decimal {
	result := decimal.NewFromInt32(0)

	year := decimal.NewFromInt(int64(selectYear)).String()
	startstring := year + "-01-01"
	endstring := year + "-12-31"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var err error
	var res *sql.Rows

	res, err = db.QueryContext(ctx,
		`SELECT Belopp from transaktioner
  where TillKonto = ? and Datum >= ? and Datum <= ?`, kat, startstring, endstring)
	if err != nil {
		log.Fatal(err)
	}

	var amount []byte // BCD / Decimal Precision 19

	for res.Next() {
		_ = res.Scan(&amount)

		decamount, err := decimal.NewFromString(toUtf8(amount))
		if err != nil {
			log.Fatal(err)
		}

		result = result.Add(decamount)
	}
	res.Close()
	return result
}
