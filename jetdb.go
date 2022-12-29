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
	_ "embed"
	"io/ioutil"
	"log"
	"os"
	
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

//go:embed TOMDB.MDB
var TOMDB []byte

// Testa om vi ska använda decimalpunkt eller decimalkomma
// vid anrop till databas.
// retur True = använd decimalkomma
// retur False = använd decimalpunkt
func detectdbdec() bool {
	//log.Println("START detectdbdec")
	// Skapa tom DB
	var db *sql.DB = nil
	
	// Create and write file
	f, err := os.CreateTemp("", "db*.mdb")
	if err != nil {
		//log.Println("Failed to create file. ", err)
	} else {
		//log.Println("OpenMDB testfile created. OK.")
	}
	filename := f.Name()
	//log.Println("Got filename: ", filename)
	f.Close();
	os.Remove(filename)
	err = ioutil.WriteFile(filename, TOMDB, 0644)
	if err != nil {
		//log.Println("Failed to create file. ", err)
	} else {
		//log.Println("OpenMDB testfile created. OK.")
	}
	f.Close();
	
	// Check open succeeds
	db = openJetDB(filename, false) // Assume filename is available
	if db != nil {
		//log.Println("OpenMDB succeeded.")
	} else {
		//log.Println("OpenMDB failed to open file.")
		return false
	}

	// Lägg till transaktion och kolla resultat

	/*	result := CheckTransaction(db, "1")
	result = CheckTransaction(db, "1")
	result = CheckTransaction(db, "1,23")
	result = CheckTransaction(db, "1.23")
	result = CheckTransaction(db, "1,0")
	result = CheckTransaction(db, "1.0")
	result = CheckTransaction(db, "1001,23")
	result = CheckTransaction(db, "1001.23")
	result = CheckTransaction(db, "1 001,23")
	result = CheckTransaction(db, "1 001.23")
	result = CheckTransaction(db, "1.001,23")
	result = CheckTransaction(db, "1,001.23")
	result = CheckTransaction(db, "1.001.23")
	result = CheckTransaction(db, "1,001,23") */
	result :=  CheckTransaction(db, "1,23")

	// Radera testDB
	os.Remove(filename)

	//log.Println("END detectdbdec")
	return result
}

// true = samma värde tillbaka
// false = olika värden eller sql-fel
func CheckTransaction(db *sql.DB, value string) bool {
	//log.Println("START CheckTransaction")
	// check db avail
	//antal := antalTransaktioner(db)
	//log.Println("CheckTransaction antal=", antal)
	//SQL insert
	sqlStatement := `
	INSERT INTO Transaktioner (FrånKonto,TillKonto,Typ,Datum,Vad,Vem,Belopp,Saldo,[Fastöverföring],[Text])
	VALUES (?,?,?,?,?,?,?,?,?,?)`

	_, err := db.Exec(sqlStatement, "fromacc", "toacc", "transtyp", "2022-12-01", "what", "who", value, nil, false, "text")
	if err != nil {
		//log.Println("SQL err", err)
		return false
	}
	//log.Println("CheckTransaction inserted")
	//antal = antalTransaktioner(db)
	//log.Println("CheckTransaction antal=", antal)
	//SQL get result
	var val []byte
	Row := db.QueryRow(`SELECT Belopp FROM Transaktioner ORDER BY Löpnr DESC`)
	err = Row.Scan(&val)
	if err != nil {
		//log.Println("SQL err", err)
		return false
	}
	dbval := SanitizeAmountb(val)
	//return bool if eq
	//log.Println("END CheckTransaction")
	//log.Printf("CheckTransaction CMP '%s' '%s'", dbval, SanitizeAmount(value))
	
	return dbval == SanitizeAmount(value)
}
