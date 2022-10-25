//-*- coding: utf-8 -*-

package main

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/shopspring/decimal" // MIT License
)

// SanitizeAmount Remove any non-digit
// allow only one decimal character ","
// set to "0" if empty string
func SanitizeAmount(amount string) string {
	fmt.Println("Sanitize IN:" + amount)
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

	fmt.Println("Sanitize OUT:" + amount)
	return amount
}

func SanitizeAmountb(amount []byte) string {
	return SanitizeAmount(toUtf8(amount))
}

// Amount2DecStr Convert string for use with new decimal
func Amount2DecStr(amount string) string {
	fmt.Println("Amount2DecStr IN:" + amount)
	outamount := strings.ReplaceAll(amount, ",", ".")
	fmt.Println("Amount2DecStr OUT:" + outamount)
	return outamount
}

// AmountDec2DBStr Convert decimal to string for use with sql-statements
func AmountDec2DBStr(summa decimal.Decimal) string {
	fmt.Println("AmountDec2DBStr IN:" + summa.String())
	outsumma := AmountStr2DBStr(summa.String())
	fmt.Println("AmountDec2DBStr OUT:" + outsumma)
	return outsumma
}
func AmountStr2DBStr(summa string) string {
	fmt.Println("AmountStr2DBStr IN:" + summa)

	if JetDBSupport && (dbtype == 1) {
		if !dbdecimaldot {
			outsumma := strings.ReplaceAll(summa, ".", ",")
			fmt.Println("AmountStr2DBStr OUT1:" + outsumma)
			return outsumma
		}
	}
	fmt.Println("AmountStr2DBStr OUT2:" + summa)
	return summa
}
