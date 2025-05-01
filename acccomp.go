//-*- coding: utf-8 -*-

package main

import (
	"bufio"
	"database/sql"
	_ "embed"
	"encoding/csv"
	"html/template"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/extrame/xls"        // Apache-2.0 License
	"github.com/shopspring/decimal" // MIT License
	"github.com/xuri/excelize/v2"   // BSD-3-Clause License
	"golang.org/x/text/encoding/charmap"
)

type matchning struct {
	dblopnr   int
	utdragid  int
	klassning int // 0 = ingen match, 1=perfekt match, 2=ungefärlig match
}

func readXlsFile(f multipart.File) (res [][]string) {
	w, err := xls.OpenReader(f, "iso-8859-1")
	if err != nil {
		log.Println(err, getCurrentFuncName())
		return
	}
	res = w.ReadAllCells(100000)

	/*	for radnr, rad := range res {
			log.Print("XLS Radnr:", radnr)

			for colnr, data := range rad {
				log.Print("XLS Colnr:", colnr, " data: ", data)
			}
			log.Println("")
		}
	*/
	return res
}

func readXlsxFile(filen multipart.File) (res [][]string) {
	f, err := excelize.OpenReader(filen)
	if err != nil {
		log.Println(err)
		return
	}
	// Get all Sheets
	var sheetname string
	for index, name := range f.GetSheetMap() {
		log.Println("Found sheetname: ", index, name)
		sheetname = name
	}
	// Get all the rows in the Sheet1.
	rows, err := f.GetRows(sheetname)
	if err != nil {
		log.Println(err, getCurrentFuncName())
		return
	}
	res = append(res, rows...)
	return res
}

func readCsvFile(f multipart.File, filtyp string) [][]string {
	var res [][]string
	var r io.Reader

	// Select char encoding, UTF-8 or ISO8859
	if filtyp == "okq8csv" {
		r = f
	} else if filtyp == "lunarcsv" {
		r = f
	} else if filtyp == "morrowcsv" {
		r = f
	} else {
		r = charmap.ISO8859_1.NewDecoder().Reader(f)
	}
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		textline := scanner.Text()
		rcsv := csv.NewReader(strings.NewReader(textline))
		if filtyp == "okq8csv" {
			rcsv.Comma = ';'
			rcsv.LazyQuotes = true
		}
		for {
			record, err := rcsv.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Println(err, getCurrentFuncName())
				return res
			}
			res = append(res, record)
		}
	}

	return res
}

func finddaterange(records [][]string, filtyp string) (string, string) {
	var firstdate = "2999-12-31"
	var lastdate = "1100-01-01"

	headlines := bankheadlines(filtyp)

	for radnr, rad := range records {
		if radnr < headlines {
			// ignore header
		} else {
			date := findDateCol(rad, filtyp)
			if date < firstdate {
				firstdate = date
			}
			if date > lastdate {
				lastdate = date
			}
		}
	}
	return firstdate, lastdate
}

func radnrInAvst(radnr int, avst []matchning) (lopnr int, radnrInAvst bool, klassning int) {
	for _, rad := range avst {
		if rad.utdragid == radnr {
			return rad.dblopnr, true, rad.klassning
		}
	}
	return -1, false, 0
}

func lopnrInAvst(lopnr int, avst []matchning) (radnr int, lopnrInAvst bool, klassning int) {
	for _, rad := range avst {
		if rad.dblopnr == lopnr {
			return rad.utdragid, true, rad.klassning
		}
	}
	return -1, false, 0
}

func ExcelDay(serialNumber int) string {
	date, err := time.Parse("2006-01-02", "1900-01-01")
	if err != nil {
		log.Println(err, getCurrentFuncName())
		return ""
	}
	date = date.AddDate(0, 0, serialNumber-2)
	return date.Format("2006-01-02")
}

func DateWithinRange(date time.Time, dateRangeBase time.Time, rangeval int) (bool, error) {
	lowDate := dateRangeBase.AddDate(0, 0, -rangeval) // minus range days
	hiDate := dateRangeBase.AddDate(0, 0, rangeval)   // plus range days

	if date.Before(hiDate) && date.After(lowDate) {
		return true, nil
	}

	// default
	return false, nil
}

