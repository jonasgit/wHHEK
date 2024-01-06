//-*- coding: utf-8 -*-
//go:build !(windows && 386)

package main

import (
	"database/sql"
)

// JetDBSupport Global variables
var JetDBSupport = false

func openJetDB(filename string, ro bool) *sql.DB {
	// No support for JetDB
	return nil
}

var TOMDB []byte  // ONLY REQUIRED FOR Go UNIT TEST

// se jetdb.go
func detectdbdec() bool {
	return false
}
