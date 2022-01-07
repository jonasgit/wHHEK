//-*- coding: utf-8 -*-
package main

import (
	"database/sql"
	"strconv"
	"testing"

	"github.com/shopspring/decimal" // MIT License
)

func transaktionInit(t *testing.T, filnamn string) *sql.DB {
	// Förberedelser
	if JetDBSupport {
		var filename = "got" + filnamn + ".mdb"
		t.Log("Jet Supported.")

		SkapaTomMDB(t, filename)
		db = openJetDB(filename, false)
	} else {
		var filename = "got" + filnamn + ".db"
		t.Log("Jet NOT Supported.")
		SkapaTomDB(filename)
		db = openSqlite(filename)
	}

	if db == nil {
		t.Fatal("Ingen databas.")
	}
	return db
}

func TestTransaktionTomDB1(t *testing.T) {
	t.Log("TestTransaktionTomDB1")
	transaktionInit(t, "trt")

	// Denna testen
	antal := antalTransaktioner(db)

	if antal != 0 {
		t.Error("Antal transaktioner != (0).")
	} else {
		t.Log("Antal transaktioner ok (0).")
	}

	closeDB()
}

func TestTransaktionDB1(t *testing.T) {
	t.Log("TestTransaktionDB1")
	transaktionInit(t, "tr1")

	// Denna testen
	// Kontrollera att vi utgår från startsaldo 0.00
	// Gör insättning 0,10kr och kontrollerar att resultatet blir 0,10kr
	antal := antalTransaktioner(db)

	if antal != 0 {
		t.Error("Antal transaktioner (0) != " + strconv.Itoa(antal))
	} else {
		t.Log("Antal transaktioner ok (0).")
	}

	saldoExpected := decimal.NewFromInt(0)
	konto := hämtaKonto(db, 1)

	if !konto.Saldo.Equal(saldoExpected) {
		t.Error("Konto saldo '" + saldoExpected.String() + "' != '" + konto.Saldo.String() + "'.")
	} else {
		t.Log("Test saldo ok. 0.00")
	}

	summa, err := decimal.NewFromString("0.1")
	if err != nil {
		t.Error(err)
	}
	addTransaktionInsättning("Plånboken", "2021-07-27", "Övriga inkomster", "Gemensamt", summa, "Tom € Räksmörgås")

	antal = antalTransaktioner(db)

	if antal != 1 {
		t.Error("Antal transaktioner (1) != " + strconv.Itoa(antal))
	} else {
		t.Log("Antal transaktioner ok (1).")
	}

	saldoExpected, err = decimal.NewFromString("0.1")
	konto = hämtaKonto(db, 1)

	if !konto.Saldo.Equal(saldoExpected) {
		t.Error("Konto saldo '" + saldoExpected.String() + "' != '" + konto.Saldo.String() + "'.")
	} else {
		t.Log("Test saldo ok. 0.1")
	}

	saldo := saldoKonto(db, "Plånboken", "2020-07-27")
	if !saldo.Equal(decimal.NewFromInt(0)) {
		t.Error("Konto saldo innan '" + "0,00" + "' != '" + saldo.String() + "'.")
	} else {
		t.Log("Test innan saldo ok.")
	}

	saldo = saldoKonto(db, "Plånboken", "2021-07-28")
	if !saldo.Equal(saldoExpected) {
		t.Error("Konto saldo efter '" + saldo.String() + "' != '" + saldoExpected.String() + "'.")
	} else {
		t.Log("Test efter saldo ok. " + saldoExpected.String())
	}

	closeDB()
}