// Check if transaction is already matched, true if previously not found
func transNotMatched(radnr int, lopnr int, result []matchning) bool {
	for _, rad := range result {
		if (rad.dblopnr == lopnr) || (rad.utdragid == radnr) {
			return false
		}
	}
	return true
}

func amountEquals(dbrad transaction, amount decimal.Decimal, kontonamn string, filtyp string) bool {
	//if DEBUG_ON {
	//	log.Println("amountEquals", dbrad.lopnr, dbrad.amount, amount, kontonamn, filtyp, getCurrentFuncName())
	//}
	if dbrad.tType == "Inköp" {
		if filtyp == "eurocardxls" {
			return dbrad.amount.Equal(amount)
		} else if filtyp == "morrowcsv" {
			return dbrad.amount.Equal(amount)
		} else {
			return dbrad.amount.Equal(amount.Neg())
		}
	}
	if dbrad.tType == "Fast Utgift" {
		return dbrad.amount.Equal(amount.Neg())
	}
	if dbrad.tType == "Insättning" {
		return dbrad.amount.Equal(amount)
	}
	if (dbrad.tType == "Överföring") && (dbrad.fromAcc == kontonamn) {
		if filtyp == "eurocardxls" {
			return dbrad.amount.Equal(amount)
		} else if filtyp == "morrowcsv" {
			return dbrad.amount.Equal(amount)
		} else {
			return dbrad.amount.Equal(amount.Neg())
		}
	}
	if (dbrad.tType == "Överföring") && (dbrad.toAcc == kontonamn) {
		return dbrad.amount.Equal(amount)
	}
	if DEBUG_ON {
		log.Println("jämför", amount, dbrad.amount, getCurrentFuncName())
	}
	return dbrad.amount.Equal(amount)
}

func findAmount(rad []string, filtyp string) string {
	// Index för vilken kolumn som innehåller summan för transaktionen
	// första/vänstra kolumnen = 0
	switch filtyp {
	case "morrowcsv":
		return rad[6]
	case "swedbcsv":
		return rad[10]
	case "resursxlsx":
		return rad[4]
	case "revolutcsv":
		return rad[5]
	case "eurocardxls":
		return rad[6]
	case "okq8csv":
		return rad[3]
	case "lunarcsv":
		return rad[8]
	default:
		log.Println("Okänd filtyp:", filtyp, getCurrentFuncName())
		return ""
	}
}

// parse from string "29 feb 2006" to type time.Time
func parsedateSwe(datum string) time.Time {
	var date time.Time
	strs := strings.Split(datum, " ")
	day, err := strconv.Atoi(strs[0])
	if err != nil {
		log.Println("Fatal day", err, getCurrentFuncName())
		return time.Unix(0, 0)
	}
	year, err := strconv.Atoi(strs[2])
	if err != nil {
		log.Println("Fatal year", err, getCurrentFuncName())
		return time.Unix(0, 0)
	}
	var month time.Month = 0
	switch strs[1] {
	case "jan":
		month = 1
	case "feb":
		month = 2
	case "mar":
		month = 3
	case "apr":
		month = 4
	case "maj":
		month = 5
	case "jun":
		month = 6
	case "jul":
		month = 7
	case "aug":
		month = 8
	case "sep":
		month = 9
	case "okt":
		month = 10
	case "nov":
		month = 11
	case "dec":
		month = 12
	default:
		log.Println("Okänd månad:", strs[1], getCurrentFuncName())
		return time.Unix(0, 0)
	}

	location := time.FixedZone("CET", 0)
	date = time.Date(year, month, day, 12, 0, 0, 0, location)
	return date
}

func findDateCol(rad []string, filtyp string) string {
	// Index för vilken kolumn som innehåller datum för transaktionen
	// första/vänstra kolumnen = 0
	switch filtyp {
	case "morrowcsv":
		dateraw := rad[0]
		date, err := time.Parse("02.01.2006", dateraw)
		if err != nil {
			log.Println(err, getCurrentFuncName())
			return ""
		}
		return date.Format("2006-01-02")
	case "swedbcsv":
		return rad[6]
	case "resursxlsx":
		return rad[0]
	case "revolutcsv":
		strs := strings.Split(rad[2], " ")
		return strs[0]
	case "eurocardxls":
		days, err := strconv.Atoi(rad[0])
		if err != nil {
			log.Println(err, getCurrentFuncName())
			return ""
		}
		return ExcelDay(days)
	case "okq8csv":
		dateraw := rad[0]
		date := parsedateSwe(dateraw)
		return date.Format("2006-01-02")
	case "lunarcsv":
		return rad[0]
	default:
		log.Println("Okänd filtyp:", filtyp, getCurrentFuncName())
		return ""
	}
}

