package main

import (
	_ "embed"
	"io/ioutil"
	"os"
	"testing"
)

func SkapaTomMDB(t *testing.T, filename string) {
	if FileExists(filename) {
		// Delete file
		err := os.Remove(filename)
		if err != nil {
			t.Error("Failed to remove file. ", err)
		} else {
			t.Log("OpenMDB testfile removed. OK.")
		}
	} else {
		t.Log("OpenMDB testfile did not exist. OK.")
	}

	// Create and write file
	if JetDBSupport {
	   err := ioutil.WriteFile(filename, TOMDB, 0644)
	   if err != nil {
	      t.Error("Failed to create file. ", err)
	   } else {
	      t.Log("OpenMDB testfile created. OK.")
	   }
	}
}

func TestOpenMDB(t *testing.T) {
	var filename = "gotest.mdb"

	if !JetDBSupport {
		t.Log("MDB/JetDB not supported.")
		return
	} else {
		t.Log("MDB/JetDB supported. OK.")
	}
	SkapaTomMDB(t, filename)

	// Check open succeeds
	db = openJetDB(filename, false)
	if db != nil {
		t.Log("OpenMDB succeeded.")
		closeDB()
	} else {
		t.Error("OpenMDB failed to open file.")
	}
}

func TestOpenDB(t *testing.T) {
	var filename = "gotest."

	if JetDBSupport {
		t.Log("Jet Supported.")
		filename = filename + "mdb"
		SkapaTomMDB(t, filename)

		// Check open succeeds
		db = openJetDB(filename, false)
		if db != nil {
			t.Log("OpenMDB succeeded.")
			closeDB()
		} else {
			t.Error("OpenMDB failed to open file.")
		}
	} else {
		t.Log("Jet NOT Supported.")
		filename = filename + "db"
		SkapaTomDB(filename)

		// Check open succeeds
		db = openSqlite(filename)
		if db != nil {
			t.Log("OpenDB succeeded.")
			closeDB()
		} else {
			t.Error("OpenDB failed to open file.")
		}
	}
}
