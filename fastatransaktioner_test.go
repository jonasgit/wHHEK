//-*- coding: utf-8 -*-

package main

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"  // MIT License
)

func fasttransaktionInit(t *testing.T, filnamn string) {
	// Förberedelser
	var filename string = "got"+filnamn+".mdb"
	
	SkapaTomMDB(t, filename)
	db = openJetDB(filename, false)
	if db == nil {
 		t.Fatal("Ingen databas.")
	}
}

func TestFastTransaktionTomMDB1(t *testing.T) {
	fasttransaktionInit(t, "ftrt1")
	
	// Denna testen
	antal := antalFastaTransaktioner()
	
	if antal != 0 {
		t.Error("Antal fasta transaktioner != (0).")
	} else {
		t.Log("Antal fasta transaktioner ok (0).")
	}
	
	closeDB()
}

func TestFastTransaktionTomMDB2(t *testing.T) {
	// Förbered test
	fasttransaktionInit(t, "ftrt2")
	skapaPlats("Mack", "", false, "")

	antal := antalTransaktioner()
	if antal != 0 {
		t.Error("Antal transaktioner != (0).")
	} else {
		t.Log("Antal transaktioner ok (0).")
	}

	// Denna testen: Fast utgift, månadsvis, testa datum slut på månad

	// Skapa ny fast transaktion, datum 2021-01-31, månadsvis
	summa := decimal.NewFromInt(1)

	skapaFastUtgift("Bil", "Plånboken", "Gemensamt", "Mack", summa, "2021-01-31", false, false, "registrera", "Varje månad")

	antal = antalFastaTransaktioner()
	if antal != 1 {
		t.Error("Antal fasta transaktioner != (1).")
	} else {
		t.Log("Antal fasta transaktioner ok (1).")
	}
	
	// Registrera denna 1 gång
	registreraFastTransaktion(1)
	antal = antalFastaTransaktioner()
	if antal != 1 {
		t.Error("Antal fasta transaktioner != (1).")
	} else {
		t.Log("Antal fasta transaktioner ok (1).")
	}
	antal = antalTransaktioner()
	if antal != 1 {
		t.Error("Antal transaktioner != (1).")
	} else {
		t.Log("Antal transaktioner ok (1).")
	}
	// Kontrollera resultatet, nästa datum = 2021-02-28
	ft := hämtaFastTransaktion(1)
	expectedDate, _ := time.Parse("2006-01-02", "2021-02-28")
	if ft.date != expectedDate {
		t.Error("Förväntat datum 2021-02-28 != ", ft.date)
	} else {
		t.Log("Förväntat datum ok (2021-02-28).")
	}
	// Registrera denna 1 gång
	registreraFastTransaktion(1)
	antal = antalTransaktioner()
	if antal != 2 {
		t.Error("Antal transaktioner != (2).")
	} else {
		t.Log("Antal transaktioner ok (2).")
	}
	// Kontrollera resultatet, nästa datum = 2021-03-28
	// Registrera denna 1 gång
	registreraFastTransaktion(1)
	antal = antalTransaktioner()
	// Kontrollera resultatet, datum = 2021-03-28, saldo -3kr
	if antal != 3 {
		t.Error("Antal transaktioner != (3).")
	} else {
		t.Log("Antal transaktioner ok (3).")
	}
	expectedDate, _ = time.Parse("2006-01-02", "2021-03-28")
	if ft.date != expectedDate {
		t.Error("Förväntat datum 2021-02-28 != ", ft.date)
	} else {
		t.Log("Förväntat datum ok (2021-02-28).")
	}

	currentTime := time.Now()
	currDate := currentTime.Format("2006-01-02")
	_, totSaldo := saldonKonto("Plånboken", currDate)
	expSaldo := decimal.NewFromInt(-3)
	if totSaldo.Equal(expSaldo) {
		t.Log("Förväntat saldo ok (-3,00).")
	} else {
		t.Error("Förväntat saldo -3,00 != ", totSaldo, expSaldo)
	}

	closeDB()
}
