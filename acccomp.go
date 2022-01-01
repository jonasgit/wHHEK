//-*- coding: utf-8 -*-

package main

import (
	"bufio"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/extrame/xls"        // Apache-2.0 License
	"github.com/shopspring/decimal" // MIT License
	"github.com/xuri/excelize/v2"   //  BSD-3-Clause License
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
		log.Fatal(err)
	}
	res = w.ReadAllCells(100000)

	/*	for radnr, rad := range res {
			fmt.Print("XLS Radnr:", radnr)

			for colnr, data := range rad {
				fmt.Print("XLS Colnr:", colnr, " data: ", data)
			}
			fmt.Println("")
		}
	*/
	return res
}

func readXlsxFile(filen multipart.File) (res [][]string) {
	f, err := excelize.OpenReader(filen)
	if err != nil {
		fmt.Println(err)
		return
	}
	// Get all Sheets
	var sheetname string
	for index, name := range f.GetSheetMap() {
		fmt.Println("Found sheetname: ", index, name)
		sheetname = name
	}
	// Get all the rows in the Sheet1.
	rows, err := f.GetRows(sheetname)
	if err != nil {
		log.Fatal(err)
	}
	for _, row := range rows {
		var record []string
		for _, colCell := range row {
			record = append(record, colCell)
		}
		res = append(res, record)
	}
	return res
}

func readCsvFile(f multipart.File, filtyp string) [][]string {
	var res [][]string
	var r io.Reader
	if filtyp == "okq8csv" {
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
				log.Fatal(err)
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
		log.Fatal(err)
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

func transNotMatched(radnr int, lopnr int, result []matchning) bool {
	for _, rad := range result {
		if (rad.dblopnr == lopnr) || (rad.utdragid == radnr) {
			return false
		}
	}
	return true
}

func amountEquals(dbrad transaction, amount decimal.Decimal, kontonamn string, filtyp string) bool {
	if dbrad.tType == "Inköp" {
		if filtyp == "eurocardxls" {
			return dbrad.amount.Equals(amount)
		} else {
			return dbrad.amount.Equals(amount.Neg())
		}
	}
	if dbrad.tType == "Fast Utgift" {
		return dbrad.amount.Equals(amount.Neg())
	}
	if dbrad.tType == "Insättning" {
		return dbrad.amount.Equals(amount)
	}
	if (dbrad.tType == "Överföring") && (dbrad.fromAcc == kontonamn) {
		if filtyp == "eurocardxls" {
			return dbrad.amount.Equals(amount)
		} else {
			return dbrad.amount.Equals(amount.Neg())
		}
	}
	if (dbrad.tType == "Överföring") && (dbrad.toAcc == kontonamn) {
		return dbrad.amount.Equals(amount)
	}
	return dbrad.amount.Equals(amount)
}

func findAmount(rad []string, filtyp string) string {
	// Index för vilken kolumn som innehåller summan för transaktionen
	// första/vänstra kolumnen = 0
	switch filtyp {
	case "komplettcsv":
		return rad[4]
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
		return rad[2]
	default:
		log.Fatal("Okänd filtyp:", filtyp)
	}
	return "-1"
}

// parse from string "29 feb 2006" to type time.Time
func parsedateSwe(datum string) time.Time {
	var date time.Time
	strs := strings.Split(datum, " ")
	day, err := strconv.Atoi(strs[0])
	year, err := strconv.Atoi(strs[2])
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
		log.Fatal("Okänd månad:", strs[1])
	}

	location, err := time.LoadLocation("Europe/Stockholm")
	if err != nil {
		panic(err)
	}

	date = time.Date(year, month, day, 12, 0, 0, 0, location)
	return date
}

func findDateCol(rad []string, filtyp string) string {
	switch filtyp {
	case "komplettcsv":
		return rad[0]
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
			log.Fatal(err)
		}
		return ExcelDay(days)
	case "okq8csv":
		dateraw := rad[0]
		date := parsedateSwe(dateraw)
		return date.Format("2006-01-02")
	case "lunarcsv":
		return rad[0]
	default:
		log.Fatal("Okänd filtyp:", filtyp)
	}
	return "-1"
}

