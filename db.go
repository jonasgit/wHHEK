//-*- coding: utf-8 -*-

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	
	_ "github.com/mattn/go-sqlite3"
)

/* The below function checks if a regular file (not directory) with a
   given filepath exist */
func FileExists (filepath string) bool {
	
	fileinfo, err := os.Stat(filepath)
	
	if os.IsNotExist(err) {
		return false
	}
	// Return false if the fileinfo says the file path is a directory.
	return !fileinfo.IsDir()
}

func openSqlite(filename string) *sql.DB {
	currentDatabase = "NONE"
	dbtype = 0
	
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		log.Fatal(err)
	}
	
	currentDatabase = filename
	dbtype = 2
	
	return db
}

func SkapaTomDB(filename string) {
	if FileExists(filename) {
		// Delete file
		err := os.Remove(filename)
		if err != nil {
			fmt.Println("Failed to remove file. ", err)
		} else {
			fmt.Println("SkapaTomDB file removed. OK.")
		}
	} else {
		fmt.Println("SkapaTomDB file did not exist. OK.")
	}
	
	// Create file
	openSqlite(filename)
        db, err := sql.Open("sqlite3", filename)
 	if err != nil {
		fmt.Println("Failed to create file. ", err)
	} else {
		fmt.Println("SkapTomDB file created. OK.")
	}
 	if db == nil {
		fmt.Println("Failed to create database. ", err)
	} else {
		fmt.Println("SkapTomDB database created. OK.")
	}
}
