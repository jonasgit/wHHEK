//-*- coding: utf-8 -*-

package main

import (
	"bufio"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"mime/multipart"
	"strings"
	"time"
	
	"golang.org/x/text/encoding/charmap"
	"github.com/shopspring/decimal"  // MIT License
)

type matchning struct {
	dblopnr int
	utdragid int
 	klassning int  // 0 = ingen match, 1=perfekt match, 2=ungefärlig match
}

func readCsvFile(f multipart.File) [][]string {
	var res [][]string
	r := charmap.ISO8859_1.NewDecoder().Reader(f)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		textline := scanner.Text()
		rcsv := csv.NewReader(strings.NewReader(textline))
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

func finddaterange(records [][]string) (string, string) {
	var firstdate string = "2999-12-31"
	var lastdate string = "1900-01-01"
	for radnr, rad := range records {
		if radnr == 0 {
			// ignore header
		} else {
			date := rad[0]
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

func lopnrInAvst(lopnr int, avst []matchning) (lopnrInAvst bool, klassning int) {
	for _, rad := range avst {
		if rad.dblopnr == lopnr {
			return true, rad.klassning
		}
	}
	return false, 0
}

func DateWithinRange(date time.Time, dateRangeBase time.Time, rangeval int) (bool, error) {
	lowDate := dateRangeBase.AddDate(0, 0, -rangeval)  // minus range days
	hiDate := dateRangeBase.AddDate(0, 0, rangeval) // plus range days
	
	if date.Before(hiDate) && date.After(lowDate) {
		return true, nil
	} else {
		return false, nil
	}
	
	// default
	return false, nil
}

func lopnrNotUsed(lopnr int , result []matchning) bool {
	for _, rad := range result {
		if rad.dblopnr == lopnr {
			return false
		}
	}
	return true
}

func amountEquals(dbrad transaction, amount decimal.Decimal, kontonamn string) bool {
	if dbrad.tType == "Inköp" {
		return dbrad.amount.Equals(amount.Neg())
	}
	if dbrad.tType == "Insättning" {
		return dbrad.amount.Equals(amount)
	}
	if (dbrad.tType == "Överföring") && (dbrad.fromAcc == kontonamn) {
		return dbrad.amount.Equals(amount.Neg())
	}
	if (dbrad.tType == "Överföring") && (dbrad.toAcc == kontonamn) {
		return dbrad.amount.Equals(amount)
	}
	return dbrad.amount.Equals(amount)
}

func matchaUtdrag(records [][]string, dbtrans []transaction, kontonamn string) (result []matchning) {
	// Pass 1: Exakt matchning
	var firstline bool = true
	for radnr, rad := range records {
		if firstline {
			firstline = false
		} else {
			amountstrs := strings.Split(rad[4], " ")
			amountstr := strings.Replace(amountstrs[0], ",", ".", 1)
			rad_amount, err := decimal.NewFromString(amountstr)
			if err != nil {
				log.Fatal("matcha:", err)
			}
			raddate, err := time.Parse("2006-01-02", rad[0])
			
			for _, dbrad := range dbtrans {
				if amountEquals(dbrad, rad_amount, kontonamn) {
					//if dbrad.amount.Equals(rad_amount.Abs()) {
					if raddate == dbrad.date {
						if lopnrNotUsed(dbrad.lopnr, result) {
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
	firstline = true
	for radnr, rad := range records {
		if firstline {
			firstline = false
		} else {
			amountstrs := strings.Split(rad[4], " ")
			amountstr := strings.Replace(amountstrs[0], ",", ".", 1)
			rad_amount, err := decimal.NewFromString(amountstr)
			if err != nil {
				log.Fatal("matcha fuzzy:", err)
			}
			raddate, err := time.Parse("2006-01-02", rad[0])
			
			for _, dbrad := range dbtrans {
				//if dbrad.amount.Equals(rad_amount.Abs()) {
				if amountEquals(dbrad, rad_amount, kontonamn) {
					inRange, err := DateWithinRange(raddate, dbrad.date, 10)
					if err != nil {
						log.Fatal(err)
					}

					if inRange {
						if lopnrNotUsed(dbrad.lopnr, result) {
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

	records := readCsvFile(filen)
	firstdatestr, lastdatestr := finddaterange(records)
	firstdate,err := time.Parse("2006-01-02", firstdatestr)
	if err != nil {
		log.Fatal(err)
	}
	lastdate,err := time.Parse("2006-01-02", lastdatestr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(w, "Datum från %s till %s<p>", firstdate.Format("2006-01-02"), lastdate.Format("2006-01-02"))
	// expand date range for use in database
	firstdate = firstdate.AddDate(0, 0, -10)
	lastdate = lastdate.AddDate(0, 0, +10)
	
	
	dbtrans := getTransactionsInDateRange(db, kontonamn, firstdate.Format("2006-01-02"), lastdate.Format("2006-01-02"))

	avst := matchaUtdrag(records, dbtrans, kontonamn)
	
	fmt.Fprintf(w, "<table style=\"width:100%%\">")
	fmt.Fprintf(w, "<tr>Transaktioner från bankens fil</tr>")
	for radnr, rad := range records {
		fmt.Fprintf(w, "<tr>")
		if radnr == 0 { 
			fmt.Fprintf(w, "<th>Radnr</th>")
			fmt.Fprintf(w, "<th>Matchande löpnr</th>")
			for _, data := range rad {
				fmt.Fprintf(w, "<th>%s</th>", data)
			}
		} else {
			fmt.Fprintf(w, "<td>%d</td>", radnr)
			lopnr, radnrInAvst, klassning := radnrInAvst(radnr, avst)
			fmt.Fprintf(w, "<td>%d</td>", lopnr)
			for _, data := range rad {
				if radnrInAvst {
					if klassning == 1 {
						fmt.Fprintf(w, "<td bgcolor=\"green\">%s</td>",data)
					} else if klassning == 2 {
						fmt.Fprintf(w, "<td bgcolor=\"orange\">%s</td>",data)
					} else {
						fmt.Fprintf(w, "<td>%s</td>",data)
					}
				} else {
					fmt.Fprintf(w, "<td>%s</td>",data)
				}
			}
		}
		fmt.Fprintf(w, "</tr>")
	}
	fmt.Fprintf(w, "</table>")
	fmt.Fprintf(w, "<p>")
	
	fmt.Fprintf(w, "<table style=\"width:100%%\">")
	fmt.Fprintf(w, "<tr>Transaktioner från databasen</tr>")
	for _, rad := range dbtrans {
		fmt.Fprintf(w, "<tr>")
		
		lopnrInAvst, klassning := lopnrInAvst(rad.lopnr, avst)
		if lopnrInAvst {
			if klassning == 1 {
				fmt.Fprintf(w, "<td bgcolor=\"green\">%d</td>", rad.lopnr)
			} else if klassning == 2 {
				fmt.Fprintf(w, "<td bgcolor=\"orange\">%d</td>", rad.lopnr)
			} else {
				fmt.Fprintf(w, "<td>%d</td>", rad.lopnr)
			}
		} else {
			fmt.Fprintf(w, "<td>%d</td>", rad.lopnr)
		}
		fmt.Fprintf(w, "<td>%s</td>", rad.fromAcc)
		fmt.Fprintf(w, "<td>%s</td>", rad.toAcc)
		fmt.Fprintf(w, "<td>%s</td>", rad.tType)
		fmt.Fprintf(w, "<td>%s</td>", rad.what)
		fmt.Fprintf(w, "<td>%s</td>", rad.date.Format("2006-01-02"))
		fmt.Fprintf(w, "<td>%s</td>", rad.who)
		fmt.Fprintf(w, "<td>%s</td>", rad.amount)
		fmt.Fprintf(w, "<td>%s</td>", rad.comment)
		fmt.Fprintf(w, "<td>%t</td>", rad.fixed)

		fmt.Fprintf(w, "</tr>")
	}
	fmt.Fprintf(w, "</table>")

	fmt.Fprintf(w, "<p>Grön bakgrund betyder att raden/transaktionen matchar väl. Orange betyder att de matchar mindre bra men kan stämma. Övriga rader matchar inte alls.\n")
}

func printAccCompFooter(w http.ResponseWriter, db *sql.DB) {
	fmt.Fprintf(w, "<p><a href=\"summary\">Översikt</a>\n")
	fmt.Fprintf(w, "</body>\n")
	fmt.Fprintf(w, "</html>\n")
}

func compareaccount(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "<html>\n")
	fmt.Fprintf(w, "<head>\n")
	fmt.Fprintf(w, "<style>\n")
	fmt.Fprintf(w, "table,th,td { border: 1px solid black }\n")
	fmt.Fprintf(w, "</style>\n")
	fmt.Fprintf(w, "</head>\n")
	fmt.Fprintf(w, "<body>\n")

	fmt.Fprintf(w, "<h1>%s</h1>\n", currentDatabase)
	fmt.Fprintf(w, "<h2>Avstämning konto</h2>\n")

	req.ParseMultipartForm(32 << 20)
	file, fileMetaData, err := req.FormFile("uploadfile")
	if file != nil {
		defer file.Close()
	}
	
	var kontonamn string = ""
	var filtyp string = ""
	
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
		
		fmt.Fprintf(w, "<form enctype=\"multipart/form-data\" method=\"POST\" action=\"/acccmp\">\n")
		
		fmt.Fprintf(w, "  <label for=\"kontonamn\">Konto:</label>")
		fmt.Fprintf(w, "  <select name=\"kontonamn\" id=\"kontonamn\">")
		for _, s := range kontonamnlista {
			fmt.Fprintf(w, "    <option value=\"%s\">%s</option>", s, s)
		}
		fmt.Fprintf(w, "  </select><br>\n")

		fmt.Fprintf(w, "  <label for=\"filtyp\">Filtyp:</label>")
		fmt.Fprintf(w, "  <select name=\"filtyp\" id=\"filtyp\">")
		fmt.Fprintf(w, "    <option value=\"%s\">%s</option>", "komplettcsv", "KomplettBank CSV")
		fmt.Fprintf(w, "  </select><br>\n")

		fmt.Fprintf(w, "<input type=\"file\" name=\"uploadfile\" />")
		fmt.Fprintf(w, "  </select><br>\n")

		fmt.Fprintf(w, "<input type=\"submit\" value=\"Submit\"></form>\n")
		
	}
	printAccCompFooter(w, db)
}
