//-*- coding: utf-8 -*-
// +build windows,386

package main

// TODO: testa https://github.com/mattn/go-adodb

// File format should be Access 1.1 or 2.0 as described by
// http://fileformats.archiveteam.org/wiki/Access
// that is byte 0x40A should be 8 or 9. And Engine Type should be 2 or 3.

// See also https://www.loc.gov/preservation/digital/formats/fdd/fdd000462.shtml

// TODO: Testa om det går att skapa en tom databas med:
// "Provider=Microsoft.Jet.OLEDB.4.0;Jet OLEDB:Engine Type=" & Format & ";Data Source=" & DestDB
// Där format är 2 eller 3 alltså

import (
	"database/sql"
	"log"

	_ "github.com/alexbrainman/odbc" // BSD-3-Clause License
)

// Global variables
var JetDBSupport bool = true

func openJetDB(filename string, ro bool) *sql.DB {
	currentDatabase = "NONE"
	dbtype = 0

	readonlyCommand := ""
	if ro {
		readonlyCommand = "READONLY;"
	}

	databaseAccessCommand := "Driver={Microsoft Access Driver (*.mdb)};" +
		readonlyCommand +
		"DBQ=" + filename
	//fmt.Println("Database access command: "+databaseAccessCommand)
	db, err := sql.Open("odbc",
		databaseAccessCommand)
	if err != nil {
		// TÃ¤nkbara alternativa connect-strÃ¤ngar om ovanstÃ¥ende inte fungerar:
		// "Driver={Microsoft Access Driver (*.mdb)};dbq="+file
		// "Driver={MS Access Database};DBQ="
		// "Driver={MS Access Driver (*.mdb)};dbq="+
		// "Provider=Microsoft.Jet.OLEDB.4.0
		// https://www.connectionstrings.com/access/
		// https://docs.microsoft.com/en-us/office/client-developer/access/desktop-database-reference/microsoft-ole-db-provider-for-microsoft-jet
		log.Fatal(err)
		return nil
	}
	currentDatabase = filename
	dbtype = 1
	return db
}
