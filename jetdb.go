package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/alexbrainman/odbc" // BSD-3-Clause License
)

func openJetDB(filename string, ro bool) *sql.DB {
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
		os.Exit(1)
	}
	return db
}
