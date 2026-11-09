// -*- coding: utf-8 -*-
package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/shopspring/decimal" // MIT License
)

func TestTransaktionTomDB1(t *testing.T) {
	t.Parallel()
	t.Log("TestTransaktionTomDB1")
	db := newTestDB(t, "trt")

	// Denna testen
	antal := antalTransaktioner(db)

	if antal != 0 {
		t.Error("Antal transaktioner != (0).")
	} else {
		t.Log("Antal transaktioner ok (0).")
	}

}

func TestTransaktionDB1(t *testing.T) {
	t.Parallel()
	t.Log("TestTransaktionDB1")
	db := newTestDB(t, "tr1")

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
	addTransaktionInsättning(db, "Plånboken", "2021-07-27", "Övriga inkomster", "Gemensamt", summa, "Tom € Räksmörgås")

	antal = antalTransaktioner(db)

	if antal != 1 {
		t.Error("Antal transaktioner (1) != " + strconv.Itoa(antal))
	} else {
		t.Log("Antal transaktioner ok (1).")
	}

	saldoExpected, _ = decimal.NewFromString("0.1")
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

}

func TestTransaktionDB2(t *testing.T) {
	t.Parallel()
	t.Log("TestTransaktionDB2")
	db := newTestDB(t, "tr2")

	// Denna testen
	// Gör insättning 0,10kr 9 ggr och kontrollerar att resultatet blir 0,90kr
	// Gör insättning 0,10kr 2 ggr och kontrollerar att resultatet blir 1,10kr
	summa, err := decimal.NewFromString("0.1")
	if err != nil {
		t.Error(err)
	}
	addTransaktionInsättning(db, "Plånboken", "2021-07-27", "Övriga inkomster", "Gemensamt", summa, "Tom € Räksmörgås")
	addTransaktionInsättning(db, "Plånboken", "2021-07-27", "Övriga inkomster", "Gemensamt", summa, "Tom € Räksmörgås")
	addTransaktionInsättning(db, "Plånboken", "2021-07-27", "Övriga inkomster", "Gemensamt", summa, "Tom € Räksmörgås")
	addTransaktionInsättning(db, "Plånboken", "2021-07-27", "Övriga inkomster", "Gemensamt", summa, "Tom € Räksmörgås")
	addTransaktionInsättning(db, "Plånboken", "2021-07-27", "Övriga inkomster", "Gemensamt", summa, "Tom € Räksmörgås")
	addTransaktionInsättning(db, "Plånboken", "2021-07-27", "Övriga inkomster", "Gemensamt", summa, "Tom € Räksmörgås")
	addTransaktionInsättning(db, "Plånboken", "2021-07-27", "Övriga inkomster", "Gemensamt", summa, "Tom € Räksmörgås")
	addTransaktionInsättning(db, "Plånboken", "2021-07-27", "Övriga inkomster", "Gemensamt", summa, "Tom € Räksmörgås")
	addTransaktionInsättning(db, "Plånboken", "2021-07-27", "Övriga inkomster", "Gemensamt", summa, "Tom € Räksmörgås")

	antal := antalTransaktioner(db)

	if antal != 9 {
		t.Error("Antal transaktioner (9) != " + strconv.Itoa(antal))
	} else {
		t.Log("Antal transaktioner ok (9).")
	}

	saldoExpected, _ := decimal.NewFromString("0.9")
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

	addTransaktionInsättning(db, "Plånboken", "2021-07-27", "Övriga inkomster", "Gemensamt", summa, "Tom € Räksmörgås")
	addTransaktionInsättning(db, "Plånboken", "2021-07-27", "Övriga inkomster", "Gemensamt", summa, "Tom € Räksmörgås")

	antal = antalTransaktioner(db)

	if antal != 11 {
		t.Error("Antal transaktioner (11) != " + strconv.Itoa(antal))
	} else {
		t.Log("Antal transaktioner ok (11).")
	}

	saldoExpected, _ = decimal.NewFromString("1.1")
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

}

