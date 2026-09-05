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
	"strings"
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

type Y3MonthRow struct {
	MonthName string
	Income    string
	Expense   string
	Sum       string
}

type Y3ResultData struct {
	Months      []Y3MonthRow
	ZeroLine    string
	MonthValues []MonthValueType
	GridLines   []GridLineType
	AxisLabels  []AxisLabelType
	MinSaldo    string
	MaxSaldo    string
	ChartTitle  string
	AverageLine string
}

func decimalToPlotPx(val, rangeVal decimal.Decimal, plotHeight int) int {
	if rangeVal.IsZero() {
		return 0
	}
	px := val.Mul(decimal.NewFromInt(int64(plotHeight))).Div(rangeVal)
	return int(px.Round(0).IntPart())
}

func y3ChartFromSums(sums []decimal.Decimal, monthTimes []time.Time) (int, []MonthValueType, []GridLineType, []AxisLabelType, decimal.Decimal, decimal.Decimal, string) {
	minSaldo := sums[0]
	maxSaldo := sums[0]
	for _, s := range sums {
		if s.GreaterThan(maxSaldo) {
			maxSaldo = s
		}
		if s.LessThan(minSaldo) {
			minSaldo = s
		}
	}
	if decimal.Zero.LessThan(minSaldo) {
		minSaldo = decimal.Zero
	}
	if decimal.Zero.GreaterThan(maxSaldo) {
		maxSaldo = decimal.Zero
	}
	minSaldo = minSaldo.Floor()
	maxSaldo = maxSaldo.Ceil()
	rangeVal := maxSaldo.Sub(minSaldo)
	if rangeVal.IsZero() {
		maxSaldo = decimal.NewFromInt(1)
		rangeVal = maxSaldo.Sub(minSaldo)
	}

	const colWidth = 20
	const chartHeight = 500
	const chartWidth = 900
	const marginLeft = 60
	const marginRight = 100
	const marginTop = 40
	const marginBottom = 50
	const plotWidth = chartWidth - marginLeft - marginRight
	const plotHeight = chartHeight - marginTop - marginBottom
	zeroDecimal := decimal.Zero

	var zeroLine int
	if maxSaldo.GreaterThan(zeroDecimal) && minSaldo.LessThan(zeroDecimal) {
		zeroLine = marginTop + decimalToPlotPx(maxSaldo.Sub(zeroDecimal), rangeVal, plotHeight)
	} else if minSaldo.GreaterThanOrEqual(zeroDecimal) {
		zeroLine = marginTop + plotHeight
	} else {
		zeroLine = marginTop
	}

	var monthValues []MonthValueType
	for i, currSum := range sums {
		var barHeight int
		var barY int
		if currSum.GreaterThanOrEqual(zeroDecimal) {
			barHeight = decimalToPlotPx(currSum.Sub(zeroDecimal), rangeVal, plotHeight)
			barY = zeroLine - barHeight
		} else {
			barHeight = decimalToPlotPx(zeroDecimal.Sub(currSum), rangeVal, plotHeight)
			barY = zeroLine
		}
		isPositive := "false"
		if currSum.GreaterThanOrEqual(zeroDecimal) {
			isPositive = "true"
		}
		monthValues = append(monthValues, MonthValueType{
			X:          strconv.Itoa(marginLeft + i*colWidth),
			Y:          strconv.Itoa(barY),
			Width:      strconv.Itoa(colWidth - 1),
			Height:     strconv.Itoa(barHeight),
			Day:        monthTimes[i].Format("2006-01"),
			Saldo:      Dec2Str(currSum),
			IsPositive: isPositive,
		})
	}

	dec3 := decimal.NewFromInt(3)
	var avgPoints []string
	for i := 1; i < (len(sums) - 1); i++ {
		avg := sums[i-1].Add(sums[i]).Add(sums[i+1]).Div(dec3)
		x := marginLeft + i*colWidth + colWidth/2
		var y int
		if avg.GreaterThanOrEqual(zeroDecimal) {
			y = zeroLine - decimalToPlotPx(avg.Sub(zeroDecimal), rangeVal, plotHeight)
		} else {
			y = zeroLine + decimalToPlotPx(zeroDecimal.Sub(avg), rangeVal, plotHeight)
		}
		avgPoints = append(avgPoints, fmt.Sprintf("%d,%d", x, y))
	}
	averageLine := strings.Join(avgPoints, " ")

	var gridLines []GridLineType
	numGridLines := 5
	for i := 0; i <= numGridLines; i++ {
		gridY := marginTop + (i * plotHeight / numGridLines)
		diff := gridY - zeroLine
		if diff < 0 {
			diff = -diff
		}
		gridLines = append(gridLines, GridLineType{
			X1:     strconv.Itoa(marginLeft),
			Y1:     strconv.Itoa(gridY),
			X2:     strconv.Itoa(marginLeft + plotWidth),
			Y2:     strconv.Itoa(gridY),
			IsZero: diff <= 2,
		})
	}
	zeroLineInGrid := false
	for _, gl := range gridLines {
		if gl.IsZero {
			zeroLineInGrid = true
			break
		}
	}
	if !zeroLineInGrid && zeroLine >= marginTop && zeroLine <= marginTop+plotHeight {
		gridLines = append(gridLines, GridLineType{
			X1:     strconv.Itoa(marginLeft),
			Y1:     strconv.Itoa(zeroLine),
			X2:     strconv.Itoa(marginLeft + plotWidth),
			Y2:     strconv.Itoa(zeroLine),
			IsZero: true,
		})
	}
	for i := 0; i < 36; i += 3 {
		gridX := marginLeft + i*colWidth
		gridLines = append(gridLines, GridLineType{
			X1:     strconv.Itoa(gridX),
			Y1:     strconv.Itoa(marginTop),
			X2:     strconv.Itoa(gridX),
			Y2:     strconv.Itoa(marginTop + plotHeight),
			IsZero: false,
		})
	}

	shortMonths := []string{"", "Jan", "Feb", "Mar", "Apr", "Maj", "Jun",
		"Jul", "Aug", "Sep", "Okt", "Nov", "Dec"}
	var axisLabels []AxisLabelType
	for i := 0; i < 36; i += 3 {
		m := monthTimes[i]
		axisLabels = append(axisLabels, AxisLabelType{
			X:       strconv.Itoa(marginLeft + i*colWidth + colWidth/2),
			Y:       strconv.Itoa(marginTop + plotHeight + 20),
			Text:    fmt.Sprintf("%s %02d", shortMonths[int(m.Month())], m.Year()%100),
			IsXAxis: true,
		})
	}
	numYLabels := 6
	numY := decimal.NewFromInt(int64(numYLabels))
	for i := 0; i <= numYLabels; i++ {
		labelY := marginTop + (i * plotHeight / numYLabels)
		ratio := decimal.NewFromInt(int64(i)).Div(numY)
		saldoValue := maxSaldo.Sub(rangeVal.Mul(ratio))
		axisLabels = append(axisLabels, AxisLabelType{
			X:       strconv.Itoa(marginLeft - 10),
			Y:       strconv.Itoa(labelY + 5),
			Text:    Dec2Str3Sig(saldoValue),
			IsXAxis: false,
		})
	}

	return zeroLine, monthValues, gridLines, axisLabels, minSaldo, maxSaldo, averageLine
}

