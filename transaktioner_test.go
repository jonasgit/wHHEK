//-*- coding: utf-8 -*-
package main

import (
	"strconv"
	"testing"

	"github.com/shopspring/decimal"  // MIT License
)

func transaktionInit(t *testing.T, filnamn string) {
	// Förberedelser
	var filename string = "got"+filnamn+".mdb"

	SkapaTomMDB(t, filename)
	db = openJetDB(filename, false)
}

func TestTransaktionTomMDB1(t *testing.T) {
	transaktionInit(t, "trt")

	// Denna testen
	antal := antalTransaktioner()
	
	if antal != 0 {
		t.Error("Antal transaktioner != (0).")
	} else {
		t.Log("Antal transaktioner ok (0).")
	}

	closeDB()
}

func TestTransaktionMDB1(t *testing.T) {
	transaktionInit(t, "tr1")

	// Denna testen
	// Kontrollera att vi utgår från startsaldo 0.00
	// Gör insättning 0,10kr och kontrollerar att resultatet blir 0,10kr
	antal := antalTransaktioner()
	
	if antal != 0 {
		t.Error("Antal transaktioner (0) != "+strconv.Itoa(antal))
	} else {
		t.Log("Antal transaktioner ok (0).")
	}

	saldoExpected := decimal.NewFromInt(0)
	konto := hämtaKonto(1)
	
	if !konto.Saldo.Equal(saldoExpected) {
		t.Error("Konto saldo '"+saldoExpected.String()+"' != '"+konto.Saldo.String()+"'.")
	} else {
		t.Log("Test saldo ok. 0.00")
	}

	summa, err := decimal.NewFromString("0.1")
	if err != nil {
		t.Error(err)
	}
	addTransaktionInsättning("Plånboken", "2021-07-27", "Övriga inkomster", "Gemensamt", summa, "Tom € Räksmörgås")

	antal = antalTransaktioner()
	
	if antal != 1 {
		t.Error("Antal transaktioner (1) != "+strconv.Itoa(antal))
	} else {
		t.Log("Antal transaktioner ok (1).")
	}

	saldoExpected, err = decimal.NewFromString("0.1")
	konto = hämtaKonto(1)
	
	if !konto.Saldo.Equal(saldoExpected) {
		t.Error("Konto saldo '"+saldoExpected.String()+"' != '"+konto.Saldo.String()+"'.")
	} else {
		t.Log("Test saldo ok. 0.1")
	}

	saldo := saldoKonto("Plånboken", "2020-07-27")
	if !saldo.Equal(decimal.NewFromInt(0)) {
		t.Error("Konto saldo innan '"+"0,00"+"' != '"+saldo.String()+"'.")
	} else {
		t.Log("Test innan saldo ok.")
	}

	saldo = saldoKonto("Plånboken", "2021-07-28")
	if !saldo.Equal(saldoExpected) {
		t.Error("Konto saldo efter '"+saldo.String()+"' != '"+saldoExpected.String()+"'.")
	} else {
		t.Log("Test efter saldo ok. "+saldoExpected.String())
	}

	closeDB()
}

func TestTransaktionMDB2(t *testing.T) {
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

	antal := antalTransaktioner()
	
	if antal != 9 {
		t.Error("Antal transaktioner (9) != "+strconv.Itoa(antal))
	} else {
		t.Log("Antal transaktioner ok (9).")
	}

	saldoExpected, err := decimal.NewFromString("0.9")
	konto := hämtaKonto(1)
	
	if !konto.Saldo.Equal(saldoExpected) {
		t.Error("Konto saldo '"+saldoExpected.String()+"' != '"+konto.Saldo.String()+"'.")
	} else {
		t.Log("Test saldo ok. 0.9")
	}

	saldo := saldoKonto("Plånboken", "2020-07-27")
	if !saldo.Equal(decimal.NewFromInt(0)) {
		t.Error("Konto saldo innan '"+"0,00"+"' != '"+saldo.String()+"'.")
	} else {
		t.Log("Test innan saldo ok.")
	}

	saldo = saldoKonto("Plånboken", "2021-07-28")
	if !saldo.Equal(saldoExpected) {
		t.Error("Konto saldo efter '"+saldo.String()+"' != '"+saldoExpected.String()+"'.")
	} else {
		t.Log("Test efter saldo ok. "+saldoExpected.String())
	}
	
	addTransaktionInsättning("Plånboken", "2021-07-27", "Övriga inkomster", "Gemensamt", summa, "Tom € Räksmörgås")
	addTransaktionInsättning("Plånboken", "2021-07-27", "Övriga inkomster", "Gemensamt", summa, "Tom € Räksmörgås")

	antal = antalTransaktioner()
	
	if antal != 11 {
		t.Error("Antal transaktioner (11) != "+strconv.Itoa(antal))
	} else {
		t.Log("Antal transaktioner ok (11).")
	}

	saldoExpected, err = decimal.NewFromString("1.1")
	konto = hämtaKonto(1)
	
	if !konto.Saldo.Equal(saldoExpected) {
		t.Error("Konto saldo '"+saldoExpected.String()+"' != '"+konto.Saldo.String()+"'.")
	} else {
		t.Log("Test saldo ok. 1.1")
	}

	saldo = saldoKonto("Plånboken", "2020-07-27")
	if !saldo.Equal(decimal.NewFromInt(0)) {
		t.Error("Konto saldo innan '"+saldo.String()+"' != '"+"0.00"+"'.")
	} else {
		t.Log("Test innan saldo ok.")
	}

	saldo = saldoKonto("Plånboken", "2021-07-28")
	if !saldo.Equal(saldoExpected) {
		t.Error("Konto saldo efter '"+saldo.String()+"' != '"+saldoExpected.String()+"'.")
	} else {
		t.Log("Test efter saldo ok. "+saldoExpected.String())
	}
	
	closeDB()
}

