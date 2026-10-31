//-*- coding: utf-8 -*-
//go:build !(windows && 386)

package main

import (
	"database/sql"
	"errors"
)

// JetDBSupport Global variables
var JetDBSupport = false

func openJetDB(_ string, _ bool) *sql.DB {
	// No support for JetDB
	return nil
}

func connectJetDB(_ string, _ bool) (*sql.DB, error) {
	return nil, errors.New("JetDB/mdb is not supported in this build")
}

var TOMDB []byte // ONLY REQUIRED FOR Go UNIT TEST

// se jetdb.go
func detectdbdec() bool {
	return false
}
