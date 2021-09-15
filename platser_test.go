//-*- coding: utf-8 -*-

package main

import (
	"strconv"
	"testing"
)

func platserInit(t *testing.T) {
	// Förberedelser
	var filename string = "gotestpl.mdb"

	SkapaTomMDB(t, filename)
	db = openJetDB(filename, false)
	if db == nil {
 		t.Fatal("Ingen databas.")
	}
}

func TestPlatserTomMDB1(t *testing.T) {
	platserInit(t)

	// Denna testen
	antal := antalPlatser()
	
	if antal != 0 {
		t.Error("Antal platser != 0.")
	} else {
		t.Log("Antal platser ok.")
	}
	closeDB()
}

func TestPlatserMDB1(t *testing.T) {
	platserInit(t)

	// Denna testen
	skapaPlats("Platsnamnet", "12345-7", false, "")

	antal := antalPlatser()
	
	if antal != 1 {
		t.Error("Antal platser != 1.")
	} else {
		t.Log("Antal platser ok.")
	}
	closeDB()
}

func TestPlatserMDB2(t *testing.T) {
	platserInit(t)

	// Denna testen
	skapaPlats("Platsnamn1", "12345-7", false, "")
	skapaPlats("Platsnamn2", "", false, "")
	skapaPlats("Platsnamn3", "", false, "")
	skapaPlats("Platsnamn4", "12345-7", true, "")

	antal := antalPlatser()
	
	if antal != 4 {
		t.Error("Antal platser != 4.")
	} else {
		t.Log("Antal platser ok.")
	}
	closeDB()
}

func TestPlatserMDB3(t *testing.T) {
	platserInit(t)

	// Denna testen
	namn := "Tom € Räksmörgås"
	gironummer := "12345-7"
	kontokort := false
	skapaPlats(namn, gironummer, kontokort, "")

	plats := hämtaPlats(1)
	
	if plats.Namn != namn {
		t.Error("Platsnamn '"+namn+"' != '"+plats.Namn+"'.")
	} else {
		t.Log("Test namn ok.")
	}
	if plats.Gironummer != gironummer {
		t.Error("Plats gironummer '"+gironummer+"' != '"+plats.Gironummer+"'.")
	} else {
		t.Log("Test gironummer ok.")
	}
	if plats.Typ != kontokort {
		t.Error("Plats kontokort '"+strconv.FormatBool(kontokort)+"' != '"+strconv.FormatBool(plats.Typ)+"'.")
	} else {
		t.Log("Test kontokort ok.")
	}

	namn = "Tom2 € Räksmörgås"
	gironummer = " "
	kontokort = false
	skapaPlats(namn, gironummer, kontokort, "")

	plats = hämtaPlats(2)
	
	if plats.Namn != namn {
		t.Error("Platsnamn '"+namn+"' != '"+plats.Namn+"'.")
	} else {
		t.Log("Test namn ok.")
	}
	if plats.Gironummer != gironummer {
		t.Error("Plats gironummer '"+gironummer+"' != '"+plats.Gironummer+"'.")
	} else {
		t.Log("Test gironummer ok.")
	}
	if plats.Typ != kontokort {
		t.Error("Plats kontokort '"+strconv.FormatBool(kontokort)+"' != '"+strconv.FormatBool(plats.Typ)+"'.")
	} else {
		t.Log("Test kontokort ok.")
	}


	closeDB()
}

func TestPlatserMDB4(t *testing.T) {
	platserInit(t)

	// Denna testen
	namn := "Tom € Räksmörgås"
	gironummer := "12345-7"
	kontokort := false
	skapaPlats(namn, gironummer, kontokort, "")
	// TODO: skapaPlats(namn, gironummer, kontokort, "") // This should fail and report error due to duplicated name

	antal := antalPlatser()
	
	if antal != 1 {
		t.Error("Antal platser != 1.")
	} else {
		t.Log("Antal platser ok.")
	}

	closeDB()
}

// TODO: skapaplats borde validera gironummer.
// 123-4 felaktigt
// 123-0 ok
// Ett bankgironummer består av 7 eller 8 siffror och den sista siffran i numret är en kontrollsiffra som beräknas enligt Luhn-algoritmen, dvs enligt samma princip som beräkning av kontrollsiffran till svenska personnummer och organisationsnummer (som vi tagit upp i tidigare blogginlägg). Ett plusgironummer består av 2 till 8 siffror och även här består den sista siffran av en kontrollsiffra som beräknas enligt Luhn-algoritmen.

// TODO: Plats klassad som "Kontokortsföretag" ska ha Konto definierat

// TODO: Flaggan Kontokortsföretag ska inte påverka kolumnen Typ
