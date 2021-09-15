//-*- coding: utf-8 -*-

package main

import (
	"testing"
	"strconv"
)

func personerInit(t *testing.T) {
	// Förberedelser
	var filename string = "gotestp.mdb"

	SkapaTomMDB(t, filename)
	db = openJetDB(filename, false)
	if db == nil {
 		t.Fatal("Ingen databas.")
	}
}

func TestPersonTomMDB1(t *testing.T) {
	personerInit(t)

	// Denna testen
	antal := antalPersoner()
	
	if antal != 1 {
		t.Error("Antal personer != 1.")
	} else {
		t.Log("Antal personer ok.")
	}
	namn := "Gemensamt"
	birth := 0
	sex := "Gemensamt"
	person := hämtaPerson(1)
	
	if person.namn != namn {
		t.Error("Personnamn '"+namn+"' != '"+person.namn+"'.")
	} else {
		t.Log("Test namn ok.")
	}
	if person.birth != birth {
		t.Error("Person född '"+strconv.Itoa(birth)+"' != '"+strconv.Itoa(person.birth)+"'.")
	} else {
		t.Log("Test född ok.")
	}
	if person.sex != sex {
		t.Error("Person kön '"+sex+"' != '"+person.sex+"'.")
	} else {
		t.Log("Test kön ok.")
	}
	closeDB()
}

func TestPersonMDB1(t *testing.T) {
	personerInit(t)

	// Denna testen
	skapaPerson("Namn Person", 1994, "Man")

	antal := antalPersoner()
	
	if antal != 2 {
		t.Error("Antal personer != (1+1).")
	} else {
		t.Log("Antal personer ok.")
	}
	closeDB()
}

func TestPersonMDB2(t *testing.T) {
	personerInit(t)

	// Denna testen
	skapaPerson("Namn Person", 1994, "Man")
	skapaPerson("Namn Person", 1996, "Kvinna")
	skapaPerson("Namn Person", 2004, "Man")
	skapaPerson("Namn Person", 2006, "Kvinna")

	antal := antalPersoner()
	
	if antal != 5 {
		t.Error("Antal personer != (1+4).")
	} else {
		t.Log("Antal personer ok.")
	}
	closeDB()
}

func TestPersonMDB3(t *testing.T) {
	personerInit(t)

	// Denna testen
	namn := "Tom € Räksmörgås"
	birth := 1999
	sex := "Man"
	skapaPerson(namn, birth, sex)

	person := hämtaPerson(2)
	
	if person.namn != namn {
		t.Error("Personnamn '"+namn+"' != '"+person.namn+"'.")
	} else {
		t.Log("Test namn ok.")
	}
	if person.birth != birth {
		t.Error("Person född '"+strconv.Itoa(birth)+"' != '"+strconv.Itoa(person.birth)+"'.")
	} else {
		t.Log("Test född ok.")
	}
	if person.sex != sex {
		t.Error("Person kön '"+sex+"' != '"+person.sex+"'.")
	} else {
		t.Log("Test kön ok.")
	}


	namn  = "** \"\" ');  **"  // Note: ' ej tillåtet enligt HH
	birth = 2000
	sex   = "Kvinna"
	skapaPerson(namn, birth, sex)

	person = hämtaPerson(3)
	
	if unEscapeSQL(person.namn) != namn {
		t.Error("Personnamn '"+namn+"' != '"+unEscapeSQL(person.namn)+"'.")
	} else {
		t.Log("Test namn ok.")
	}
	if person.birth != birth {
		t.Error("Person född '"+strconv.Itoa(birth)+"' != '"+strconv.Itoa(person.birth)+"'.")
	} else {
		t.Log("Test född ok.")
	}
	if person.sex != sex {
		t.Error("Person kön '"+sex+"' != '"+person.sex+"'.")
	} else {
		t.Log("Test kön ok.")
	}
	closeDB()
}

// TODO: test birth < 1900 should fail
// TODO: test birth > 2200 should fail
// TODO: test birth = -1 should fail