func TestTransaktionDB2(t *testing.T) {
	t.Log("TestTransaktionDB2")
	transaktionInit(t, "tr2")

	// Denna testen
	// Gör insättning 0,10kr 9 ggr och kontrollerar att resultatet blir 0,90kr
	// Gör insättning 0,10kr 2 ggr och kontrollerar att resultatet blir 1,10kr
	summa, err := decimal.NewFromString("0.1")
	if err != nil {
		t.Error(err)
	}
	addTransaktionInsättning("Plånboken", "2021-07-27", "Övriga inkomster", "Gemensamt", summa, "Tom € Räksmörgås")
	addTransaktionInsättning("Plånboken", "2021-07-27", "Övriga inkomster", "Gemensamt", summa, "Tom € Räksmörgås")
	addTransaktionInsättning("Plånboken", "2021-07-27", "Övriga inkomster", "Gemensamt", summa, "Tom € Räksmörgås")
	addTransaktionInsättning("Plånboken", "2021-07-27", "Övriga inkomster", "Gemensamt", summa, "Tom € Räksmörgås")
	addTransaktionInsättning("Plånboken", "2021-07-27", "Övriga inkomster", "Gemensamt", summa, "Tom € Räksmörgås")
	addTransaktionInsättning("Plånboken", "2021-07-27", "Övriga inkomster", "Gemensamt", summa, "Tom € Räksmörgås")
	addTransaktionInsättning("Plånboken", "2021-07-27", "Övriga inkomster", "Gemensamt", summa, "Tom € Räksmörgås")
	addTransaktionInsättning("Plånboken", "2021-07-27", "Övriga inkomster", "Gemensamt", summa, "Tom € Räksmörgås")
	addTransaktionInsättning("Plånboken", "2021-07-27", "Övriga inkomster", "Gemensamt", summa, "Tom € Räksmörgås")

	antal := antalTransaktioner(db)

	if antal != 9 {
		t.Error("Antal transaktioner (9) != " + strconv.Itoa(antal))
	} else {
		t.Log("Antal transaktioner ok (9).")
	}

	saldoExpected, err := decimal.NewFromString("0.9")
	konto := hämtaKonto(db, 1)

	if !konto.Saldo.Equal(saldoExpected) {
		t.Error("Konto saldo '" + saldoExpected.String() + "' != '" + konto.Saldo.String() + "'.")
	} else {
		t.Log("Test saldo ok. 0.9")
	}

	saldo := saldoKonto(db, "Plånboken", "2020-07-27")
	if !saldo.Equal(decimal.NewFromInt(0)) {
		t.Error("Konto saldo innan '" + "0,00" + "' != '" + saldo.String() + "'.")
	} else {
		t.Log("Test innan saldo ok.")
	}

	saldo = saldoKonto(db, "Plånboken", "2021-07-28")
	if !saldo.Equal(saldoExpected) {
		t.Error("Konto saldo efter '" + saldo.String() + "' != '" + saldoExpected.String() + "'.")
	} else {
		t.Log("Test efter saldo ok. " + saldoExpected.String())
	}

	addTransaktionInsättning("Plånboken", "2021-07-27", "Övriga inkomster", "Gemensamt", summa, "Tom € Räksmörgås")
	addTransaktionInsättning("Plånboken", "2021-07-27", "Övriga inkomster", "Gemensamt", summa, "Tom € Räksmörgås")

	antal = antalTransaktioner(db)

	if antal != 11 {
		t.Error("Antal transaktioner (11) != " + strconv.Itoa(antal))
	} else {
		t.Log("Antal transaktioner ok (11).")
	}

	saldoExpected, err = decimal.NewFromString("1.1")
	konto = hämtaKonto(db, 1)

	if !konto.Saldo.Equal(saldoExpected) {
		t.Error("Konto saldo '" + saldoExpected.String() + "' != '" + konto.Saldo.String() + "'.")
	} else {
		t.Log("Test saldo ok. 1.1")
	}

	saldo = saldoKonto(db, "Plånboken", "2020-07-27")
	if !saldo.Equal(decimal.NewFromInt(0)) {
		t.Error("Konto saldo innan '" + saldo.String() + "' != '" + "0.00" + "'.")
	} else {
		t.Log("Test innan saldo ok.")
	}

	saldo = saldoKonto(db, "Plånboken", "2021-07-28")
	if !saldo.Equal(saldoExpected) {
		t.Error("Konto saldo efter '" + saldo.String() + "' != '" + saldoExpected.String() + "'.")
	} else {
		t.Log("Test efter saldo ok. " + saldoExpected.String())
	}

	closeDB()
}

