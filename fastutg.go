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

//go:embed html/fasta7.html
var htmlfasta7 string

type Fasta7Data struct {
	Antal         string
	Transaktioner []transType
}

func showFastUtg(w http.ResponseWriter, db *sql.DB) {
	antal := GetCountFastUtg(db)
	if antal > 0 {
		var res *sql.Rows
		var err error
		res, err = db.Query("SELECT FrånKonto,TillKonto,Belopp,Datum,HurOfta,Vad,Vem,Löpnr,Kontrollnr,TillDatum,Rakning FROM Överföringar WHERE Vad <> '---' and FrånKonto <> '---'")

		if err != nil {
			log.Fatal(err)
		}

		var FrånKonto []byte // size 40
		var TillKonto []byte // size 40
		var Belopp []byte    // BCD / Decimal Precision 19
		var Datum []byte     // size 10
		var HurOfta []byte   // size 15
		var Vad []byte       // size 40
		var Vem []byte       // size 40
		var Löpnr []byte     // Autoinc Primary Key, index
		var Kontrollnr int   // Integer
		var TillDatum []byte // size 10
		var Rakning []byte   // size 1

		var transaktioner []transType
		var transaktion transType

		for res.Next() {
			_ = res.Scan(&FrånKonto, &TillKonto, &Belopp, &Datum, &HurOfta, &Vad, &Vem, &Löpnr, &Kontrollnr, &TillDatum, &Rakning)

			transaktion.Löpnr = toUtf8(Löpnr)
			transaktion.FranKonto = toUtf8(FrånKonto)
			transaktion.TillKonto = toUtf8(TillKonto)
			transaktion.Belopp = toUtf8(Belopp)
			transaktion.Datum = toUtf8(Datum)
			transaktion.HurOfta = toUtf8(HurOfta)
			transaktion.Vad = toUtf8(Vad)
			transaktion.Vem = toUtf8(Vem)
			transaktion.Kontrollnr = strconv.Itoa(Kontrollnr)
			transaktion.TillDatum = toUtf8(TillDatum)
			transaktion.Rakning = toUtf8(Rakning)
			transaktioner = append(transaktioner, transaktion)
		}

		res.Close()

		tmpl7 := template.New("wHHEK Fasta")
		tmpl7, err = tmpl7.Parse(htmlfasta7)
		if err != nil {
			log.Fatal(err)
		}
		data := Fasta7Data{
			Antal:         strconv.Itoa(antal),
			Transaktioner: transaktioner,
		}
		_ = tmpl7.Execute(w, data)
	}
}

//go:embed html/fasta5.html
var htmlfasta5 string

type Fasta5Data struct {
	CurrDBName string
	CurrDate   string
}

func editfastutgHTML(w http.ResponseWriter, req *http.Request) {
	log.Println("editfastutgHTML start")
	currentTime := time.Now()
	currDate := currentTime.Format("2006-01-02")

	t := template.New("Fasta5")
	t, _ = t.Parse(htmlfasta5)
	data := Fasta5Data{
		CurrDBName: currentDatabase,
		CurrDate:   currDate,
	}
	_ = t.Execute(w, data)

	r_e_fastutg(req, db)
	editfastutg(w, req, db)

	showFastUtg(w, db)

	_, _ = fmt.Fprintf(w, "<a href=\"summary\">Översikt</a>\n")
	_, _ = fmt.Fprintf(w, "</body>\n")
	_, _ = fmt.Fprintf(w, "</html>\n")
	log.Println("editfastutgHTML slut")
}

func GetCountFastUtg(db *sql.DB) int {
	var cnt int
	// TODO
	_ = db.QueryRow(`select count(*) from Överföringar WHERE Vad <> '---' and FrånKonto <> '---'`).Scan(&cnt)
	log.Println("GetCountFastUtg: ", cnt)
	return cnt
}

