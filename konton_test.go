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

func TestKontoUppdateraDecimal(t *testing.T) {
	t.Parallel()
	db := newTestDB(t, "kon6")
	want := decimal.RequireFromString("123.45")
	for _, in := range []string{"123.45", "123,45"} {
		updateKonto(1, "Plånboken", in, "Jan", db)
		if k := hämtaKonto(db, 1); !k.StartSaldo.Equal(want) {
			t.Errorf("startsaldo %q: förväntade %s, fick %s", in, want, k.StartSaldo)
		}
		if s := saldoKonto(db, "Plånboken", ""); !s.Equal(want) {
			t.Errorf("saldo efter startsaldo %q: förväntade %s, fick %s", in, want, s)
		}
	}
}

func TestKontoNyttSvensktBelopp(t *testing.T) {
	t.Parallel()
	db := newTestDB(t, "kon7")
	addKontow("Sparkonto", "1 234,50", "1", db)
	k := hämtaKonto(db, hämtakontoID(db, "Sparkonto"))
	if want := decimal.RequireFromString("1234.50"); !k.StartSaldo.Equal(want) {
		t.Errorf("förväntade startsaldo %s, fick %s", want, k.StartSaldo)
	}
}

func TestKontoLäserGammaltKommatecken(t *testing.T) {
	t.Parallel()
	if JetDBSupport && dbdecimaldot {
		t.Skip("Med decimalpunkt i Windows tolkar Access '123,45' som 12345. Gamla kommatecken fanns bara i SQLite.")
	}
	db := newTestDB(t, "kon8")
	// Så sparade updateKonto tidigare startsaldo i SQLite.
	if _, err := db.Exec("UPDATE Konton SET StartSaldo = '123,45' WHERE Löpnr = 1"); err != nil {
		t.Fatal(err)
	}
	want := decimal.RequireFromString("123.45")
	if k := hämtaKonto(db, 1); !k.StartSaldo.Equal(want) {
		t.Errorf("förväntade %s, fick %s", want, k.StartSaldo)
	}
	if s := saldoKonto(db, "Plånboken", ""); !s.Equal(want) {
		t.Errorf("förväntade saldo %s, fick %s", want, s)
	}
}
