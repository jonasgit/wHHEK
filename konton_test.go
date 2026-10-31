package main

import (
	"strconv"
	"testing"

	"github.com/shopspring/decimal" // MIT License
)

func TestKontoTomDB1(t *testing.T) {
	t.Parallel()
	db := newTestDB(t, "kon1")

	// Denna testen
	antal := antalKonton(db)

	if antal != 1 {
		t.Error("Antal konton != (1).")
	} else {
		t.Log("Antal konton ok.")
	}
	konto := hämtaKonto(db, 1)

	if konto.Benämning != "Plånboken" {
		t.Error("Kontonamn '" + "Plånboken" + "' != '" + konto.Benämning + "'.")
	} else {
		t.Log("Test namn ok.")
	}
	startsaldo, err := decimal.NewFromString("0")
	if err != nil {
		t.Error(err)
	}
	if !konto.StartSaldo.Equal(startsaldo) {
		t.Error("Konto startsaldo '" + startsaldo.String() + "' != '" + konto.StartSaldo.String() + "'.")
	} else {
		t.Log("Test startsaldo ok.")
	}
	if !konto.Saldo.Equal(startsaldo) {
		t.Error("Konto saldo '" + startsaldo.String() + "' != '" + konto.Saldo.String() + "'.")
	} else {
		t.Log("Test saldo ok.")
	}
	if konto.StartManad != "Jan" {
		t.Error("Konto startmånad '" + "Jan" + "' != '" + konto.StartManad + "'.")
	} else {
		t.Log("Test startmånad ok.")
	}
	if konto.ArsskifteManad != "Jan" {
		t.Error("Konto skiftesmånad '" + "Jan" + "' != '" + konto.ArsskifteManad + "'.")
	} else {
		t.Log("Test skiftesmånad ok.")
	}
}

func TestKontoDB1(t *testing.T) {
	t.Parallel()
	db := newTestDB(t, "kon2")

	// Denna testen
	startsaldo, err := decimal.NewFromString("0.0")
	if err != nil {
		t.Error(err)
	}
	addKonto("Kontonamn1", startsaldo, "jan", db)

	antal := antalKonton(db)

	if antal != 2 {
		t.Error("Antal konton != (1+1).")
	} else {
		t.Log("Antal konton ok.")
	}
}

func TestKontoDB2(t *testing.T) {
	t.Parallel()
	db := newTestDB(t, "kon3")

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

	antal := antalKonton(db)

	if antal != 5 {
		t.Error("Antal konton != (1+4).")
	} else {
		t.Log("Antal konton ok.")
	}
}

func TestKontoDB3(t *testing.T) {
	t.Parallel()
	db := newTestDB(t, "kon4")

	// Denna testen
	namn := "Tom € Räksmörgås"
	startsaldo, err := decimal.NewFromString("19.99")
	if err != nil {
		t.Error(err)
	}

	startmånad := "Apr"
	addKonto(namn, startsaldo, startmånad, db)

	konto := hämtaKonto(db, 2)

	if konto.Benämning != namn {
		t.Error("Kontonamn '" + namn + "' != '" + konto.Benämning + "'.")
	} else {
		t.Log("Test namn ok.")
	}
	if !konto.StartSaldo.Equal(startsaldo) {
		t.Error("Konto startsaldo '" + startsaldo.String() + "' != '" + konto.StartSaldo.String() + "'.")
	} else {
		t.Log("Test startsaldo ok.")
	}
	if !konto.Saldo.Equal(startsaldo) {
		t.Error("Konto saldo '" + startsaldo.String() + "' != '" + konto.Saldo.String() + "'.")
	} else {
		t.Log("Test saldo ok.")
	}
	if konto.StartManad != startmånad {
		t.Error("Konto startmånad '" + startmånad + "' != '" + konto.StartManad + "'.")
	} else {
		t.Log("Test startmånad ok.")
	}
	if konto.ArsskifteManad != startmånad {
		t.Error("Konto skiftesmånad '" + startmånad + "' != '" + konto.ArsskifteManad + "'.")
	} else {
		t.Log("Test skiftesmånad ok.")
	}

}

func TestKontoDB4(t *testing.T) {
	t.Parallel()
	db := newTestDB(t, "kon5")

	// Denna testen
	namn := "Tom € Räksmörgås"
	startsaldo, err := decimal.NewFromString("19.99")
	if err != nil {
		t.Error(err)
	}

	startmånad := "Apr"
	addKonto(namn, startsaldo, startmånad, db)

	kontoid := hämtakontoID(db, "Plånboken")

	if kontoid != 1 {
		t.Error("Kontoid '" + "1" + "' != '" + strconv.Itoa(kontoid) + "'.")
	} else {
		t.Log("Test id 1 ok.")
	}

	kontoid = hämtakontoID(db, namn)

	if kontoid != 2 {
		t.Error("Kontoid '" + "1" + "' != '" + strconv.Itoa(kontoid) + "'.")
	} else {
		t.Log("Test id 1 ok.")
	}

}

// TODO: test startmånad = "foo" should fail
// TODO: test startsaldo = "Foo" should fail
// TODO: test namn = "😁" should fail? Due to Windows-1252 charset.