func TestTransaktionDB3(t *testing.T) {
	t.Log("TestTransaktionDB3")
	transaktionInit(t, "tr3")

	// Denna testen
	// Kontrollera att vi utgår från startsaldo 0.00
	// Gör insättning 1,20kr
	// Gör inköp 0,10kr kontrollera att saldo blir 1,10kr
	// Gör inköp 0,10kr 2ggr kontrollera att saldo blir 0,90kr
	antal := antalTransaktioner(db)

	if antal != 0 {
		t.Error("Antal transaktioner (0) != " + strconv.Itoa(antal))
	} else {
		t.Log("Antal transaktioner ok (0).")
	}

	saldoExpected := decimal.NewFromInt(0)
	konto := hämtaKonto(db, 1)

	if !konto.Saldo.Equal(saldoExpected) {
		t.Error("Konto saldo '" + saldoExpected.String() + "' != '" + konto.Saldo.String() + "'.")
	} else {
		t.Log("Test saldo ok. 0.00")
	}

	plats := "TestPlats"
	_ = skapaPlats(db, plats, "123-4", true, "")

	summa, err := decimal.NewFromString("1.2")
	if err != nil {
		t.Error(err)
	}
	addTransaktionInsättning("Plånboken", "2021-07-27", "Övriga inkomster", "Gemensamt", summa, "Tom € Räksmörgås")

	antal = antalTransaktioner(db)

	if antal != 1 {
		t.Error("Antal transaktioner (1) != " + strconv.Itoa(antal))
	} else {
		t.Log("Antal transaktioner ok (1).")
	}

	saldoExpected, err = decimal.NewFromString("1.2")
	konto = hämtaKonto(db, 1)

	if !konto.Saldo.Equal(saldoExpected) {
		t.Error("Konto saldo '" + saldoExpected.String() + "' != '" + konto.Saldo.String() + "'.")
	} else {
		t.Log("Test saldo ok. 1.2")
	}

	saldo := saldoKonto(db, "Plånboken", "2021-07-28")
	if !saldo.Equal(saldoExpected) {
		t.Error("Konto saldo efter '" + saldo.String() + "' != '" + saldoExpected.String() + "'.")
	} else {
		t.Log("Test efter saldo ok. " + saldoExpected.String())
	}

	summa, err = decimal.NewFromString("0.1")
	if err != nil {
		t.Error(err)
	}
	addTransaktionInköp("Plånboken", plats, "2021-07-27", "Övriga utgifter", "Gemensamt", summa, "Tom € Räksmörgås")
	konto = hämtaKonto(db, 1)

	saldoExpected, err = decimal.NewFromString("1.1")
	if !konto.Saldo.Equal(saldoExpected) {
		t.Error("Konto saldo '" + saldoExpected.String() + "' != '" + konto.Saldo.String() + "'.")
	} else {
		t.Log("Test saldo ok. 1.1")
	}

	saldo = saldoKonto(db, "Plånboken", "2021-07-28")
	if !saldo.Equal(saldoExpected) {
		t.Error("Konto saldo efter '" + saldo.String() + "' != '" + saldoExpected.String() + "'.")
	} else {
		t.Log("Test efter saldo ok. " + saldoExpected.String())
	}

	addTransaktionInköp("Plånboken", plats, "2021-07-27", "Övriga utgifter", "Gemensamt", summa, "Tom € Räksmörgås")
	addTransaktionInköp("Plånboken", plats, "2021-07-27", "Övriga utgifter", "Gemensamt", summa, "Tom € Räksmörgås")
	konto = hämtaKonto(db, 1)

	saldoExpected, err = decimal.NewFromString("0.9")
	if !konto.Saldo.Equal(saldoExpected) {
		t.Error("Konto saldo '" + saldoExpected.String() + "' != '" + konto.Saldo.String() + "'.")
	} else {
		t.Log("Test saldo ok. 0.9")
	}

	saldo = saldoKonto(db, "Plånboken", "2021-07-28")
	if !saldo.Equal(saldoExpected) {
		t.Error("Konto saldo efter '" + saldo.String() + "' != '" + saldoExpected.String() + "'.")
	} else {
		t.Log("Test efter saldo ok. " + saldoExpected.String())
	}

	closeDB()
}

