// +build !windows,!386

package main

import (
	"database/sql"
)

// Global variables
var JetDBSupport = false

func openJetDB(filename string, ro bool) *sql.DB {
	// No support for JetDB
	return nil
}
