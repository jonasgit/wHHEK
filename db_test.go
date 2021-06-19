package main

import (
	_ "embed"
	"io/ioutil"
	"testing"
	"os"
)

//go:embed TOMDB.MDB
var TOMDB []byte

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

func TestOpenMDB(t *testing.T) {
	var filename string = "gotest.mdb"

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
	err := ioutil.WriteFile(filename, TOMDB, 0644)
	if err != nil {
		t.Error("Failed to create file. ", err)
	} else {
		t.Log("OpenMDB testfile created. OK.")
	}
	
	// Check open succeeds
	db = openJetDB(filename, false)
	if db != nil {
		t.Log("OpenMDB succeeded.")
	} else {
		t.Error("OpenMDB failed to open file.")
	}
}
