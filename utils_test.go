//-*- coding: utf-8 -*-

package main

import (
	_ "embed"
	"testing"

	"github.com/shopspring/decimal" // MIT License
)

func TestSaniAmount(t *testing.T) {
	var amount string
	amount = SanitizeAmount("1,30")
	if amount == "1.30" {
		t.Log("T1 succeeded.")
	} else {
		t.Error("T1 failed: " + amount)
	}
	amount = SanitizeAmount("1.30")
	if amount == "1.30" {
		t.Log("T2 ok.")
	} else {
		t.Error("T2 failed: " + amount)
	}
	amount = SanitizeAmount("1.30kr")
	if amount == "1.30" {
		t.Log("T3 ok.")
	} else {
		t.Error("T3 failed: " + amount)
	}
	amount = SanitizeAmount("-1.30")
	if amount == "-1.30" {
		t.Log("T4 ok.")
	} else {
		t.Error("T4 failed: " + amount)
	}
	amount = SanitizeAmount("1 001.30")
	if amount == "1001.30" {
		t.Log("T5 ok.")
	} else {
		t.Error("T5 failed: " + amount)
	}
	amount = SanitizeAmount("1 001,30")
	if amount == "1001.30" {
		t.Log("T6 ok.")
	} else {
		t.Error("T6 failed: " + amount)
	}
	amount = SanitizeAmount("1.001,30")
	if amount == "1001.30" {
		t.Log("T7 ok.")
	} else {
		t.Error("T7 failed: " + amount)
	}
	amount = SanitizeAmount("1,001.30")
	if amount == "1001.30" {
		t.Log("T8 ok.")
	} else {
		t.Error("T8 failed: " + amount)
	}
	amount = SanitizeAmount(".30")
	if amount == ".30" {
		t.Log("T9 ok.")
	} else {
		t.Error("T9 failed: " + amount)
	}
	amount = SanitizeAmount("")
	if amount == "0" {
		t.Log("T10 ok.")
	} else {
		t.Error("T10 failed: " + amount)
	}
	amount = SanitizeAmount("SEK1.30kr")
	if amount == "1.30" {
		t.Log("T11 ok.")
	} else {
		t.Error("T11 failed: " + amount)
	}
	amount = SanitizeAmount("räksmörgås")
	if amount == "0" {
		t.Log("T12 ok.")
	} else {
		t.Error("T12 failed: " + amount)
	}
	amount = SanitizeAmount("Ⅷ")
	if amount == "0" {
		t.Log("T13 ok.")
	} else {
		t.Error("T13 failed: " + amount)
	}
	amount = SanitizeAmount("½")
	if amount == "0" {
		t.Log("T14 ok.")
	} else {
		t.Error("T14 failed: " + amount)
	}
	amount = SanitizeAmount(",30")
	if amount == ".30" {
		t.Log("T15 ok.")
	} else {
		t.Error("T15 failed: " + amount)
	}
	amount = SanitizeAmount("1,001.")
	if amount == "1001." {
		t.Log("T16 ok.")
	} else {
		t.Error("T16 failed: " + amount)
	}
	amount = SanitizeAmount("1,001")
	if amount == "1.001" {
		t.Log("T17 ok.")
	} else {
		t.Error("T17 failed: " + amount)
	}
}

// generic test for decimal
func TestDec1(t *testing.T) {
	var n = decimal.NewFromInt(0)
	incr, _ := decimal.NewFromString(".01")

	for i := 0; i < 1000; i++ {
		n = n.Add(incr)
	}
	var expected = decimal.NewFromInt(10)
	if n.Equal(expected) {
		t.Log("TestDec1 ok.")
	} else {
		t.Error("TestDec1 failed: " + n.String())
	}
	if n.String() == "10" {
		t.Log("TestDec1 ok.")
	} else {
		t.Error("TestDec1 failed: " + n.String())
	}
}