func bankheadlines(filtyp string) int {
	// Antal rubrikrader att hoppa över
	var headlines = 0
	switch filtyp {
	case "morrowcsv":
		fallthrough
	case "swedbcsv":
		headlines = 2
	case "resursxlsx":
		headlines = 1
	case "revolutcsv":
		headlines = 1
	case "eurocardxls":
		headlines = 5
	case "okq8csv":
		headlines = 1
	case "lunarcsv":
		headlines = 1
	default:
		log.Println("Okänd filtyp:", filtyp, getCurrentFuncName())
		return -1
	}
	return headlines
}

func matchaUtdrag(records [][]string, dbtrans []transaction, kontonamn string, filtyp string) (result []matchning) {
	headlines := bankheadlines(filtyp)

	// Pass 1: Exakt matchning
	for radnr, rad := range records {
		if radnr < headlines {
			// ignore header
		} else {
			amountcol := findAmount(rad, filtyp)
			amountstrs := strings.Split(amountcol, " ")
			amountstr := SanitizeAmount(amountstrs[0])
			if DEBUG_ON {
				//log.Println("jämför amount string ", amountstr, " ", amountcol, getCurrentFuncName())
			}
			radAmount, err := decimal.NewFromString(amountstr)
			if err != nil {
				log.Println("matcha error:", err, amountstr, getCurrentFuncName())
				return
			}
			datecol := findDateCol(rad, filtyp)
			raddate, err := time.Parse("2006-01-02", datecol)
			if err != nil {
				log.Println("date parse error:", err, datecol, getCurrentFuncName())
				return
			}
			if DEBUG_ON {
				//log.Println("jämför datum string ", datecol, " time.Time ", raddate)
			}
			for _, dbrad := range dbtrans {
				if amountEquals(dbrad, radAmount, kontonamn, filtyp) {
					if raddate == dbrad.date {
						if transNotMatched(radnr, dbrad.lopnr, result) {
							var match matchning
							match.dblopnr = dbrad.lopnr
							match.utdragid = radnr
							match.klassning = 1
							result = append(result, match)
							if DEBUG_ON {
								//log.Println("matchar allt ", dbrad.lopnr, " ", radnr)
							}
						} else {
							if DEBUG_ON {
								//log.Println("hittad tidigare")
							}
						}
					} else {
						if DEBUG_ON {
							//log.Println("datum matchar inte")
						}
					}
				} else {
					if DEBUG_ON {
						//log.Println("summan matchar inte ", dbrad.amount, dbrad.tType)
					}
				}
			}
		}
	}

	// Pass 2: Fuzzy matchning
	for radnr, rad := range records {
		if radnr < headlines {
			// ignore header
		} else {
			amountcol := findAmount(rad, filtyp)
			amountstrs := strings.Split(amountcol, " ")
			amountstr := SanitizeAmount(amountstrs[0])
			radAmount, err := decimal.NewFromString(amountstr)
			if err != nil {
				log.Println("matcha fuzzy:", err, getCurrentFuncName())
				return
			}
			datecol := findDateCol(rad, filtyp)
			raddate, err := time.Parse("2006-01-02", datecol)
			if err != nil {
				log.Println("date parse error:", err, datecol, getCurrentFuncName())
				return
			}

			for _, dbrad := range dbtrans {
				if amountEquals(dbrad, radAmount, kontonamn, filtyp) {
					inRange, err := DateWithinRange(raddate, dbrad.date, 10)
					if err != nil {
						log.Println(err, getCurrentFuncName())
						return
					}

					if inRange {
						if transNotMatched(radnr, dbrad.lopnr, result) {
							var match matchning
							match.dblopnr = dbrad.lopnr
							match.utdragid = radnr
							match.klassning = 2
							result = append(result, match)

						}
					}
				}
			}
		}
	}

	return result
}

