//-*- coding: utf-8 -*-

package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/shopspring/decimal" // MIT License
)

// Fältlängder i tecken, enligt kolumnerna i Hogias mdb-fil (TOMDB.MDB).
const (
	maxLenKonto        = 40 // Konton.Benämning, Transaktioner.FrånKonto/TillKonto
	maxLenTyp          = 40 // Transaktioner.Typ, Budget.Typ
	maxLenVad          = 40 // Transaktioner.Vad, Överföringar.Vad
	maxLenVem          = 50 // Transaktioner.Vem, Personer.Namn
	maxLenVemFast      = 40 // Överföringar.Vem
	maxLenText         = 60 // Transaktioner.Text
	maxLenHurOfta      = 15 // Överföringar.HurOfta
	maxLenRakning      = 1  // Överföringar.Rakning
	maxLenManad        = 10 // Konton.StartManad, Budget.StartMånad
	maxLenPlatsNamn    = 40 // Platser.Namn
	maxLenGironummer   = 20 // Platser.Gironummer
	maxLenPlatsTyp     = 2  // Platser.Typ
	maxLenRefKonto     = 40 // Platser.RefKonto
	maxLenKon          = 10 // Personer.Kön
	maxLenLosenord     = 8  // DtbVer.Losenord
	maxLenFilnamn      = 200
	maxLenChatQuestion = 2000
	maxLenChatHistory  = 1 << 20
)

// Belopp sparas som DECIMAL(19,4).
const (
	maxAmountIntDigits = 15
	maxAmountDecimals  = 4
)

// formCheck läser och kontrollerar fält från ett formulär. Felen samlas så att
// alla kan visas på en gång.
type formCheck struct {
	req  *http.Request
	errs []string
}

func newFormCheck(req *http.Request) *formCheck {
	return &formCheck{req: req}
}

func (c *formCheck) fail(format string, args ...any) {
	c.errs = append(c.errs, fmt.Sprintf(format, args...))
}

// Err returnerar nil om alla fält var giltiga.
func (c *formCheck) Err() error {
	if len(c.errs) == 0 {
		return nil
	}
	return errors.New(strings.Join(c.errs, "; "))
}

// checkText kontrollerar längd och tecken i en enradig text.
func checkText(value string, maxLen int) error {
	if !utf8.ValidString(value) {
		return fmt.Errorf("innehåller ogiltiga tecken")
	}
	for _, r := range value {
		if unicode.IsControl(r) {
			return fmt.Errorf("innehåller styrtecken")
		}
	}
	if n := utf8.RuneCountInString(value); n > maxLen {
		return fmt.Errorf("är för lång (%d tecken, högst %d)", n, maxLen)
	}
	return nil
}

// Text läser ett valfritt textfält. Inledande och avslutande blanksteg tas bort.
func (c *formCheck) Text(name, label string, maxLen int) string {
	value := strings.TrimSpace(c.req.FormValue(name))
	if err := checkText(value, maxLen); err != nil {
		c.fail("%s %v", label, err)
		return ""
	}
	return value
}

// RequiredText läser ett textfält som inte får vara tomt.
func (c *formCheck) RequiredText(name, label string, maxLen int) string {
	value := c.Text(name, label, maxLen)
	if value == "" && strings.TrimSpace(c.req.FormValue(name)) == "" {
		c.fail("%s saknas", label)
	}
	return value
}

// OneOf läser ett fält som måste ha ett av de tillåtna värdena. Ett tomt fält
// är tillåtet bara om "" finns bland värdena.
func (c *formCheck) OneOf(name, label string, allowed ...string) string {
	value := strings.TrimSpace(c.req.FormValue(name))
	for _, a := range allowed {
		if value == a {
			return value
		}
	}
	if value == "" {
		c.fail("%s saknas", label)
	} else {
		c.fail("%s har ett ogiltigt värde", label)
	}
	return ""
}

// parseDate tolkar ett datum ÅÅÅÅ-MM-DD (även ÅÅÅÅ-M-D) mellan år 1900 och 2199.
func parseDate(value string) (time.Time, error) {
	t, err := parseFixedDate(value)
	if err != nil {
		return t, err
	}
	if t.Year() < 1900 || t.Year() > 2199 {
		return t, fmt.Errorf("datum '%s' är utanför 1900–2199", value)
	}
	return t, nil
}

// Date läser ett datum och returnerar det som ÅÅÅÅ-MM-DD. Är fältet tomt och inte
// obligatoriskt returneras "".
func (c *formCheck) Date(name, label string, required bool) string {
	value := strings.TrimSpace(c.req.FormValue(name))
	if value == "" {
		if required {
			c.fail("%s saknas", label)
		}
		return ""
	}
	t, err := parseDate(value)
	if err != nil {
		c.fail("%s: ogiltigt datum '%s' (ska vara ÅÅÅÅ-MM-DD)", label, value)
		return ""
	}
	return t.Format("2006-01-02")
}