func TestDec2Str(t *testing.T) {
	var n = decimal.NewFromInt(0)
	amount := Dec2Str(n)
	expected := "0,00"
	if amount == expected {
		t.Log("T1 succeeded.")
	} else {
		t.Error("T1 failed: " + amount + " " + expected)
	}
	n = decimal.NewFromInt(10)
	amount = Dec2Str(n)
	expected = "10,00"
	if amount == expected {
		t.Log("T1 succeeded.")
	} else {
		t.Error("T1 failed: " + amount + " " + expected)
	}
	n = decimal.NewFromInt(123)
	amount = Dec2Str(n)
	expected = "123,00"
	if amount == expected {
		t.Log("T1 succeeded.")
	} else {
		t.Error("T1 failed: " + amount + " " + expected)
	}
	n = decimal.NewFromInt(1000)
	amount = Dec2Str(n)
	expected = "1 000,00"
	if amount == expected {
		t.Log("T1 succeeded.")
	} else {
		t.Error("T1 failed: " + amount + " " + expected)
	}
	n = decimal.NewFromInt(10000)
	amount = Dec2Str(n)
	expected = "10 000,00"
	if amount == expected {
		t.Log("T1 succeeded.")
	} else {
		t.Error("T1 failed: " + amount + " " + expected)
	}
	n = decimal.NewFromInt(1000000)
	amount = Dec2Str(n)
	expected = "1 000 000,00"
	if amount == expected {
		t.Log("T1 succeeded.")
	} else {
		t.Error("T1 failed: " + amount + " " + expected)
	}
	n,_ = decimal.NewFromString("0.1")
	amount = Dec2Str(n)
	expected = "0,10"
	if amount == expected {
		t.Log("T1 succeeded.")
	} else {
		t.Error("T1 failed: " + amount + " " + expected)
	}
	n,_ = decimal.NewFromString("0.12")
	amount = Dec2Str(n)
	expected = "0,12"
	if amount == expected {
		t.Log("T1 succeeded.")
	} else {
		t.Error("T1 failed: " + amount + " " + expected)
	}
	n,_ = decimal.NewFromString("0.123")
	amount = Dec2Str(n)
	expected = "0,123"
	if amount == expected {
		t.Log("T1 succeeded.")
	} else {
		t.Error("T1 failed: " + amount + " " + expected)
	}
	n,_ = decimal.NewFromString("3000000.12")
	amount = Dec2Str(n)
	expected = "3 000 000,12"
	if amount == expected {
		t.Log("T1 succeeded.")
	} else {
		t.Error("T1 failed: " + amount + " " + expected)
	}
	n,_ = decimal.NewFromString("3123456.12")
	amount = Dec2Str(n)
	expected = "3 123 456,12"
	if amount == expected {
		t.Log("T1 succeeded.")
	} else {
		t.Error("T1 failed: " + amount + " " + expected)
	}
	n,_ = decimal.NewFromString("3567.12")
	amount = Dec2Str(n)
	expected = "3 567,12"
	if amount == expected {
		t.Log("T1 succeeded.")
	} else {
		t.Error("T1 failed: " + amount + " " + expected)
	}
	n,_ = decimal.NewFromString("-0.12")
	amount = Dec2Str(n)
	expected = "-0,12"
	if amount == expected {
		t.Log("T1 succeeded.")
	} else {
		t.Error("T1 failed: " + amount + " " + expected)
	}
	n,_ = decimal.NewFromString("-378.70")
	amount = Dec2Str(n)
	expected = "-378,70"
	if amount == expected {
		t.Log("T1 succeeded.")
	} else {
		t.Error("T1 failed: " + amount + " " + expected)
	}
	n,_ = decimal.NewFromString("-2378.70")
	amount = Dec2Str(n)
	expected = "-2 378,70"
	if amount == expected {
		t.Log("T1 succeeded.")
	} else {
		t.Error("T1 failed: " + amount + " " + expected)
	}
}