//go:embed html/acccomp2.html
var htmlacccomp2 string

type ItemData struct {
	Data         string
	Matches      bool
	FuzzyMatches bool
}
type ItemDataArr []ItemData

type Acccomp2Data struct {
	DBName     string
	StartDate  string
	EndDate    string
	BankHeader []string
	BankRader  []ItemDataArr
	DBHeader   []string
	DBRader    []ItemDataArr
}

func printAvstämning(w http.ResponseWriter, db *sql.DB, kontonamn string, filtyp string, filen multipart.File) {
	log.Println("printavstämning kontonamn:", kontonamn, getCurrentFuncName())
	log.Println("printavstämning filtyp:", filtyp, getCurrentFuncName())

	var records [][]string
	switch filtyp {
	case "morrowcsv":
		records = readCsvFile(filen, filtyp)
	case "swedbcsv":
		records = readCsvFile(filen, filtyp)
	case "resursxlsx":
		records = readXlsxFile(filen)
	case "revolutcsv":
		records = readCsvFile(filen, filtyp)
	case "eurocardxls":
		records = readXlsFile(filen)
		// last line is summary, remove it for now
		if len(records) > 0 {
			records = records[:len(records)-1]
		}
	case "okq8csv":
		records = readCsvFile(filen, filtyp)
	case "lunarcsv":
		records = readCsvFile(filen, filtyp)
	default:
		log.Println("Okänd filtyp:", filtyp, getCurrentFuncName())
		return
	}
	log.Println("Read file. Antal rader:", len(records), getCurrentFuncName())
	firstdatestr, lastdatestr := finddaterange(records, filtyp)

	var firstdate time.Time
	var lastdate time.Time
	var err error

	firstdate, err = time.Parse("2006-01-02", firstdatestr)
	if err != nil {
		log.Println("first date", err, getCurrentFuncName())
		return
	}
	lastdate, err = time.Parse("2006-01-02", lastdatestr)
	if err != nil {
		log.Println("last date", err, getCurrentFuncName())
		return
	}

	bankfirstdate := firstdate
	banklastdate := lastdate
	// expand date range for use in database
	firstdate = firstdate.AddDate(0, 0, -10)
	lastdate = lastdate.AddDate(0, 0, +10)

	headlines := bankheadlines(filtyp)

	dbtrans := getTransactionsInDateRange(db, kontonamn, firstdate.Format("2006-01-02"), lastdate.Format("2006-01-02"))

	avst := matchaUtdrag(records, dbtrans, kontonamn, filtyp)

	var bankheader []string
	var bankrader []ItemDataArr

	for radnr, rad := range records {
		if radnr < headlines {
			bankheader = append(bankheader, rad...)
		} else {
			var bankrad []ItemData

			_, radnrInAvst, klassning := radnrInAvst(radnr, avst)
			for colnr, data := range rad {
				var item ItemData
				if filtyp == "eurocardxls" && colnr < 2 {
					days, err := strconv.Atoi(data)
					if err != nil {
						log.Println(err, getCurrentFuncName())
						return
					}
					data = ExcelDay(days)
				}
				if radnrInAvst {
					if klassning == 1 {
						item.Matches = true
					} else if klassning == 2 {
						item.FuzzyMatches = true
					}
				}
				item.Data = data
				bankrad = append(bankrad, item)
			}
			bankrader = append(bankrader, bankrad)
		}
	}

	dbheader := []string{"Löpnr", "Radnr", "Från konto", "Till konto", "Typ", "Vad", "Datum", "Vem", "Belopp", "Text", "Fast"}
	var dbrader []ItemDataArr
	for _, rad := range dbtrans {
		var dbrad []ItemData
		var item ItemData

		radnr, lopnrInAvst, klassning := lopnrInAvst(rad.lopnr, avst)
		if lopnrInAvst {
			if klassning == 1 {
				item.Matches = true
			} else if klassning == 2 {
				item.FuzzyMatches = true
			}
		}
		item.Data = strconv.Itoa(rad.lopnr)
		dbrad = append(dbrad, item)

		item.Data = strconv.Itoa(radnr)
		dbrad = append(dbrad, item)

		item.Data = rad.fromAcc
		dbrad = append(dbrad, item)

		item.Data = rad.toAcc
		dbrad = append(dbrad, item)

		item.Data = rad.tType
		dbrad = append(dbrad, item)

		item.Data = rad.what
		dbrad = append(dbrad, item)

		item.Data = rad.date.Format("2006-01-02")
		dbrad = append(dbrad, item)

		item.Data = rad.who
		dbrad = append(dbrad, item)

		item.Data = Dec2Str(rad.amount)
		dbrad = append(dbrad, item)

		item.Data = rad.comment
		dbrad = append(dbrad, item)

		item.Data = strconv.FormatBool(rad.fixed)
		dbrad = append(dbrad, item)

		dbrader = append(dbrader, dbrad)
	}

	t := template.New("Acccomp2")
	t, _ = t.Parse(htmlacccomp2)
	data := Acccomp2Data{
		DBName:     currentDatabase,
		StartDate:  bankfirstdate.Format("2006-01-02"),
		EndDate:    banklastdate.Format("2006-01-02"),
		BankHeader: bankheader,
		BankRader:  bankrader,
		DBHeader:   dbheader,
		DBRader:    dbrader,
	}
	err = t.Execute(w, data)
	if err != nil {
		log.Println("While serving HTTP acccomp2: ", err, getCurrentFuncName())
	}

}

