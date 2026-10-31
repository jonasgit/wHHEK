//-*- coding: utf-8 -*-

package main

import (
	"strconv"
	"testing"
)

func TestPersonTomDB1(t *testing.T) {
	t.Parallel()
	db := newTestDB(t, "prs1")

	// Denna testen
	antal := antalPersoner(db)

	if antal != 1 {
		t.Error("Antal personer != 1.")
	} else {
		t.Log("Antal personer ok.")
	}
	namn := "Gemensamt"
	birth := 0
	sex := "Gemensamt"
	person := hämtaPerson(db, 1)

	if person.namn != namn {
		t.Error("Personnamn '" + namn + "' != '" + person.namn + "'.")
	} else {
		t.Log("Test namn ok.")
	}
	if person.birth != birth {
		t.Error("Person född '" + strconv.Itoa(birth) + "' != '" + strconv.Itoa(person.birth) + "'.")
	} else {
		t.Log("Test född ok.")
	}
	if person.sex != sex {
		t.Error("Person kön '" + sex + "' != '" + person.sex + "'.")
	} else {
		t.Log("Test kön ok.")
	}
}

func TestPersonDB1(t *testing.T) {
	t.Parallel()
	db := newTestDB(t, "prs2")

	// Denna testen
	_ = skapaPerson(db, "Namn Person", 1994, "Man")

	antal := antalPersoner(db)

	if antal != 2 {
		t.Error("Antal personer != (1+1).")
	} else {
		t.Log("Antal personer ok.")
	}
}

func TestPersonDB2(t *testing.T) {
	t.Parallel()
	db := newTestDB(t, "prs3")

	// Denna testen
	_ = skapaPerson(db, "Namn Person", 1994, "Man")
	_ = skapaPerson(db, "Namn Person", 1996, "Kvinna")
	_ = skapaPerson(db, "Namn Person", 2004, "Man")
	_ = skapaPerson(db, "Namn Person", 2006, "Kvinna")

	antal := antalPersoner(db)

	if antal != 5 {
		t.Error("Antal personer != (1+4).")
	} else {
		t.Log("Antal personer ok.")
	}
}

func TestPersonDB3(t *testing.T) {
	t.Parallel()
	db := newTestDB(t, "prs4")

	// Denna testen
	namn := "Tom € Räksmörgås"
	birth := 1999
	sex := "Man"
	_ = skapaPerson(db, namn, birth, sex)

	person := hämtaPerson(db, 2)

	if person.namn != namn {
		t.Error("Personnamn '" + namn + "' != '" + person.namn + "'.")
	} else {
		t.Log("Test namn ok.")
	}
	if person.birth != birth {
		t.Error("Person född '" + strconv.Itoa(birth) + "' != '" + strconv.Itoa(person.birth) + "'.")
	} else {
		t.Log("Test född ok.")
	}
	if person.sex != sex {
		t.Error("Person kön '" + sex + "' != '" + person.sex + "'.")
	} else {
		t.Log("Test kön ok.")
	}

	namn = "** \"\" ');  **" // Note: ' ej tillåtet enligt HH
	birth = 2000
	sex = "Kvinna"
	_ = skapaPerson(db, namn, birth, sex)

	person = hämtaPerson(db, 3)

	if person.namn != namn {
		t.Error("Personnamn '" + namn + "' != '" + person.namn + "'.")
	} else {
		t.Log("Test namn ok.")
	}
	if person.birth != birth {
		t.Error("Person född '" + strconv.Itoa(birth) + "' != '" + strconv.Itoa(person.birth) + "'.")
	} else {
		t.Log("Test född ok.")
	}
	if person.sex != sex {
		t.Error("Person kön '" + sex + "' != '" + person.sex + "'.")
	} else {
		t.Log("Test kön ok.")
	}
}

// TODO: test birth < 1900 should fail
// TODO: test birth > 2200 should fail
// TODO: test birth = -1 should fail
