package main

import (
	"testing"

	"github.com/shopspring/decimal"  // MIT License
)

func kontonInit(t *testing.T) {
	// Förberedelser
	var filename string = "gotestk.mdb"

	SkapaTomMDB(t, filename)
	db = openJetDB(filename, false)
}

func TestKontoMDB1(t *testing.T) {
	kontonInit(t)

	// Denna testen
	startsaldo, err := decimal.NewFromString("0.0")
	if err != nil {
		t.Error(err)
	}
	addKonto("Kontonamn1", startsaldo, "jan", db)

	antal := antalKonton()
	
	if antal != 2 {
		t.Error("Antal konton != (1+1).")
	} else {
		t.Log("Antal konton ok.")
	}
	closeDB()
}

func TestKontoMDB2(t *testing.T) {
	kontonInit(t)

	// Denna testen
	startsaldo, err := decimal.NewFromString("0.0")
	if err != nil {
		t.Error(err)
	}
	addKonto("Kontonamn1", startsaldo, "jan", db)
	startsaldo, err = decimal.NewFromString("1000000.0")
	if err != nil {
		t.Error(err)
	}
	addKonto("Kontonamn2", startsaldo, "jan", db)
	startsaldo, err = decimal.NewFromString("0.0")
	if err != nil {
		t.Error(err)
	}
	addKonto("Kontonamn3", startsaldo, "jul", db)
	addKonto("Kontonamn4", startsaldo, "dec", db)

	antal := antalKonton()
	
	if antal != 5 {
		t.Error("Antal konton != (1+4).")
	} else {
		t.Log("Antal konton ok.")
	}
	closeDB()
}

func TestKontoMDB3(t *testing.T) {
	kontonInit(t)

	// Denna testen
	namn := "Tom € Räksmörgås"
	startsaldo, err := decimal.NewFromString("19.99")
	if err != nil {
		t.Error(err)
	}
	
	startmånad := "Apr"
	addKonto(namn, startsaldo, startmånad, db)

	konto := hämtaKonto(2)
	
	if konto.Benämning != namn {
		t.Error("Kontonamn '"+namn+"' != '"+konto.Benämning+"'.")
	} else {
		t.Log("Test namn ok.")
	}
	if !konto.StartSaldo.Equal(startsaldo) {
		t.Error("Konto startsaldo '"+startsaldo.String()+"' != '"+konto.StartSaldo.String()+"'.")
	} else {
		t.Log("Test startsaldo ok.")
	}
	if !konto.Saldo.Equal(startsaldo) {
		t.Error("Konto saldo '"+startsaldo.String()+"' != '"+konto.Saldo.String()+"'.")
	} else {
		t.Log("Test saldo ok.")
	}
	if konto.StartManad != startmånad {
		t.Error("Konto startmånad '"+startmånad+"' != '"+konto.StartManad+"'.")
	} else {
		t.Log("Test startmånad ok.")
	}

	closeDB()
}

// TODO: test startmånad = "foo" should fail
// TODO: test startsaldo = "Foo" should fail
// TODO: test namn = "😁" should fail? Due to Windows-1252 charset.
