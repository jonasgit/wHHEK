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

//go:embed html/fasta7ovf.html
var htmlfasta7ovf string

func showFastOvf(w http.ResponseWriter, db *sql.DB) {
	antal := GetCountFastOvf(db)
	if antal > 0 {
		var res *sql.Rows
		var err error
		res, err = db.Query("SELECT FrånKonto,TillKonto,Belopp,Datum,HurOfta,Vad,Vem,Löpnr,Kontrollnr,TillDatum,Rakning FROM Överföringar WHERE Vad = '---' and FrånKonto <> '---'")

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
		tmpl7, err = tmpl7.Parse(htmlfasta7ovf)
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

//go:embed html/fasta5ovf.html
var htmlfasta5ovf string

//go:embed html/fasta_footer.html
var htmlfastafooterovf string

func editfastovfHTML(w http.ResponseWriter, req *http.Request) {
	log.Println("editfastovfHTML start")
	currentTime := time.Now()
	currDate := currentTime.Format("2006-01-02")

	t := template.New("Fasta5")
	t, _ = t.Parse(htmlfasta5ovf)
	data := Fasta5Data{
		CurrDBName: currentDatabase,
		CurrDate:   currDate,
	}
	_ = t.Execute(w, data)

	if msg := r_e_fastovf(req, db); msg != "" {
		writeMessage(w, msg)
	}
	editfastovf(w, req, db)

	showFastOvf(w, db)

	t = template.New("FastaFooter")
	t, _ = t.Parse(htmlfastafooterovf)
	_ = t.Execute(w, nil)
	log.Println("editfastovfHTML slut")
}

func GetCountFastOvf(db *sql.DB) int {
	var cnt int
	_ = db.QueryRow(`select count(*) from Överföringar WHERE Vad = '---' and FrånKonto <> '---'`).Scan(&cnt)
	log.Println("GetCountFastOvf: ", cnt)
	return cnt
}

func skapaFastOverforing(db *sql.DB, franKonto string, tillKonto string, vem string, summa decimal.Decimal, datum string, hurofta string) error {
	if db == nil {
		log.Fatal("skapaFastOverforing anropad med db=nil")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := db.ExecContext(ctx,
		`INSERT INTO Överföringar (FrånKonto,TillKonto,Belopp,Datum,HurOfta, Vad, Vem, Kontrollnr, TillDatum, Rakning) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		franKonto, tillKonto, summa, datum, hurofta, "---", vem, 1, "---", false)

	if err != nil {
		log.Fatal(err)
	}

	return err
}

//go:embed html/fasta6ovf.html
var htmlfasta6ovf string

func editfastovf(w http.ResponseWriter, req *http.Request, db *sql.DB) {
	log.Println("editfastovf start")
	if !parseFormOrFail(w, req) {
		return
	}

	transtyp := req.FormValue("transtyp")
	log.Println("Val: ", transtyp, getCurrentFuncName())

	if transtyp == "FastTrans" {
		transidnum, err := fixedTransID(req, "transid")
		if err != nil {
			writeMessage(w, "Ogiltig transaktion: "+err.Error()+".")
			return
		}

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
			log.Println("Hittade inte fast transaktion ", transidnum, ": ", err)
			writeMessage(w, "Den fasta transaktionen hittades inte.")
			return
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
		t, _ = t.Parse(htmlfasta6ovf)
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
	log.Println("editfastovf slut")
}

// r_e_fastovf hanterar formulärets action och returnerar ett meddelande att visa
// (tomt om allt gick bra).
func r_e_fastovf(req *http.Request, db *sql.DB) string {
	log.Println("Func r_e_fastovf")

	if err := req.ParseForm(); err != nil {
		return "Kunde inte läsa formulärdata."
	}

	formaction := req.FormValue("action")
	lopnr, lopnrErr := fixedTransID(req, "lopnr")

	switch formaction {
	case "editform":
		log.Println("editform not implemented")
	case "update":
		if lopnrErr != nil {
			return "Ändringen sparades inte: " + lopnrErr.Error()
		}
		if err := updateFastOvfHTML(lopnr, req, db); err != nil {
			return "Ändringen sparades inte: " + err.Error()
		}
	default:
		log.Println("Okänd form action: ", formaction, getCurrentFuncName())
	}
	log.Println("Func r_e_fastovf finished")
	return ""
}

func updateFastOvfHTML(lopnr int, req *http.Request, db *sql.DB) error {
	log.Println("updateFastOvfHTML lopnr: ", lopnr)

	f, err := parseFixedForm(req)
	if err != nil {
		return err
	}

	err = updateFastOvfSQL(lopnr, db, f.fromAcc, f.toAcc, f.date, f.hurofta, f.tilldatum, f.who, f.amount, f.rakning)
	if err != nil {
		log.Println("Error updating fixed transfer: ", err)
		return err
	}
	return nil
}

func updateFastOvfSQL(lopnr int, db *sql.DB, fromAcc string, toAcc string, date string, hurofta string, tilldatum string, who string, amount string, rakning string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Only save what can be registered later (date ÅÅÅÅ-MM-DD and a known period).
	var err error
	date, hurofta, err = normalizeFixed(date, hurofta)
	if err != nil {
		return err
	}

	// empty string is not allowed in MDB
	if len(rakning) < 1 {
		rakning = " "
	}
	_, err = db.ExecContext(ctx,
		`UPDATE Överföringar SET FrånKonto = ?, TillKonto = ?, Datum = ?, HurOfta = ?, TillDatum = ?, Vad = ?, Vem = ?, Belopp = ?, Rakning = ? WHERE (Löpnr=?)`,
		fromAcc,
		toAcc,
		date,
		hurofta,
		tilldatum,
		"---",
		who,
		AmountStr2DBStr(amount),
		rakning,
		lopnr)

	if err != nil {
		log.Fatal(err)
	}
	return err
}
