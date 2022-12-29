// +build !windows,!386

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

func detectdbdec() bool {
	return true
}
