//-*- coding: utf-8 -*-

package main

import (
	"database/sql"
	"testing"
)

func budgetInit(t *testing.T, filnamn string) *sql.DB {
	// Förberedelser
	if JetDBSupport {
		t.Log("Jet Supported.")
		var filename = "got" + filnamn + ".mdb"
		SkapaTomMDB(t, filename)
		db = openJetDB(filename, false)
	} else {
		t.Log("Jet NOT Supported.")
		var filename = "got" + filnamn + ".db"
		SkapaTomDB(filename)
		db = openSqlite(filename)
	}

	if db == nil {
		t.Fatal("Ingen databas.")
	}

	return db
}

func TestBudgetTomMDB1(t *testing.T) {
	budgetInit(t, "bdg1")
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
	closeDB()
}
