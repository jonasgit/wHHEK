package main

import (
	"database/sql"
)

func openJetDB(filename string, ro bool) *sql.DB {
	// No support for JetDB
	return nil
}
