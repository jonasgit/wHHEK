//-*- coding: utf-8 -*-

package main

import (
	"testing"
	"time"

	"github.com/shopspring/decimal" // MIT License
)

func TestFastTransaktionTomDB1(t *testing.T) {
	t.Parallel()
	db := newTestDB(t, "ftrt1")

	// Denna testen
	antal := antalFastaTransaktioner(db)

	if antal != 0 {
		t.Error("Antal fasta transaktioner != (0).")
	} else {
		t.Log("Antal fasta transaktioner ok (0).")
	}

}

func compareDates(t *testing.T, t1 time.Time, t2 time.Time) bool {
	/*	t.Log("t1.Year: ", t1.Year())
		t.Log("t1.Month: ", int(t1.Month()))
		t.Log("t1.Day: ", t1.Day())
		t.Log("t2.Year: ", t2.Year())
		t.Log("t2.Month: ", int(t2.Month()))
		t.Log("t2.Day: ", t2.Day())*/

	if t1.Year() != t2.Year() {
		//t.Log("False")
		return false
	}
	if int(t1.Month()) != int(t2.Month()) {
		//t.Log("False")
		return false
	}
	if t1.Day() != t2.Day() {
		//t.Log("False")
		return false
	}
	t.Log("True")
	return true
}

func TestFastTransaktionTomDB2(t *testing.T) {
	t.Parallel()
	// Förbered test
	db := newTestDB(t, "ftrt2")
	_ = skapaPlats(db, "Mack", "", false, "")

	antal := antalTransaktioner(db)
	if antal != 0 {
		t.Error("Antal transaktioner != (0).")
	} else {
		t.Log("Antal transaktioner ok (0).")
	}

	// Denna testen: Fast utgift, månadsvis, testa datum slut på månad
	// Skapa ny fast transaktion, datum 2021-01-31, månadsvis
	summa := decimal.NewFromInt(1)

	_ = skapaFastUtgift(db, "Bil", "Plånboken", "Gemensamt", "Mack", summa, "2021-01-31", "Varje månad")

	antal = antalFastaTransaktioner(db)
	if antal != 1 {
		t.Error("Antal fasta transaktioner != (1).")
	} else {
		t.Log("Antal fasta transaktioner ok (1).")
	}

	// Registrera denna 1 gång
	t.Log("pre-reg 1.")
	registreraFastTransaktion(db, 1)
	t.Log("post-reg 1.")
	antal = antalFastaTransaktioner(db)
	if antal != 1 {
		t.Error("Antal fasta transaktioner != (1).")
	} else {
		t.Log("Antal fasta transaktioner ok (1).")
	}
	antal = antalTransaktioner(db)
	if antal != 1 {
		t.Error("Antal transaktioner != (1).")
	} else {
		t.Log("Antal transaktioner ok (1).")
	}
	// Kontrollera resultatet, nästa datum = 2021-02-28
	ft := hämtaFastTransaktion(db, 1)
	expectedDate, _ := time.Parse("2006-01-02", "2021-02-28")
	if compareDates(t, ft.date, expectedDate) {
		t.Log("Förväntat datum ok (2021-02-28).")
	} else {
		t.Error("Förväntat datum 2021-02-28 != ", ft.date)
	}
	// Registrera denna 1 gång
	registreraFastTransaktion(db, 1)
	antal = antalTransaktioner(db)
	if antal != 2 {
		t.Error("Antal transaktioner != (2).")
	} else {
		t.Log("Antal transaktioner ok (2).")
	}
	// Kontrollera resultatet, nästa datum = 2021-03-28
	// Registrera denna 1 gång
	registreraFastTransaktion(db, 1)
	antal = antalTransaktioner(db)
	// Kontrollera resultatet, datum = 2021-03-28, saldo -3kr
	if antal != 3 {
		t.Error("Antal transaktioner != (3).")
	} else {
		t.Log("Antal transaktioner ok (3).")
	}
	expectedDate, _ = time.Parse("2006-01-02", "2021-04-28")
	ft = hämtaFastTransaktion(db, 1)
	if compareDates(t, ft.date, expectedDate) {
		t.Log("Förväntat datum ok (2021-04-28).")
	} else {
		t.Error("Förväntat datum 2021-04-28 != ", ft.date)
	}

	currentTime := time.Now()
	currDate := currentTime.Format("2006-01-02")
	_, totSaldo := saldonKonto(db, "Plånboken", currDate)
	expSaldo := decimal.NewFromInt(-3)
	if totSaldo.Equal(expSaldo) {
		t.Log("Förväntat saldo ok (-3,00).")
	} else {
		t.Error("Förväntat saldo -3,00 != ", totSaldo, expSaldo)
	}

}