func skapaFastUtgift(db *sql.DB, vad string, konto string, vem string, plats string, summa decimal.Decimal, datum string, hurofta string) error {
	if db == nil {
		log.Fatal("skapaFastUtgift anropad med db=nil")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := db.ExecContext(ctx,
		`INSERT INTO Överföringar (FrånKonto,TillKonto,Belopp,Datum,HurOfta, Vad, Vem, Kontrollnr, TillDatum, Rakning) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, konto, plats, summa, datum, hurofta, vad, vem, 1, "---", false)

	if err != nil {
		log.Fatal(err)
	}

	return err
}

//go:embed html/fasta6.html
var htmlfasta6 string

type Fasta6Data struct {
	Transaktion     transType
	KontonamnLista  []string
	PlatserLista    []string
	PersonerLista   []string
	VadInkomstLista []string
	VadUtgiftLista  []string
	Lopnr           int
}

func editfastutg(w http.ResponseWriter, req *http.Request, db *sql.DB) {
	log.Println("editfastutg start")
	err := req.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	transtyp := req.FormValue("transtyp")
	date := req.FormValue("date")
	who := req.FormValue("who")
	amount := req.FormValue("amount")
	text := req.FormValue("text")
	log.Println("Val: ", transtyp, getCurrentFuncName())
	log.Println("Val: ", date)
	log.Println("Val: ", who)
	log.Println("Val: ", amount)
	log.Println("Val: ", text)

	if transtyp == "FastTrans" {
		transid := req.FormValue("transid")
		transidnum, _ := strconv.Atoi(transid)

		var err error

		var FrånKonto []byte  // size 40
		var TillKonto []byte  // size 40
		var Belopp []byte     // BCD / Decimal Precision 19
		var Datum []byte      // size 10
		var HurOfta []byte    // size 15
		var Vad []byte        // size 40
		var Vem []byte        // size 40
		var Löpnr []byte      // Autoinc Primary Key, index
		var Kontrollnr []byte // Borde vara?Integer
		var TillDatum []byte  // size 10
		var Rakning []byte    // size 1

		var transaktion transType

		err = db.QueryRow("SELECT FrånKonto,TillKonto,Belopp,Datum,HurOfta,Vad,Vem,Löpnr,Kontrollnr,TillDatum,Rakning FROM Överföringar WHERE Löpnr = ?", transidnum).Scan(&FrånKonto, &TillKonto, &Belopp, &Datum, &HurOfta, &Vad, &Vem, &Löpnr, &Kontrollnr, &TillDatum, &Rakning)

		if err != nil {
			log.Fatal(err)
		}

		transaktion.Löpnr = toUtf8(Löpnr)
		transaktion.FranKonto = toUtf8(FrånKonto)
		transaktion.TillKonto = toUtf8(TillKonto)
		transaktion.Belopp = toUtf8(Belopp)
		transaktion.Datum = toUtf8(Datum)
		transaktion.HurOfta = toUtf8(HurOfta)
		transaktion.Vad = toUtf8(Vad)
		transaktion.Vem = toUtf8(Vem)
		transaktion.Kontrollnr = toUtf8(Kontrollnr)
		transaktion.TillDatum = toUtf8(TillDatum)
		transaktion.Rakning = toUtf8(Rakning)

		kontonamnlista := getAccNames()
		platserlista := getPlaceNames()
		personerlista := getPersonNames()
		vadinkomstlista := getTypeInNames()
		vadutgiftlista := getTypeOutNames()

		t := template.New("EditFixed6")
		t, _ = t.Parse(htmlfasta6)
		data := Fasta6Data{
			Transaktion:     transaktion,
			KontonamnLista:  kontonamnlista,
			PlatserLista:    platserlista,
			PersonerLista:   personerlista,
			VadInkomstLista: vadinkomstlista,
			VadUtgiftLista:  vadutgiftlista,
			Lopnr:           transidnum,
		}
		err = t.Execute(w, data)
		if err != nil {
			return
		}
	}
	log.Println("editfastutg slut")
}

func r_e_fastutg(req *http.Request, db *sql.DB) {
	log.Println("Func r_e_fastutg")

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
		log.Println("editform not implemented")
	case "update":
		updateFastUtgHTML(lopnr, req, db)
	default:
		log.Println("Okänd form action: ", formaction, getCurrentFuncName())
	}
	log.Println("Func r_e_fastutg finished")
}

func updateFastUtgHTML(lopnr int, req *http.Request, db *sql.DB) {
	log.Println("updateFastUtgHTML lopnr: ", lopnr)

	var fromAcc = ""
	if len(req.FormValue("fromAcc")) > 0 {
		fromAcc = req.FormValue("fromAcc")
	}
	var toAcc = ""
	if len(req.FormValue("toAcc")) > 0 {
		toAcc = req.FormValue("toAcc")
	}
	var date = ""
	if len(req.FormValue("date")) > 0 {
		date = req.FormValue("date")
	}
	var hurofta = ""
	if len(req.FormValue("hurofta")) > 0 {
		hurofta = req.FormValue("hurofta")
	}
	var tilldatum = ""
	if len(req.FormValue("tilldatum")) > 0 {
		tilldatum = req.FormValue("tilldatum")
	}
	var what = ""
	if len(req.FormValue("what")) > 0 {
		what = req.FormValue("what")
	}
	var who = ""
	if len(req.FormValue("who")) > 0 {
		who = req.FormValue("who")
	}
	var amount = ""
	if len(req.FormValue("amount")) > 0 {
		amount = SanitizeAmount(req.FormValue("amount"))
	}
	var rakning = ""
	if len(req.FormValue("rakning")) > 0 {
		rakning = req.FormValue("rakning")
	}

	err := updateFastUtgSQL(lopnr, db, fromAcc, toAcc, date, hurofta, tilldatum, what, who, amount, rakning)
	if err != nil {
		log.Println("Error updating fixed expense: ", err)
		return
	}
}

func updateFastUtgSQL(lopnr int, db *sql.DB, fromAcc string, toAcc string, date string, hurofta string, tilldatum string, what string, who string, amount string, rakning string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// empty string is not allowed in MDB
	if len(rakning) < 1 {
		rakning = " "
	}
	_, err := db.ExecContext(ctx,
		`UPDATE Överföringar SET FrånKonto = ?, TillKonto = ?, Datum = ?, HurOfta = ?, TillDatum = ?, Vad = ?, Vem = ?, Belopp = ?, Rakning = ? WHERE (Löpnr=?)`,
		fromAcc,
		toAcc,
		date,
		hurofta,
		tilldatum,
		what,
		who,
		AmountStr2DBStr(amount),
		rakning,
		lopnr)

	if err != nil {
		log.Fatal(err)
	}
	return err
}