func TestTransaktionDB3(t *testing.T) {
	t.Parallel()
	t.Log("TestTransaktionDB3")
	db := newTestDB(t, "tr3")

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
	addTransaktionInsättning(db, "Plånboken", "2021-07-27", "Övriga inkomster", "Gemensamt", summa, "Tom € Räksmörgås")

	antal = antalTransaktioner(db)

	if antal != 1 {
		t.Error("Antal transaktioner (1) != " + strconv.Itoa(antal))
	} else {
		t.Log("Antal transaktioner ok (1).")
	}

	saldoExpected, _ = decimal.NewFromString("1.2")
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
	addTransaktionInköp(db, "Plånboken", plats, "2021-07-27", "Övriga utgifter", "Gemensamt", summa, "Tom € Räksmörgås", false)
	konto = hämtaKonto(db, 1)

	saldoExpected, _ = decimal.NewFromString("1.1")
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

	addTransaktionInköp(db, "Plånboken", plats, "2021-07-27", "Övriga utgifter", "Gemensamt", summa, "Tom € Räksmörgås", false)
	addTransaktionInköp(db, "Plånboken", plats, "2021-07-27", "Övriga utgifter", "Gemensamt", summa, "Tom € Räksmörgås", false)
	konto = hämtaKonto(db, 1)

	saldoExpected, _ = decimal.NewFromString("0.9")
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

}

func TestTransaktionDB4(t *testing.T) {
	t.Parallel()
	t.Log("TestTransaktionDB4")
	db := newTestDB(t, "tr4")

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
	addTransaktionInsättning(db, "Plånboken", "2021-07-27", "Övriga inkomster", "Gemensamt", summa, "Tom € Räksmörgås")

	antal = antalTransaktioner(db)

	if antal != 1 {
		t.Error("Antal transaktioner (1) != " + strconv.Itoa(antal))
	} else {
		t.Log("Antal transaktioner ok (1).")
	}

	saldoExpected, _ = decimal.NewFromString("1.2")
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
	addTransaktionInköp(db, "Plånboken", plats, "2099-07-27", "Övriga utgifter", "Gemensamt", summa, "Tom € Räksmörgås", false)
	konto = hämtaKonto(db, 1)

	saldoExpected, _ = decimal.NewFromString("1.1")
	if !konto.Saldo.Equal(saldoExpected) {
		t.Error("Konto saldo '" + saldoExpected.String() + "' != '" + konto.Saldo.String() + "'.")
	} else {
		t.Log("Test saldo ok. 1.1")
	}

	saldoExpected, _ = decimal.NewFromString("1.2")
	saldo = saldoKonto(db, "Plånboken", "2021-07-28")
	if !saldo.Equal(saldoExpected) {
		t.Error("Konto saldo efter '" + saldo.String() + "' != '" + saldoExpected.String() + "'.")
	} else {
		t.Log("Test efter saldo ok. " + saldoExpected.String())
	}

	saldoExpected, _ = decimal.NewFromString("1.1")
	saldo = saldoKonto(db, "Plånboken", "")
	if !saldo.Equal(saldoExpected) {
		t.Error("Konto saldo efter '" + saldo.String() + "' != '" + saldoExpected.String() + "'.")
	} else {
		t.Log("Test efter saldo ok. " + saldoExpected.String())
	}

}

func TestTransaktionDB5(t *testing.T) {
	t.Parallel()
	t.Log("TestTransaktionDB5")
	db := newTestDB(t, "tr5")

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
	addTransaktionInsättning(db, "Plånboken", "2021-07-27", "Övriga inkomster", "Gemensamt", summa, comment)

	antal = antalTransaktioner(db)

	if antal != 1 {
		t.Error("Antal transaktioner (1) != " + strconv.Itoa(antal))
	} else {
		t.Log("Antal transaktioner ok (1).")
	}

	trans := hämtaTransaktion(db, 1)

	if trans.comment != comment {
		t.Error("Transaktion text '" + comment + "' != '" + trans.comment + "'.")
	} else {
		t.Log("Test Text ok.", "Tom '€' \"Räksmörgås\"")
	}

}