func TestTransaktionMDB3(t *testing.T) {
	transaktionInit(t, "tr3")

	// Denna testen
	// Kontrollera att vi utgår från startsaldo 0.00
	// Gör insättning 1,20kr
	// Gör inköp 0,10kr kontrollera att saldo blir 1,10kr
	// Gör inköp 0,10kr 2ggr kontrollera att saldo blir 0,90kr
	antal := antalTransaktioner()
	
	if antal != 0 {
		t.Error("Antal transaktioner (0) != "+strconv.Itoa(antal))
	} else {
		t.Log("Antal transaktioner ok (0).")
	}

	saldoExpected := decimal.NewFromInt(0)
	konto := hämtaKonto(1)
	
	if !konto.Saldo.Equal(saldoExpected) {
		t.Error("Konto saldo '"+saldoExpected.String()+"' != '"+konto.Saldo.String()+"'.")
	} else {
		t.Log("Test saldo ok. 0.00")
	}

	plats := "TestPlats"
	skapaPlats(plats, "123-4", true, "")

	summa, err := decimal.NewFromString("1.2")
	if err != nil {
		t.Error(err)
	}
	addTransaktionInsättning("Plånboken", "2021-07-27", "Övriga inkomster", "Gemensamt", summa, "Tom € Räksmörgås")

	antal = antalTransaktioner()
	
	if antal != 1 {
		t.Error("Antal transaktioner (1) != "+strconv.Itoa(antal))
	} else {
		t.Log("Antal transaktioner ok (1).")
	}

	saldoExpected, err = decimal.NewFromString("1.2")
	konto = hämtaKonto(1)
	
	if !konto.Saldo.Equal(saldoExpected) {
		t.Error("Konto saldo '"+saldoExpected.String()+"' != '"+konto.Saldo.String()+"'.")
	} else {
		t.Log("Test saldo ok. 1.2")
	}

	saldo := saldoKonto("Plånboken", "2021-07-28")
	if !saldo.Equal(saldoExpected) {
		t.Error("Konto saldo efter '"+saldo.String()+"' != '"+saldoExpected.String()+"'.")
	} else {
		t.Log("Test efter saldo ok. "+saldoExpected.String())
	}

	summa, err = decimal.NewFromString("0.1")
	if err != nil {
		t.Error(err)
	}
	addTransaktionInköp("Plånboken", plats, "2021-07-27", "Övriga utgifter", "Gemensamt", summa, "Tom € Räksmörgås")
	konto = hämtaKonto(1)
	
	saldoExpected, err = decimal.NewFromString("1.1")
	if !konto.Saldo.Equal(saldoExpected) {
		t.Error("Konto saldo '"+saldoExpected.String()+"' != '"+konto.Saldo.String()+"'.")
	} else {
		t.Log("Test saldo ok. 1.1")
	}

	saldo = saldoKonto("Plånboken", "2021-07-28")
	if !saldo.Equal(saldoExpected) {
		t.Error("Konto saldo efter '"+saldo.String()+"' != '"+saldoExpected.String()+"'.")
	} else {
		t.Log("Test efter saldo ok. "+saldoExpected.String())
	}

	addTransaktionInköp("Plånboken", plats, "2021-07-27", "Övriga utgifter", "Gemensamt", summa, "Tom € Räksmörgås")
	addTransaktionInköp("Plånboken", plats, "2021-07-27", "Övriga utgifter", "Gemensamt", summa, "Tom € Räksmörgås")
	konto = hämtaKonto(1)
	
	saldoExpected, err = decimal.NewFromString("0.9")
	if !konto.Saldo.Equal(saldoExpected) {
		t.Error("Konto saldo '"+saldoExpected.String()+"' != '"+konto.Saldo.String()+"'.")
	} else {
		t.Log("Test saldo ok. 0.9")
	}

	saldo = saldoKonto("Plånboken", "2021-07-28")
	if !saldo.Equal(saldoExpected) {
		t.Error("Konto saldo efter '"+saldo.String()+"' != '"+saldoExpected.String()+"'.")
	} else {
		t.Log("Test efter saldo ok. "+saldoExpected.String())
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
