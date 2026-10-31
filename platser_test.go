//-*- coding: utf-8 -*-

package main

import (
	"strconv"
	"testing"
)

func TestPlatserTomDB1(t *testing.T) {
	t.Parallel()
	db := newTestDB(t, "pl1")

	// Denna testen
	antal := antalPlatser(db)

	if antal != 0 {
		t.Error("Antal platser != 0.")
	} else {
		t.Log("Antal platser ok.")
	}
}

func TestPlatserDB1(t *testing.T) {
	t.Parallel()
	db := newTestDB(t, "pl2")

	// Denna testen
	_ = skapaPlats(db, "Platsnamnet", "12345-7", false, "")

	antal := antalPlatser(db)

	if antal != 1 {
		t.Error("Antal platser != 1.")
	} else {
		t.Log("Antal platser ok.")
	}
}

func TestPlatserDB2(t *testing.T) {
	t.Parallel()
	db := newTestDB(t, "pl3")

	// Denna testen
	_ = skapaPlats(db, "Platsnamn1", "12345-7", false, "")
	_ = skapaPlats(db, "Platsnamn2", "", false, "")
	_ = skapaPlats(db, "Platsnamn3", "", false, "")
	_ = skapaPlats(db, "Platsnamn4", "12345-7", true, "")

	antal := antalPlatser(db)

	if antal != 4 {
		t.Error("Antal platser != 4.")
	} else {
		t.Log("Antal platser ok.")
	}
}

func TestPlatserDB3(t *testing.T) {
	t.Parallel()
	db := newTestDB(t, "pl4")

	// Denna testen
	namn := "Tom € Räksmörgås"
	gironummer := "12345-7"
	kontokort := false
	_ = skapaPlats(db, namn, gironummer, kontokort, "")

	plats := hämtaPlats(db, 1)

	if plats.Namn != namn {
		t.Error("Platsnamn '" + namn + "' != '" + plats.Namn + "'.")
	} else {
		t.Log("Test namn ok.")
	}
	if plats.Gironummer != gironummer {
		t.Error("Plats gironummer '" + gironummer + "' != '" + plats.Gironummer + "'.")
	} else {
		t.Log("Test gironummer ok.")
	}
	if plats.Typ != kontokort {
		t.Error("Plats kontokort '" + strconv.FormatBool(kontokort) + "' != '" + strconv.FormatBool(plats.Typ) + "'.")
	} else {
		t.Log("Test kontokort ok.")
	}

	namn = "Tom2 € Räksmörgås"
	gironummer = " "
	kontokort = false
	_ = skapaPlats(db, namn, gironummer, kontokort, "")

	plats = hämtaPlats(db, 2)

	if plats.Namn != namn {
		t.Error("Platsnamn '" + namn + "' != '" + plats.Namn + "'.")
	} else {
		t.Log("Test namn ok.")
	}
	if plats.Gironummer != gironummer {
		t.Error("Plats gironummer '" + gironummer + "' != '" + plats.Gironummer + "'.")
	} else {
		t.Log("Test gironummer ok.")
	}
	if plats.Typ != kontokort {
		t.Error("Plats kontokort '" + strconv.FormatBool(kontokort) + "' != '" + strconv.FormatBool(plats.Typ) + "'.")
	} else {
		t.Log("Test kontokort ok.")
	}

}

func TestPlatserDB4(t *testing.T) {
	t.Parallel()
	db := newTestDB(t, "pl5")

	// Denna testen
	namn := "Tom € Räksmörgås"
	gironummer := "12345-7"
	kontokort := false
	_ = skapaPlats(db, namn, gironummer, kontokort, "")
	// TODO: skapaPlats(namn, gironummer, kontokort, "") // This should fail and report error due to duplicated name

	antal := antalPlatser(db)

	if antal != 1 {
		t.Error("Antal platser != 1.")
	} else {
		t.Log("Antal platser ok.")
	}

}

// TODO: skapaplats borde validera gironummer.
// 123-4 felaktigt
// 123-0 ok
// Ett bankgironummer består av 7 eller 8 siffror och den sista siffran i numret är en kontrollsiffra som beräknas enligt Luhn-algoritmen, dvs enligt samma princip som beräkning av kontrollsiffran till svenska personnummer och organisationsnummer (som vi tagit upp i tidigare blogginlägg). Ett plusgironummer består av 2 till 8 siffror och även här består den sista siffran av en kontrollsiffra som beräknas enligt Luhn-algoritmen.

// TODO: Plats klassad som "Kontokortsföretag" ska ha Konto definierat

// TODO: Flaggan Kontokortsföretag ska inte påverka kolumnen Typ