func bankheadlines(filtyp string) int {
	var headlines = 0
	switch filtyp {
	case "komplettcsv":
		headlines = 1
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
		log.Fatal("Okänd filtyp:", filtyp)
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
			amountstr := strings.Replace(amountstrs[0], ",", ".", 1)
			//fmt.Println("matchar amount ", amountstr, " ", amountcol)
			radAmount, err := decimal.NewFromString(amountstr)
			if err != nil {
				log.Fatal("matcha:", err)
			}
			datecol := findDateCol(rad, filtyp)
			raddate, err := time.Parse("2006-01-02", datecol)
			//fmt.Println("matchar datum ", datecol, " ", raddate)
			for _, dbrad := range dbtrans {
				if amountEquals(dbrad, radAmount, kontonamn, filtyp) {
					if raddate == dbrad.date {
						if transNotMatched(radnr, dbrad.lopnr, result) {
							var match matchning
							match.dblopnr = dbrad.lopnr
							match.utdragid = radnr
							match.klassning = 1
							result = append(result, match)
							//fmt.Println("matchar allt ", dbrad.lopnr, " ", radnr)

						}
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
			amountstr := strings.Replace(amountstrs[0], ",", ".", 1)
			radAmount, err := decimal.NewFromString(amountstr)
			if err != nil {
				log.Fatal("matcha fuzzy:", err)
			}
			datecol := findDateCol(rad, filtyp)
			raddate, err := time.Parse("2006-01-02", datecol)

			for _, dbrad := range dbtrans {
				if amountEquals(dbrad, radAmount, kontonamn, filtyp) {
					inRange, err := DateWithinRange(raddate, dbrad.date, 10)
					if err != nil {
						log.Fatal(err)
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

func printAvstämning(w http.ResponseWriter, db *sql.DB, kontonamn string, filtyp string, filen multipart.File) {
	fmt.Println("printavstämning kontonamn:", kontonamn)
	fmt.Println("printavstämning filtyp:", filtyp)

	var records [][]string
	switch filtyp {
	case "komplettcsv":
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
		log.Fatal("Okänd filtyp:", filtyp)
	}
	fmt.Println("Read file. Antal rader:", len(records))
	firstdatestr, lastdatestr := finddaterange(records, filtyp)
	firstdate, err := time.Parse("2006-01-02", firstdatestr)
	if err != nil {
		log.Fatal(err)
	}
	lastdate, err := time.Parse("2006-01-02", lastdatestr)
	if err != nil {
		log.Fatal(err)
	}
	_, _ = fmt.Fprintf(w, "Datum från %s till %s<p>", firstdate.Format("2006-01-02"), lastdate.Format("2006-01-02"))
	fmt.Println("Datum från ", firstdate.Format("2006-01-02"), " till ", lastdate.Format("2006-01-02"))
	// expand date range for use in database
	firstdate = firstdate.AddDate(0, 0, -10)
	lastdate = lastdate.AddDate(0, 0, +10)

	headlines := bankheadlines(filtyp)

	dbtrans := getTransactionsInDateRange(db, kontonamn, firstdate.Format("2006-01-02"), lastdate.Format("2006-01-02"))

	avst := matchaUtdrag(records, dbtrans, kontonamn, filtyp)

	_, _ = fmt.Fprintf(w, "<table style=\"width:100%%\">")
	_, _ = fmt.Fprintf(w, "<tr>Transaktioner från bankens fil</tr>")
	for radnr, rad := range records {
		_, _ = fmt.Fprintf(w, "<tr>")
		if radnr < headlines {
			_, _ = fmt.Fprintf(w, "<th>Radnr</th>")
			_, _ = fmt.Fprintf(w, "<th>Matchande löpnr</th>")
			for _, data := range rad {
				_, _ = fmt.Fprintf(w, "<th>%s</th>", data)
			}
		} else {
			_, _ = fmt.Fprintf(w, "<td>%d</td>", radnr)
			lopnr, radnrInAvst, klassning := radnrInAvst(radnr, avst)
			_, _ = fmt.Fprintf(w, "<td>%d</td>", lopnr)
			for colnr, data := range rad {
				if filtyp == "eurocardxls" && colnr < 2 {
					days, err := strconv.Atoi(data)
					if err != nil {
						log.Fatal(err)
					}
					data = ExcelDay(days)
				}
				if radnrInAvst {
					if klassning == 1 {
						_, _ = fmt.Fprintf(w, "<td bgcolor=\"green\">%s</td>", data)
					} else if klassning == 2 {
						_, _ = fmt.Fprintf(w, "<td bgcolor=\"orange\">%s</td>", data)
					} else {
						_, _ = fmt.Fprintf(w, "<td>%s</td>", data)
					}
				} else {
					_, _ = fmt.Fprintf(w, "<td>%s</td>", data)
				}
			}
		}
		_, _ = fmt.Fprintf(w, "</tr>")
	}
	_, _ = fmt.Fprintf(w, "</table>")
	_, _ = fmt.Fprintf(w, "<p>")

	_, _ = fmt.Fprintf(w, "<table style=\"width:100%%\">")
	_, _ = fmt.Fprintf(w, "<tr>Transaktioner från databasen</tr>")
	for _, rad := range dbtrans {
		_, _ = fmt.Fprintf(w, "<tr>")

		radnr, lopnrInAvst, klassning := lopnrInAvst(rad.lopnr, avst)
		if lopnrInAvst {
			if klassning == 1 {
				_, _ = fmt.Fprintf(w, "<td bgcolor=\"green\">%d</td>", rad.lopnr)
			} else if klassning == 2 {
				_, _ = fmt.Fprintf(w, "<td bgcolor=\"orange\">%d</td>", rad.lopnr)
			} else {
				_, _ = fmt.Fprintf(w, "<td>%d</td>", rad.lopnr)
			}
		} else {
			_, _ = fmt.Fprintf(w, "<td>%d</td>", rad.lopnr)
		}
		_, _ = fmt.Fprintf(w, "<td>%d</td>", radnr)

		_, _ = fmt.Fprintf(w, "<td>%s</td>", rad.fromAcc)
		_, _ = fmt.Fprintf(w, "<td>%s</td>", rad.toAcc)
		_, _ = fmt.Fprintf(w, "<td>%s</td>", rad.tType)
		_, _ = fmt.Fprintf(w, "<td>%s</td>", rad.what)
		_, _ = fmt.Fprintf(w, "<td>%s</td>", rad.date.Format("2006-01-02"))
		_, _ = fmt.Fprintf(w, "<td>%s</td>", rad.who)
		_, _ = fmt.Fprintf(w, "<td>%s</td>", rad.amount)
		_, _ = fmt.Fprintf(w, "<td>%s</td>", rad.comment)
		_, _ = fmt.Fprintf(w, "<td>%t</td>", rad.fixed)

		_, _ = fmt.Fprintf(w, "</tr>")
	}
	_, _ = fmt.Fprintf(w, "</table>")

	_, _ = fmt.Fprintf(w, "<p>Grön bakgrund betyder att raden/transaktionen matchar väl. Orange betyder att de matchar mindre bra men kan stämma. Övriga rader matchar inte alls.\n")
}

func printAccCompFooter(w http.ResponseWriter) {
	_, _ = fmt.Fprintf(w, "<p><a href=\"summary\">Översikt</a>\n")
	_, _ = fmt.Fprintf(w, "</body>\n")
	_, _ = fmt.Fprintf(w, "</html>\n")
}

func compareaccount(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	w.WriteHeader(200)

	_, _ = fmt.Fprintf(w, "<html>\n")
	_, _ = fmt.Fprintf(w, "<head>\n")
	_, _ = fmt.Fprintf(w, "<style>\n")
	_, _ = fmt.Fprintf(w, "table,th,td { border: 1px solid black }\n")
	_, _ = fmt.Fprintf(w, "</style>\n")
	_, _ = fmt.Fprintf(w, "</head>\n")
	_, _ = fmt.Fprintf(w, "<body>\n")

	_, _ = fmt.Fprintf(w, "<h1>%s</h1>\n", currentDatabase)
	_, _ = fmt.Fprintf(w, "<h2>Avstämning konto</h2>\n")

	_ = req.ParseMultipartForm(32 << 20)
	file, fileMetaData, err := req.FormFile("uploadfile")
	if file != nil {
		defer func(file multipart.File) {
			_ = file.Close()
		}(file)
	}

	var kontonamn = ""
	var filtyp = ""

	if err != nil {
		log.Println("compareacc parseerr:", err)
	} else {
		fmt.Println("Uploaded filename:", fileMetaData.Filename)

		if len(req.FormValue("kontonamn")) > 0 {
			kontonamn = req.FormValue("kontonamn")
			fmt.Println("Found kontonamn:", kontonamn)
		}
		if len(req.FormValue("filtyp")) > 0 {
			filtyp = req.FormValue("filtyp")
			fmt.Println("Found filtyp:", filtyp)
		}
	}

	if len(kontonamn) > 0 {
		fmt.Println("Valt kontonamn: ", kontonamn)

		printAvstämning(w, db, kontonamn, filtyp, file)
	} else {
		kontonamnlista := getAccNames()

		_, _ = fmt.Fprintf(w, "<form enctype=\"multipart/form-data\" method=\"POST\" action=\"/acccmp\">\n")

		_, _ = fmt.Fprintf(w, "  <label for=\"kontonamn\">Konto:</label>")
		_, _ = fmt.Fprintf(w, "  <select name=\"kontonamn\" id=\"kontonamn\">")
		for _, s := range kontonamnlista {
			_, _ = fmt.Fprintf(w, "    <option value=\"%s\">%s</option>", s, s)
		}
		_, _ = fmt.Fprintf(w, "  </select><br>\n")

		_, _ = fmt.Fprintf(w, "  <label for=\"filtyp\">Filtyp:</label>")
		_, _ = fmt.Fprintf(w, "  <select name=\"filtyp\" id=\"filtyp\">")
		_, _ = fmt.Fprintf(w, "    <option value=\"%s\">%s</option>", "komplettcsv", "KomplettBank CSV")
		_, _ = fmt.Fprintf(w, "    <option value=\"%s\">%s</option>", "swedbcsv", "Swedbank/Sparbankerna CSV")
		_, _ = fmt.Fprintf(w, "    <option value=\"%s\">%s</option>", "resursxlsx", "Resursbank/Fordkortet Xlsx")
		_, _ = fmt.Fprintf(w, "    <option value=\"%s\">%s</option>", "revolutcsv", "Revolut CSV(Excel)")
		_, _ = fmt.Fprintf(w, "    <option value=\"%s\">%s</option>", "eurocardxls", "Eurocard Xls")
		_, _ = fmt.Fprintf(w, "    <option value=\"%s\">%s</option>", "okq8csv", "OKQ8 CSV")
		_, _ = fmt.Fprintf(w, "    <option value=\"%s\">%s</option>", "lunarcsv", "Lunar CSV")
		_, _ = fmt.Fprintf(w, "  </select><br>\n")

		_, _ = fmt.Fprintf(w, "<input type=\"file\" name=\"uploadfile\" />")
		_, _ = fmt.Fprintf(w, "  </select><br>\n")

		_, _ = fmt.Fprintf(w, "<input type=\"submit\" value=\"Submit\"></form>\n")

	}
	printAccCompFooter(w)
}