func TestFastTransaktionTomDB3(t *testing.T) {
	t.Parallel()
	// Förbered test
	db := newTestDB(t, "ftrt3")
	_ = skapaPlats(db, "Mack", "", false, "")

	antal := antalTransaktioner(db)
	if antal != 0 {
		t.Error("Antal transaktioner != (0).")
	} else {
		t.Log("Antal transaktioner ok (0).")
	}

	// Denna testen: Fast utgift, kvartalsvis, testa datum slut på månad

	// Skapa ny fast transaktion, datum 2020-11-30, kvartalsvis
	summa := decimal.NewFromInt(1)

	_ = skapaFastUtgift(db, "Bil", "Plånboken", "Gemensamt", "Mack", summa, "2020-11-30", "Varje kvartal")

	antal = antalFastaTransaktioner(db)
	if antal != 1 {
		t.Error("Antal fasta transaktioner != (1).")
	} else {
		t.Log("Antal fasta transaktioner ok (1).")
	}

	// Registrera denna 1 gång
	registreraFastTransaktion(db, 1)
	antal = antalFastaTransaktioner(db)
	if antal != 1 {
		t.Error("Antal fasta transaktioner != (1).")
	} else {
		t.Log("Antal fasta transaktioner ok (1).")
	}
	antal = antalTransaktioner(db)
	if antal != 1 {
		t.Error("Antal transaktioner != (1).")
	} else {
		t.Log("Antal transaktioner ok (1).")
	}
	// Kontrollera resultatet, nästa datum = 2021-02-28
	ft := hämtaFastTransaktion(db, 1)
	expectedDate, _ := time.Parse("2006-01-02", "2021-02-28")
	if compareDates(t, ft.date, expectedDate) {
		t.Log("Förväntat datum ok (2021-02-28).")
	} else {
		t.Error("Förväntat datum 2021-02-28 != ", ft.date)
	}
	// Registrera denna 1 gång
	registreraFastTransaktion(db, 1)
	antal = antalTransaktioner(db)
	if antal != 2 {
		t.Error("Antal transaktioner != (2).")
	} else {
		t.Log("Antal transaktioner ok (2).")
	}
	// Kontrollera resultatet, nästa datum = 2021-05-28
	// Registrera denna 1 gång
	registreraFastTransaktion(db, 1)
	antal = antalTransaktioner(db)
	// Kontrollera resultatet, datum = 2021-05-28, saldo -3kr
	if antal != 3 {
		t.Error("Antal transaktioner != (3).")
	} else {
		t.Log("Antal transaktioner ok (3).")
	}
	expectedDate, _ = time.Parse("2006-01-02", "2021-08-28")
	ft = hämtaFastTransaktion(db, 1)
	if compareDates(t, ft.date, expectedDate) {
		t.Log("Förväntat datum ok (2021-08-28).")
	} else {
		t.Error("Förväntat datum 2021-08-28 != ", ft.date)
	}

	currentTime := time.Now()
	currDate := currentTime.Format("2006-01-02")
	_, totSaldo := saldonKonto(db, "Plånboken", currDate)
	expSaldo := decimal.NewFromInt(-3)
	if totSaldo.Equal(expSaldo) {
		t.Log("Förväntat saldo ok (-3,00).")
	} else {
		t.Error("Förväntat saldo -3,00 != ", totSaldo, expSaldo)
	}

}

