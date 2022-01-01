//-*- coding: utf-8 -*-

package main

import (
	"database/sql"
	"testing"
	"time"

	"github.com/shopspring/decimal" // MIT License
)

func fasttransaktionInit(t *testing.T, filnamn string) *sql.DB {
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

func TestFastTransaktionTomDB1(t *testing.T) {
	db = fasttransaktionInit(t, "ftrt1")

	// Denna testen
	antal := antalFastaTransaktioner(db)

	if antal != 0 {
		t.Error("Antal fasta transaktioner != (0).")
	} else {
		t.Log("Antal fasta transaktioner ok (0).")
	}

	closeDB()
}

func compareDates(t *testing.T, t1 time.Time, t2 time.Time) bool {
	/*	t.Log("t1.Year: ", t1.Year())
		t.Log("t1.Month: ", int(t1.Month()))
		t.Log("t1.Day: ", t1.Day())
		t.Log("t2.Year: ", t2.Year())
		t.Log("t2.Month: ", int(t2.Month()))
		t.Log("t2.Day: ", t2.Day())*/

	if t1.Year() != t2.Year() {
		//t.Log("False")
		return false
	}
	if int(t1.Month()) != int(t2.Month()) {
		//t.Log("False")
		return false
	}
	if t1.Day() != t2.Day() {
		//t.Log("False")
		return false
	}
	t.Log("True")
	return true
}

func TestFastTransaktionTomDB2(t *testing.T) {
	// Förbered test
	db = fasttransaktionInit(t, "ftrt2")
	_ = skapaPlats(db, "Mack", "", false, "")

	antal := antalTransaktioner(db)
	if antal != 0 {
		t.Error("Antal transaktioner != (0).")
	} else {
		t.Log("Antal transaktioner ok (0).")
	}

	// Denna testen: Fast utgift, månadsvis, testa datum slut på månad
	// Skapa ny fast transaktion, datum 2021-01-31, månadsvis
	summa := decimal.NewFromInt(1)

	_ = skapaFastUtgift(db, "Bil", "Plånboken", "Gemensamt", "Mack", summa, "2021-01-31", "Varje månad")

	antal = antalFastaTransaktioner(db)
	if antal != 1 {
		t.Error("Antal fasta transaktioner != (1).")
	} else {
		t.Log("Antal fasta transaktioner ok (1).")
	}

	// Registrera denna 1 gång
	t.Log("pre-reg 1.")
	registreraFastTransaktion(db, 1)
	t.Log("post-reg 1.")
	antal = antalFastaTransaktioner(db)
	if antal != 1 {
		t.Error("Antal fasta transaktioner != (1).")
	} else {
		t.Log("Antal fasta transaktioner ok (1).")
	}
	antal = antalTransaktioner(db)
	if antal != 1 {
		t.Error("Antal transaktioner != (1).")
	} else {
		t.Log("Antal transaktioner ok (1).")
	}
	// Kontrollera resultatet, nästa datum = 2021-02-28
	ft := hämtaFastTransaktion(db, 1)
	expectedDate, _ := time.Parse("2006-01-02", "2021-02-28")
	if compareDates(t, ft.date, expectedDate) {
		t.Log("Förväntat datum ok (2021-02-28).")
	} else {
		t.Error("Förväntat datum 2021-02-28 != ", ft.date)
	}
	// Registrera denna 1 gång
	registreraFastTransaktion(db, 1)
	antal = antalTransaktioner(db)
	if antal != 2 {
		t.Error("Antal transaktioner != (2).")
	} else {
		t.Log("Antal transaktioner ok (2).")
	}
	// Kontrollera resultatet, nästa datum = 2021-03-28
	// Registrera denna 1 gång
	registreraFastTransaktion(db, 1)
	antal = antalTransaktioner(db)
	// Kontrollera resultatet, datum = 2021-03-28, saldo -3kr
	if antal != 3 {
		t.Error("Antal transaktioner != (3).")
	} else {
		t.Log("Antal transaktioner ok (3).")
	}
	expectedDate, _ = time.Parse("2006-01-02", "2021-04-28")
	ft = hämtaFastTransaktion(db, 1)
	if compareDates(t, ft.date, expectedDate) {
		t.Log("Förväntat datum ok (2021-04-28).")
	} else {
		t.Error("Förväntat datum 2021-04-28 != ", ft.date)
	}

	currentTime := time.Now()
	currDate := currentTime.Format("2006-01-02")
	_, totSaldo := saldonKonto(db, "Plånboken", currDate)
	expSaldo := decimal.NewFromInt(-3)
	if totSaldo.Equal(expSaldo) {
		t.Log("Förväntat saldo ok (-3,00).")
	} else {
		t.Error("Förväntat saldo -3,00 != ", totSaldo, expSaldo)
	}

	closeDB()
}

func TestFastTransaktionTomDB3(t *testing.T) {
	// Förbered test
	db = fasttransaktionInit(t, "ftrt3")
	_ = skapaPlats(db, "Mack", "", false, "")

	antal := antalTransaktioner(db)
	if antal != 0 {
		t.Error("Antal transaktioner != (0).")
	} else {
		t.Log("Antal transaktioner ok (0).")
	}

	// Denna testen: Fast utgift, kvartalsvis, testa datum slut på månad

	// Skapa ny fast transaktion, datum 2020-11-30, kvartalsvis
	summa := decimal.NewFromInt(1)

	_ = skapaFastUtgift(db, "Bil", "Plånboken", "Gemensamt", "Mack", summa, "2020-11-30", "Varje kvartal")

	antal = antalFastaTransaktioner(db)
	if antal != 1 {
		t.Error("Antal fasta transaktioner != (1).")
	} else {
		t.Log("Antal fasta transaktioner ok (1).")
	}

	// Registrera denna 1 gång
	registreraFastTransaktion(db, 1)
	antal = antalFastaTransaktioner(db)
	if antal != 1 {
		t.Error("Antal fasta transaktioner != (1).")
	} else {
		t.Log("Antal fasta transaktioner ok (1).")
	}
	antal = antalTransaktioner(db)
	if antal != 1 {
		t.Error("Antal transaktioner != (1).")
	} else {
		t.Log("Antal transaktioner ok (1).")
	}
	// Kontrollera resultatet, nästa datum = 2021-02-28
	ft := hämtaFastTransaktion(db, 1)
	expectedDate, _ := time.Parse("2006-01-02", "2021-02-28")
	if compareDates(t, ft.date, expectedDate) {
		t.Log("Förväntat datum ok (2021-02-28).")
	} else {
		t.Error("Förväntat datum 2021-02-28 != ", ft.date)
	}
	// Registrera denna 1 gång
	registreraFastTransaktion(db, 1)
	antal = antalTransaktioner(db)
	if antal != 2 {
		t.Error("Antal transaktioner != (2).")
	} else {
		t.Log("Antal transaktioner ok (2).")
	}
	// Kontrollera resultatet, nästa datum = 2021-05-28
	// Registrera denna 1 gång
	registreraFastTransaktion(db, 1)
	antal = antalTransaktioner(db)
	// Kontrollera resultatet, datum = 2021-05-28, saldo -3kr
	if antal != 3 {
		t.Error("Antal transaktioner != (3).")
	} else {
		t.Log("Antal transaktioner ok (3).")
	}
	expectedDate, _ = time.Parse("2006-01-02", "2021-08-28")
	ft = hämtaFastTransaktion(db, 1)
	if compareDates(t, ft.date, expectedDate) {
		t.Log("Förväntat datum ok (2021-08-28).")
	} else {
		t.Error("Förväntat datum 2021-08-28 != ", ft.date)
	}

	currentTime := time.Now()
	currDate := currentTime.Format("2006-01-02")
	_, totSaldo := saldonKonto(db, "Plånboken", currDate)
	expSaldo := decimal.NewFromInt(-3)
	if totSaldo.Equal(expSaldo) {
		t.Log("Förväntat saldo ok (-3,00).")
	} else {
		t.Error("Förväntat saldo -3,00 != ", totSaldo, expSaldo)
	}

	closeDB()
}