//go:embed html/acccomp1.html
var htmlacccomp1 string

type OptionData struct {
	Label       string
	DisplayName string
}
type Acccomp1Data struct {
	DBName string
	Konton []OptionData
	Filtyp []OptionData
}

func compareaccount(w http.ResponseWriter, req *http.Request) {
	file, fileMetaData, err := req.FormFile("uploadfile")
	if file != nil {
		defer func(file multipart.File) {
			_ = file.Close()
		}(file)
	}

	var kontonamn = ""
	var filtyp = ""

	err = req.ParseMultipartForm(32 << 20) // buffer 32MB
	if err != nil {
		log.Println("compareacc parseerr:", err, getCurrentFuncName())
	} else {
		log.Println("Uploaded filename:", fileMetaData.Filename)

		if len(req.FormValue("kontonamn")) > 0 {
			kontonamn = req.FormValue("kontonamn")
			log.Println("Found kontonamn:", kontonamn)
		}
		if len(req.FormValue("filtyp")) > 0 {
			filtyp = req.FormValue("filtyp")
			log.Println("Found filtyp:", filtyp)
		}
	}

	if len(kontonamn) > 0 {
		log.Println("Valt kontonamn: ", kontonamn)

		printAvstämning(w, db, kontonamn, filtyp, file)
	} else {
		kontonamnlista := getAccNames()
		var konton []OptionData
		for _, s := range kontonamnlista {
			var konto OptionData
			konto.Label = s
			konto.DisplayName = s
			konton = append(konton, konto)
		}

		var filtyper []OptionData
		var filtyp OptionData

		filtyp = OptionData{Label: "morrowcsv", DisplayName: "MorrowBank CSV"}
		filtyper = append(filtyper, filtyp)

		filtyp = OptionData{Label: "swedbcsv", DisplayName: "Swedbank/Sparbankerna CSV"}
		filtyper = append(filtyper, filtyp)
		filtyp = OptionData{Label: "resursxlsx", DisplayName: "Resursbank/Fordkortet Xlsx"}
		filtyper = append(filtyper, filtyp)
		filtyp = OptionData{Label: "revolutcsv", DisplayName: "Revolut CSV(Excel)"}
		filtyper = append(filtyper, filtyp)
		filtyp = OptionData{Label: "eurocardxls", DisplayName: "Eurocard Xls"}
		filtyper = append(filtyper, filtyp)
		filtyp = OptionData{Label: "okq8csv", DisplayName: "OKQ8 CSV"}
		filtyper = append(filtyper, filtyp)
		filtyp = OptionData{Label: "lunarcsv", DisplayName: "Lunar CSV"}
		filtyper = append(filtyper, filtyp)

		t := template.New("Acccomp1")
		t, _ = t.Parse(htmlacccomp1)
		data := Acccomp1Data{
			DBName: currentDatabase,
			Konton: konton,
			Filtyp: filtyper,
		}
		err = t.Execute(w, data)
		if err != nil {
			log.Println("While serving HTTP acccomp1: ", err, getCurrentFuncName())
		}
	}
}