func TestNextFixedDate(t *testing.T) {
	t.Parallel()
	for _, c := range []struct{ datum, hurofta, want string }{
		{"2021-01-31", "Varje månad", "2021-02-28"},
		{"2021-1-31", "varje månad", "2021-02-28"},
		{" 2021-01-15 ", " VARANNAN VECKA ", "2021-01-29"},
		{"2020-11-30", "Varje kvartal", "2021-02-28"},
		{"2021-01-15", "12", "2022-01-15"},
		{"2021-01-31", "Månadsvis", ""},
		{"31/1 2021", "Varje månad", ""},
		{"", "Varje månad", ""},
	} {
		got, err := nextFixedDate(c.datum, c.hurofta)
		if c.want == "" {
			if err == nil {
				t.Errorf("nextFixedDate(%q, %q) = %q, förväntade fel", c.datum, c.hurofta, got)
			}
		} else if err != nil || got != c.want {
			t.Errorf("nextFixedDate(%q, %q) = %q, %v; förväntade %q", c.datum, c.hurofta, got, err, c.want)
		}
	}
}

func TestFastTransaktionOkändPeriodRegistrerasInte(t *testing.T) {
	t.Parallel()
	db := newTestDB(t, "ftrt4")
	_ = skapaPlats(db, "Mack", "", false, "")
	_ = skapaFastUtgift(db, "Bil", "Plånboken", "Gemensamt", "Mack", decimal.NewFromInt(1), "2021-01-31", "Varje månad")
	// Äldre data eller annan källa: okänd periodicitet i databasen.
	if _, err := db.Exec("UPDATE Överföringar SET HurOfta = 'Månadsvis' WHERE Löpnr = 1"); err != nil {
		t.Fatal(err)
	}

	if err := registreraFastTransaktion(db, 1); err == nil {
		t.Error("förväntade fel för okänd periodicitet")
	}
	if antal := antalTransaktioner(db); antal != 0 {
		t.Errorf("ingen transaktion ska registreras, fick %d", antal)
	}
	if ft := hämtaFastTransaktion(db, 1); !compareDates(t, ft.date, time.Date(2021, 1, 31, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("datumet ska inte flyttas fram, fick %v", ft.date)
	}
}

func TestFastTransaktionGammaltDatumformat(t *testing.T) {
	t.Parallel()
	db := newTestDB(t, "ftrt5")
	_ = skapaPlats(db, "Mack", "", false, "")
	_ = skapaFastUtgift(db, "Bil", "Plånboken", "Gemensamt", "Mack", decimal.NewFromInt(1), "2021-1-31", "varje månad")

	if err := registreraFastTransaktion(db, 1); err != nil {
		t.Fatal(err)
	}
	if tr := hämtaTransaktion(db, 1); tr.date.Format("2006-01-02") != "2021-01-31" {
		t.Errorf("transaktionen ska få datumet 2021-01-31, fick %q", tr.date)
	}
	if ft := hämtaFastTransaktion(db, 1); !compareDates(t, ft.date, time.Date(2021, 2, 28, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("nästa datum ska vara 2021-02-28, fick %v", ft.date)
	}
}

func TestUpdateFastUtgNormaliserar(t *testing.T) {
	t.Parallel()
	db := newTestDB(t, "ftrt6")
	_ = skapaPlats(db, "Mack", "", false, "")
	_ = skapaFastUtgift(db, "Bil", "Plånboken", "Gemensamt", "Mack", decimal.NewFromInt(1), "2021-01-31", "Varje månad")

	if err := updateFastUtgSQL(1, db, "Plånboken", "Mack", "2021-3-5", "varje kvartal", "", "Bil", "Gemensamt", "1", ""); err != nil {
		t.Fatal(err)
	}
	var datum, hurofta string
	_ = db.QueryRow("SELECT Datum, HurOfta FROM Överföringar WHERE Löpnr = 1").Scan(&datum, &hurofta)
	if datum != "2021-03-05" || hurofta != "Varje kvartal" {
		t.Errorf("förväntade 2021-03-05 / Varje kvartal, fick %q / %q", datum, hurofta)
	}

	if err := updateFastUtgSQL(1, db, "Plånboken", "Mack", "2021-03-05", "Månadsvis", "", "Bil", "Gemensamt", "1", ""); err == nil {
		t.Error("förväntade fel för okänd periodicitet")
	}
	_ = db.QueryRow("SELECT HurOfta FROM Överföringar WHERE Löpnr = 1").Scan(&hurofta)
	if hurofta != "Varje kvartal" {
		t.Errorf("ogiltig ändring ska inte sparas, fick %q", hurofta)
	}
}
