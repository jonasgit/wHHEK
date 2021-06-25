package main

// TODO: testa https://github.com/mattn/go-adodb

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
		log.Fatal(err)
		return nil
	}
	currentDatabase = filename
	dbtype = 1
	return db
}
