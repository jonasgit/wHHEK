//-*- coding: utf-8 -*-

package main

import (
	"testing"
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