func hanteraY3Result(w http.ResponseWriter, req *http.Request) {
	log.Println("Func hanteraY3Result")

	if db == nil {
		t := template.New("NoDatabase")
		t, _ = t.Parse(htmlnodatabase)
		err := t.Execute(w, nil)
		if err != nil {
			log.Println("While serving HTTP y3result no_database: ", err)
		}
		return
	}

	now := time.Now()
	currentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	firstMonth := currentMonth.AddDate(0, -35, 0)
	startstring := firstMonth.Format("2006-01-02")
	endstring := currentMonth.AddDate(0, 1, -1).Format("2006-01-02")

	incomeByMonth := make([]decimal.Decimal, 36)
	expenseByMonth := make([]decimal.Decimal, 36)
	zero := decimal.NewFromInt32(0)
	for i := 0; i < 36; i++ {
		incomeByMonth[i] = zero
		expenseByMonth[i] = zero
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	res, err := db.QueryContext(ctx,
		`SELECT Typ, Belopp, Datum from transaktioner
  where (Typ = ? or Typ = ? or Typ = ? or Typ = ?) and Datum >= ? and Datum <= ?`,
		"Insättning", "Fast Inkomst", "Inköp", "Fast Utgift", startstring, endstring)
	if err != nil {
		log.Fatal(err)
	}

	var typBuf []byte
	var amountBuf []byte
	var dateBuf []byte
	for res.Next() {
		_ = res.Scan(&typBuf, &amountBuf, &dateBuf)

		decamount, err := decimal.NewFromString(toUtf8(amountBuf))
		if err != nil {
			log.Fatal(err)
		}

		dateStr := toUtf8(dateBuf)
		if len(dateStr) < 7 {
			log.Println("hanterY3Result, saknar datum:" + dateStr)
			continue
		}
		year, errY := strconv.Atoi(dateStr[0:4])
		month, errM := strconv.Atoi(dateStr[5:7])
		if errY != nil || errM != nil {
			log.Println("hanterY3Result, trasigt datum:" + dateStr[0:4] + " eller " + dateStr[5:7])
			continue
		}
		idx := (year-firstMonth.Year())*12 + (month - int(firstMonth.Month()))
		if idx < 0 || idx > 35 {
			log.Println("hanterY3Result, utanför index:" + strconv.Itoa(idx))
			continue
		}

		typ := toUtf8(typBuf)
		if typ == "Insättning" || typ == "Fast Inkomst" {
			incomeByMonth[idx] = incomeByMonth[idx].Add(decamount)
		} else {
			expenseByMonth[idx] = expenseByMonth[idx].Add(decamount)
		}
	}
	res.Close()

	monthNames := []string{"", "Januari", "Februari", "Mars", "April", "Maj", "Juni",
		"Juli", "Augusti", "September", "Oktober", "November", "December"}

	var months []Y3MonthRow
	sums := make([]decimal.Decimal, 36)
	monthTimes := make([]time.Time, 36)
	for i := 0; i < 36; i++ {
		m := firstMonth.AddDate(0, i, 0)
		income := incomeByMonth[i]
		expense := expenseByMonth[i]
		sum := income.Add(expense.Neg())
		sums[i] = sum
		monthTimes[i] = m
		months = append(months, Y3MonthRow{
			MonthName: monthNames[int(m.Month())] + " " + strconv.Itoa(m.Year()),
			Income:    Dec2Str(income),
			Expense:   Dec2Str(expense),
			Sum:       Dec2Str(sum),
		})
	}

	zeroLine, monthValues, gridLines, axisLabels, minSaldo, maxSaldo, averageLine := y3ChartFromSums(sums, monthTimes)

	t, err := template.New("y3result.html").ParseFS(htmlTemplates, "html/y3result.html")
	if err != nil {
		log.Println("While parsing y3result.html: ", err)
		return
	}
	err = t.Execute(w, Y3ResultData{
		Months:      months,
		ZeroLine:    strconv.Itoa(zeroLine),
		MonthValues: monthValues,
		GridLines:   gridLines,
		AxisLabels:  axisLabels,
		MinSaldo:    Dec2Str(minSaldo),
		MaxSaldo:    Dec2Str(maxSaldo),
		ChartTitle:  "Summa per månad - 3 år",
		AverageLine: averageLine,
	})
	if err != nil {
		log.Println("While serving HTTP y3result: ", err)
	}
}