func TestTransaktionDB4(t *testing.T) {
	t.Log("TestTransaktionDB4")
	transaktionInit(t, "tr4")

	// Denna testen
	// Kontrollera att vi utgår från startsaldo 0.00
	// Gör insättning 1,20kr idag
	// Gör inköp 0,10kr i framtiden kontrollera att saldo blir 1,10kr
	antal := antalTransaktioner(db)

	if antal != 0 {
		t.Error("Antal transaktioner (0) != " + strconv.Itoa(antal))
	} else {
		t.Log("Antal transaktioner ok (0).")
	}

	saldoExpected := decimal.NewFromInt(0)
	konto := hämtaKonto(db, 1)

	if !konto.Saldo.Equal(saldoExpected) {
		t.Error("Konto saldo '" + saldoExpected.String() + "' != '" + konto.Saldo.String() + "'.")
	} else {
		t.Log("Test saldo ok. 0.00")
	}

	plats := "TestPlats"
	_ = skapaPlats(db, plats, "123-0", true, "")

	summa, err := decimal.NewFromString("1.2")
	if err != nil {
		t.Error(err)
	}
	addTransaktionInsättning("Plånboken", "2021-07-27", "Övriga inkomster", "Gemensamt", summa, "Tom € Räksmörgås")

	antal = antalTransaktioner(db)

	if antal != 1 {
		t.Error("Antal transaktioner (1) != " + strconv.Itoa(antal))
	} else {
		t.Log("Antal transaktioner ok (1).")
	}

	saldoExpected, err = decimal.NewFromString("1.2")
	konto = hämtaKonto(db, 1)

	if !konto.Saldo.Equal(saldoExpected) {
		t.Error("Konto saldo '" + saldoExpected.String() + "' != '" + konto.Saldo.String() + "'.")
	} else {
		t.Log("Test saldo ok. 1.2")
	}

	saldo := saldoKonto(db, "Plånboken", "2021-07-28")
	if !saldo.Equal(saldoExpected) {
		t.Error("Konto saldo efter '" + saldo.String() + "' != '" + saldoExpected.String() + "'.")
	} else {
		t.Log("Test efter saldo ok. " + saldoExpected.String())
	}

	summa, err = decimal.NewFromString("0.1")
	if err != nil {
		t.Error(err)
	}
	addTransaktionInköp("Plånboken", plats, "2099-07-27", "Övriga utgifter", "Gemensamt", summa, "Tom € Räksmörgås")
	konto = hämtaKonto(db, 1)

	saldoExpected, err = decimal.NewFromString("1.1")
	if !konto.Saldo.Equal(saldoExpected) {
		t.Error("Konto saldo '" + saldoExpected.String() + "' != '" + konto.Saldo.String() + "'.")
	} else {
		t.Log("Test saldo ok. 1.1")
	}

	saldoExpected, err = decimal.NewFromString("1.2")
	saldo = saldoKonto(db, "Plånboken", "2021-07-28")
	if !saldo.Equal(saldoExpected) {
		t.Error("Konto saldo efter '" + saldo.String() + "' != '" + saldoExpected.String() + "'.")
	} else {
		t.Log("Test efter saldo ok. " + saldoExpected.String())
	}

	saldoExpected, err = decimal.NewFromString("1.1")
	saldo = saldoKonto(db, "Plånboken", "")
	if !saldo.Equal(saldoExpected) {
		t.Error("Konto saldo efter '" + saldo.String() + "' != '" + saldoExpected.String() + "'.")
	} else {
		t.Log("Test efter saldo ok. " + saldoExpected.String())
	}

	closeDB()
}

func TestTransaktionDB5(t *testing.T) {
	t.Log("TestTransaktionDB5")
	transaktionInit(t, "tr5")

	// Denna testen
	// Kommentar klarar citat-tecken
	antal := antalTransaktioner(db)

	if antal != 0 {
		t.Error("Antal transaktioner (0) != " + strconv.Itoa(antal))
	} else {
		t.Log("Antal transaktioner ok (0).")
	}

	saldoExpected := decimal.NewFromInt(0)
	konto := hämtaKonto(db, 1)

	if !konto.Saldo.Equal(saldoExpected) {
		t.Error("Konto saldo '" + saldoExpected.String() + "' != '" + konto.Saldo.String() + "'.")
	} else {
		t.Log("Test saldo ok. 0.00")
	}

	plats := "TestPlats"
	_ = skapaPlats(db, plats, "123-0", true, "")

	summa, err := decimal.NewFromString("1.2")
	if err != nil {
		t.Error(err)
	}
	comment := "Tom '€' \"Räksmörgås\""
	addTransaktionInsättning("Plånboken", "2021-07-27", "Övriga inkomster", "Gemensamt", summa, comment)

	antal = antalTransaktioner(db)

	if antal != 1 {
		t.Error("Antal transaktioner (1) != " + strconv.Itoa(antal))
	} else {
		t.Log("Antal transaktioner ok (1).")
	}

	saldoExpected, err = decimal.NewFromString("1.2")
	trans := hämtaTransaktion(1)

	if trans.comment != comment {
		t.Error("Transaktion text '" + comment + "' != '" + trans.comment + "'.")
	} else {
		t.Log("Test Text ok.", "Tom '€' \"Räksmörgås\"")
	}

	closeDB()
}

/*  Ett test:
        var n float64 = 0
	for i := 0; i < 1000; i++ {
		n += .01
	}
	fmt.Println(n)
*/
