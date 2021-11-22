//-*- coding: utf-8 -*-

package main

import (
	"database/sql"
	"testing"
	"strconv"
)

func personerInit(t *testing.T, filnamn string) *sql.DB {
	// Förberedelser
 	if JetDBSupport {
	      	var filename string = "got"+filnamn+".mdb"
	        t.Log("Jet Supported.")

		SkapaTomMDB(t, filename)
		db = openJetDB(filename, false)
	} else {
	      	var filename string = "got"+filnamn+".db"
	        t.Log("Jet NOT Supported.")
		SkapaTomDB(filename)
		db = openSqlite(filename)
	}

	if db == nil {
 		t.Fatal("Ingen databas.")
	}
	return db
}

func TestPersonTomDB1(t *testing.T) {
	db = personerInit(t, "prs1")
	
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

func TestPersonDB1(t *testing.T) {
	db = personerInit(t, "prs2")
	
	// Denna testen
	skapaPerson(db, "Namn Person", 1994, "Man")
	
	antal := antalPersoner(db)
	
	if antal != 2 {
		t.Error("Antal personer != (1+1).")
	} else {
		t.Log("Antal personer ok.")
	}
	closeDB()
}

func TestPersonDB2(t *testing.T) {
	db = personerInit(t, "prs3")
	
	// Denna testen
	skapaPerson(db, "Namn Person", 1994, "Man")
	skapaPerson(db, "Namn Person", 1996, "Kvinna")
	skapaPerson(db, "Namn Person", 2004, "Man")
	skapaPerson(db, "Namn Person", 2006, "Kvinna")
	
	antal := antalPersoner(db)
	
	if antal != 5 {
		t.Error("Antal personer != (1+4).")
	} else {
		t.Log("Antal personer ok.")
	}
	closeDB()
}

func TestPersonDB3(t *testing.T) {
	db = personerInit(t, "prs4")
	
	// Denna testen
	namn := "Tom € Räksmörgås"
	birth := 1999
	sex := "Man"
	skapaPerson(db, namn, birth, sex)
	
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
	skapaPerson(db, namn, birth, sex)
	
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