// parseAmount tolkar ett belopp som användaren skrivit, t.ex. "1 234,50", "-12.5"
// eller "+3". Blanksteg mellan siffergrupper är tillåtna. Annat än siffror, ett
// decimaltecken och ett inledande tecken ger fel, till skillnad från SanitizeAmount
// som tyst tar bort tecken.
func parseAmount(value string) (decimal.Decimal, error) {
	s := strings.TrimSpace(value)
	s = strings.NewReplacer(" ", "", " ", "", " ", "").Replace(s)
	if s == "" {
		return decimal.Zero, fmt.Errorf("belopp saknas")
	}
	digits := s
	if digits[0] == '-' || digits[0] == '+' {
		digits = digits[1:]
	}
	digits = strings.Replace(digits, ",", ".", 1)
	intPart, decPart, hasDec := strings.Cut(digits, ".")
	isDigits := func(p string) bool {
		for _, r := range p {
			if r < '0' || r > '9' {
				return false
			}
		}
		return true
	}
	if intPart == "" && decPart == "" || !isDigits(intPart) || !isDigits(decPart) || (hasDec && decPart == "") {
		return decimal.Zero, fmt.Errorf("'%s' är inte ett belopp", value)
	}
	if len(strings.TrimLeft(intPart, "0")) > maxAmountIntDigits {
		return decimal.Zero, fmt.Errorf("'%s' är för stort", value)
	}
	if len(decPart) > maxAmountDecimals {
		return decimal.Zero, fmt.Errorf("'%s' har fler än %d decimaler", value, maxAmountDecimals)
	}
	if intPart == "" {
		intPart = "0"
	}
	normalized := intPart
	if decPart != "" {
		normalized += "." + decPart
	}
	if s[0] == '-' {
		normalized = "-" + normalized
	}
	return decimal.NewFromString(normalized)
}

// Amount läser ett belopp. Ett tomt fält som inte är obligatoriskt blir 0 och
// ok=false.
func (c *formCheck) Amount(name, label string, required bool) (amount decimal.Decimal, ok bool) {
	value := strings.TrimSpace(c.req.FormValue(name))
	if value == "" {
		if required {
			c.fail("%s saknas", label)
		}
		return decimal.Zero, false
	}
	amount, err := parseAmount(value)
	if err != nil {
		c.fail("%s: %v", label, err)
		return decimal.Zero, false
	}
	return amount, true
}

// Int läser ett heltal mellan min och max. Ett tomt fält som inte är obligatoriskt
// ger def.
func (c *formCheck) Int(name, label string, min, max int, required bool, def int) int {
	value := strings.TrimSpace(c.req.FormValue(name))
	if value == "" {
		if required {
			c.fail("%s saknas", label)
		}
		return def
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		c.fail("%s: '%s' är inte ett heltal", label, value)
		return def
	}
	if n < min || n > max {
		c.fail("%s: %d är utanför %d–%d", label, n, min, max)
		return def
	}
	return n
}

// Lopnr läser ett löpnummer (positivt heltal). Saknas det returneras -1.
func (c *formCheck) Lopnr(name string) int {
	return c.Int(name, "Löpnr", 0, 1<<31-1, false, -1)
}

// Bool läser ett fält som true/false (även 1/0, on). Tomt ger false.
func (c *formCheck) Bool(name, label string) bool {
	value := strings.TrimSpace(c.req.FormValue(name))
	if value == "" {
		return false
	}
	if strings.EqualFold(value, "on") {
		return true
	}
	b, err := strconv.ParseBool(value)
	if err != nil {
		c.fail("%s: '%s' är inte sant/falskt", label, value)
		return false
	}
	return b
}

var monthNames = []string{"Jan", "Feb", "Mar", "Apr", "Maj", "Jun", "Jul", "Aug", "Sep", "Okt", "Nov", "Dec"}

// Month läser en månad, som nummer 1–12 eller namn (Jan, Feb, ...). Värdet
// returneras som det skrevs.
func (c *formCheck) Month(name, label string) string {
	value := strings.TrimSpace(c.req.FormValue(name))
	if n, err := strconv.Atoi(value); err == nil && n >= 1 && n <= 12 {
		return value
	}
	for _, m := range monthNames {
		if strings.EqualFold(value, m) {
			return m
		}
	}
	if value == "" {
		c.fail("%s saknas", label)
	} else {
		c.fail("%s: '%s' är inte en månad (1–12 eller Jan–Dec)", label, value)
	}
	return ""
}

// parseFormOrFail läser formuläret. Vid fel skrivs ett felsvar och false returneras.
func parseFormOrFail(w http.ResponseWriter, req *http.Request) bool {
	if err := req.ParseForm(); err != nil {
		log.Println("Kunde inte läsa formulärdata: ", err)
		http.Error(w, "Kunde inte läsa formulärdata.", http.StatusBadRequest)
		return false
	}
	return true
}