func TestTransaktionDB6(t *testing.T) {
	t.Parallel()
	t.Log("TestTransaktionDB6")
	db := newTestDB(t, "tr6")

	// Denna testen
	// Kontrollera att vi utgår från startsaldo 0.00
	// Gör insättning 0,10kr utan text/kommentar
	// och kontrollerar att kommentaren blir " "
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
	addTransaktionInsättning(db, "Plånboken", "2021-07-27", "Övriga inkomster", "Gemensamt", summa, "")

	antal = antalTransaktioner(db)

	if antal != 1 {
		t.Error("Antal transaktioner (1) != " + strconv.Itoa(antal))
	} else {
		t.Log("Antal transaktioner ok (1).")
	}

	trans := hämtaTransaktion(db, 1)

	if trans.comment != " " {
		t.Error("Transaktion text ' ' != '" + trans.comment + "'.")
	} else {
		t.Log("Test Text ok.", "' '")
	}

}

func TestTransaktionDB7(t *testing.T) {
	t.Parallel()
	t.Log("TestTransaktionDB7")
	db := newTestDB(t, "tr7")

	// Denna testen
	// Kontrollera att vi utgår från startsaldo 0.00
	// Gör insättning 0,10kr utan text/kommentar
	// och kontrollerar att kommentaren blir " "
	// Gör uppdatering till 0,20kr utan text/kommentar
	// och kontrollerar att kommentaren blir " "
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
	addTransaktionInsättning(db, "Plånboken", "2021-07-27", "Övriga inkomster", "Gemensamt", summa, "")

	antal = antalTransaktioner(db)

	if antal != 1 {
		t.Error("Antal transaktioner (1) != " + strconv.Itoa(antal))
	} else {
		t.Log("Antal transaktioner ok (1).")
	}

	trans := hämtaTransaktion(db, 1)

	if trans.comment != " " {
		t.Error("Transaktion text ' ' != '" + trans.comment + "'.")
	} else {
		t.Log("Test Text ok.", "' '")
	}

	summa, err = decimal.NewFromString("0.2")
	if err != nil {
		t.Error(err)
	}
	var transtyp = "Insättning"
	var fromAcc = "---"
	updateTransactionSQL(1, db, fromAcc, "Plånboken", transtyp, "2021-07-27", "Övriga inkomster", "Gemensamt", summa.String(), false, "")
	antal = antalTransaktioner(db)

	if antal != 1 {
		t.Error("Antal transaktioner (1) != " + strconv.Itoa(antal))
	} else {
		t.Log("Antal transaktioner ok (1).")
	}

	trans = hämtaTransaktion(db, 1)

	if trans.comment != " " {
		t.Error("Transaktion text ' ' != '" + trans.comment + "'.")
	} else {
		t.Log("Test Text ok.", "' '")
	}

	saldoExpected, _ = decimal.NewFromString("0.2")
	saldo := saldoKonto(db, "Plånboken", "2021-07-28")
	if !saldo.Equal(saldoExpected) {
		t.Error("Konto saldo efter '" + saldo.String() + "' != '" + saldoExpected.String() + "'.")
	} else {
		t.Log("Test efter saldo ok. " + saldoExpected.String())
	}

}

/*  Ett test:
        var n float64 = 0
	for i := 0; i < 1000; i++ {
		n += .01
	}
	fmt.Println(n)
*/

func postAddTransaction(t *testing.T, amount string) string {
	t.Helper()
	form := url.Values{"transtyp": {"Uttag"}, "fromacc": {"Plånboken"}, "date": {"2026-11-09"},
		"who": {"Gemensamt"}, "amount": {amount}, "text": {""}}
	req := httptest.NewRequest(http.MethodPost, "/hnewtrans", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	addtransaction(rec, req)
	return rec.Body.String()
}

// addtransaction använder den globala db: inte parallell.
func TestAddTransactionOgiltigtBelopp(t *testing.T) {
	keepGlobals(t)
	db = newTestDB(t, "trinv")

	body := postAddTransaction(t, "-")
	if !strings.Contains(body, "sparades inte") {
		t.Errorf("förväntade felmeddelande, fick %q", body)
	}
	if n := antalTransaktioner(db); n != 0 {
		t.Errorf("ogiltigt belopp sparades: %d transaktioner", n)
	}

	postAddTransaction(t, "1-000")
	if n := antalTransaktioner(db); n != 1 {
		t.Fatalf("förväntade 1 transaktion, fick %d", n)
	}
	if got, want := hämtaTransaktion(db, 1).amount, decimal.NewFromInt(1000); !got.Equal(want) {
		t.Errorf("belopp: förväntade %s, fick %s", want, got)
	}
}
