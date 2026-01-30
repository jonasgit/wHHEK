//-*- coding: utf-8 -*-

package main

import (
	"math"
	"runtime"
	"strings"
	"unicode"

	"golang.org/x/text/encoding/charmap"

	"github.com/shopspring/decimal" // MIT License
)

// SanitizeAmount Remove any non-digit
// allow only one decimal character ","
// set to "0" if empty string
func SanitizeAmount(amount string) string {
	//fmt.Println("Sanitize IN:" + amount)
	amount = strings.ReplaceAll(amount, ",", ".")

	var newamount = ""
	for _, char := range amount {
		if unicode.IsDigit(char) {
			newamount += string(char)
		}
		if char == '.' {
			newamount += string(char)
		}
		if char == '-' {
			newamount += string(char)
		}
	}
	amount = newamount

	if strings.Count(amount, ".") > 1 {
		amount = strings.Replace(amount, ".", "", strings.Count(amount, ".")-1)
	}

	if len(amount) == 0 {
		amount = "0"
	}

	//fmt.Println("Sanitize OUT:" + amount)
	return amount
}

func SanitizeAmountb(amount []byte) string {
	return SanitizeAmount(toUtf8(amount))
}

// Amount2DecStr Convert string for use with new decimal
func Amount2DecStr(amount string) string {
	//fmt.Println("Amount2DecStr IN:" + amount)
	outamount := strings.ReplaceAll(amount, ",", ".")
	//fmt.Println("Amount2DecStr OUT:" + outamount)
	return outamount
}

// AmountDec2DBStr Convert decimal to string for use with sql-statements
func AmountDec2DBStr(summa decimal.Decimal) string {
	//fmt.Println("AmountDec2DBStr IN:" + summa.String())
	outsumma := AmountStr2DBStr(summa.String())
	//fmt.Println("AmountDec2DBStr OUT:" + outsumma)
	return outsumma
}
func AmountStr2DBStr(summa string) string {
	//fmt.Println("AmountStr2DBStr IN:" + summa)

	if JetDBSupport && (dbtype == 1) {
		if !dbdecimaldot {
			outsumma := strings.ReplaceAll(summa, ".", ",")
			//fmt.Println("AmountStr2DBStr OUT1:" + outsumma)
			return outsumma
		}
	}
	//fmt.Println("AmountStr2DBStr OUT2:" + summa)
	return summa
}

// Konvertera en byte-array till sträng
// Arrayen antas vara Windows1252
// Se https://en.wikipedia.org/wiki/Windows-1252
// https://web.archive.org/web/20240703024933/https://devblogs.microsoft.com/oldnewthing/20240702-00/?p=109951
func toUtf8(inBuf []byte) string {
	buf := inBuf
	if dbtype == 1 {
		buf, _ = charmap.Windows1252.NewDecoder().Bytes(inBuf)
	}
	stringVal := string(buf)
	return stringVal
}

// Rensa bort känsliga tecken i filnamn
func sanitizeFilename(fname string) string {
	fname = strings.Replace(fname, "\\", "", -1)
	fname = strings.Replace(fname, "/", "", -1)
	fname = strings.Replace(fname, "'", "", -1)
	fname = strings.Replace(fname, "<", "", -1)
	fname = strings.Replace(fname, ">", "", -1)
	fname = strings.Replace(fname, "\"", "", -1)
	fname = strings.Replace(fname, ":", "", -1)

	return fname
}

// Konvertera decimal-typen till string enligt svensk standard för belopp
// Referenser
// Myndigheternas skrivregler Ds 2009:38, kap 9.5
// Valuta tas ej med (ISO 4217)
func Dec2Str(summa decimal.Decimal) string {
	var sign string
	isnegative := summa.IsNegative()
	if isnegative {
		sign = "-"
		summa = summa.Abs()
	} else {
		sign = ""
	}
	// byt punkt till komma
	s2 := strings.ReplaceAll(summa.String(), ".", ",")

	// dela upp mellan heltal och decimaler
	ss := strings.Split(s2, ",")
	var ints, decs string
	if len(ss) == 1 {
		ints = ss[0]
		decs = ",00"
	} else {
		ints = ss[0]
		decs = "," + ss[1]
	}
	// fyll ut till minst två decimaler
	for len(decs) < 3 {
		decs += "0"
	}
	// dela upp heltal med mellanslag
	if len(ints) > 6 {
		len := len(ints)
		ints = ints[0:len-6] + " " + ints[len-6:len-3] + " " + ints[len-3:len]
	} else if len(ints) > 3 {
		len := len(ints)
		ints = ints[0:len-3] + " " + ints[len-3:len]
	}
	return sign + ints + decs
}

// Dec2Str3Sig formats a decimal to a maximum of 3 significant digits
// Uses Swedish formatting (comma as decimal separator, space as thousands separator)
func Dec2Str3Sig(summa decimal.Decimal) string {
	if summa.IsZero() {
		return "0,00"
	}

	// Convert to float64 for significant digit calculation
	val, _ := summa.Float64()
	absVal := math.Abs(val)

	// Round to 3 significant digits
	var rounded float64
	if absVal != 0 {
		magnitude := math.Pow(10, math.Floor(math.Log10(absVal))-2)
		rounded = math.Round(val/magnitude) * magnitude
	} else {
		rounded = 0
	}

	// Convert back to decimal (preserving sign)
	roundedDec := decimal.NewFromFloat(rounded)

	// Use Dec2Str for formatting (it handles sign internally)
	formatted := Dec2Str(roundedDec)

	// The rounding already ensures 3 significant digits, so Dec2Str should handle the rest
	return formatted
}

// getCurrentFuncName will return the current function's name.
func getCurrentFuncName() string {
	pc, _, _, _ := runtime.Caller(1)
	return runtime.FuncForPC(pc).Name()
}
