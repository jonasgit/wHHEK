//-*- coding: utf-8 -*-

package main

import (
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/shopspring/decimal" // MIT License
)

func TestBudgetTomMDB1(t *testing.T) {
	t.Parallel()
	db := newTestDB(t, "bdg1")
	if db == nil {
		t.Fatal("Ingen databas.")
	}

	// Denna testen
	antal := antalBudgetposter(db)

	if antal != 34 {
		t.Error("Antal Budgetrader != 34:", antal)
	} else {
		t.Log("Antal budgetrader ok.")
	}

	// loopa över listan med budgettyper och kolla:
	// QUERY = select typ, inkomst from budget order by inkomst, typ
	typdata := [34 * 2]string{
		"Arbetslöshetsersättning", "J",
		"Barnbidrag", "J",
		"Bidragsförskott", "J",
		"Bostadsbidrag", "J",
		"Föräldrapenning", "J",
		"Lön efter skatt", "J",
		"Pension", "J",
		"Sjukpenning", "J",
		"Studiestöd", "J",
		"Underhållsbidrag", "J",
		"Utbildningsbidrag", "J",
		"Övriga inkomster", "J",
		"Amorteringar", "N",
		"Arbetslunch", "N",
		"Barnomsorg", "N",
		"Bil", "N",
		"Bostad/hyra utan lån och ränta", "N",
		"Bostadslån och ränta", "N",
		"Dagstidning, Tel, TV-licens", "N",
		"Fackavgifter", "N",
		"Förbrukn.varor", "N",
		"Hemförsäkring", "N",
		"Hushålls-el", "N",
		"Hygien", "N",
		"Kläder och skor", "N",
		"Kollektiva resor", "N",
		"Lek och fritid", "N",
		"Livsmedel", "N",
		"Läkare/tandläkare/medicin", "N",
		"Möbler, husgeråd, TV, radio", "N",
		"Räntor", "N",
		"Underhåll till barn", "N",
		"Övriga utg.-lån", "N",
		"Övriga utgifter", "N",
	}
	budgetposter := getAllBudgetposter(db)

	for i, s := range budgetposter {
		if (typdata[i*2] == s[0]) &&
			(typdata[i*2+1] == s[1]) {
			t.Log("Budgetpost ok. ", typdata[i*2])
		} else {
			t.Error("Budgetpost stämmer inte överens:", typdata[i*2], s[0], typdata[i*2+1], s[1])
		}
	}
}

func TestBudgetUppdateraDecimal(t *testing.T) {
	t.Parallel()
	db := newTestDB(t, "bdg2")
	var lopnr int
	if err := db.QueryRow("SELECT Löpnr FROM Budget WHERE Typ = 'Livsmedel'").Scan(&lopnr); err != nil {
		t.Fatal(err)
	}
	form := url.Values{"Typ": {"Livsmedel"}, "Inkomst": {"N"}, "HurOfta": {"1"}, "StartMånad": {"Jan"},
		"Jan": {"100.50"}, "Feb": {"1 000,25"}}
	for _, m := range []string{"Mar", "Apr", "Maj", "Jun", "Jul", "Aug", "Sep", "Okt", "Nov", "Dec"} {
		form.Set(m, "0")
	}
	req := httptest.NewRequest("POST", "/budget", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := req.ParseForm(); err != nil {
		t.Fatal(err)
	}
	if _, err := updateBudget(lopnr, req, db); err != nil {
		t.Fatal(err)
	}
	if got, want := getKatYearSum(db, "Livsmedel"), decimal.RequireFromString("1100.75"); !got.Equal(want) {
		t.Errorf("årssumma: förväntade %s, fick %s", want, got)
	}

	// Så sparade updateBudget tidigare belopp i SQLite: ska läsas utan krasch.
	if _, err := db.Exec("UPDATE Budget SET Mar = '50,25' WHERE Löpnr = ?", lopnr); err != nil {
		t.Fatal(err)
	}
	if got, want := getKatYearSum(db, "Livsmedel"), decimal.RequireFromString("1151"); !got.Equal(want) {
		t.Errorf("årssumma med gammalt kommatecken: förväntade %s, fick %s", want, got)
	}
}
