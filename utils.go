//-*- coding: utf-8 -*-

package main

import (
	"fmt"
	"strings"
	"unicode"
	
	"github.com/shopspring/decimal" // MIT License
)

// Remove any non-digit
// allow only one decimal character ","
// set to "0" if empty string
func SanitizeAmount(amount string) string {
	fmt.Println("Sanitize IN:"+amount)
	amount = strings.ReplaceAll(amount, ",", ".")

	var newamount = ""
	for _, char := range amount {
		if unicode.IsDigit(char) {
			newamount += string(char)
		}
		if char == '.' {
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
	
	fmt.Println("Sanitize OUT:"+amount)
	return amount
}

// Convert string for use with new decimal
func Amount2DecStr(amount string) string {
	return strings.ReplaceAll(amount, ",", ".")
}

// Convert decimal to string for use with sql-statements
func AmountDec2DBStr(summa decimal.Decimal) string {
	return AmountStr2DBStr(summa.String())
}
func AmountStr2DBStr(summa string) string {
	fmt.Println("AmountStr2DBStr IN:"+summa)

	if JetDBSupport && (dbtype==1) {
		fmt.Println("AmountStr2DBStr OUT1:"+strings.ReplaceAll(summa, ".", ","))
		return strings.ReplaceAll(summa, ".", ",")
	} else {
		fmt.Println("AmountStr2DBStr OUT2:"+summa)
		return summa
	}
}
