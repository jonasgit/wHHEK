package main

import (
	"database/sql"
	_ "embed"
	"os"
	"path/filepath"
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
		err := os.WriteFile(filename, TOMDB, 0644)
		if err != nil {
			t.Error("Failed to create file. ", err)
		} else {
			t.Log("OpenMDB testfile created. OK.")
		}
	}
}

// newTestDB creates an empty database of its own for one test, in a temporary
// directory that is removed after the test. It does not touch the global db,
// dbtype or currentDatabase, so tests using it can run in parallel
// (t.Parallel()). On Windows/Jet the global dbtype is set to 1 once in TestMain.
func newTestDB(t *testing.T, name string) *sql.DB {
	t.Helper()
	dir := t.TempDir()

	var conn *sql.DB
	var err error
	if JetDBSupport {
		filename := filepath.Join(dir, "got"+name+".mdb")
		t.Log("Jet Supported.")
		SkapaTomMDB(t, filename)
		conn, err = connectJetDB(filename, false)
	} else {
		filename := filepath.Join(dir, "got"+name+".db")
		t.Log("Jet NOT Supported.")
		conn, err = sql.Open("sqlite3", filename)
		if err == nil {
			InitiateDB(conn)
		}
	}
	if err != nil || conn == nil {
		t.Fatal("Ingen databas.", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

// keepGlobals restores the global database state after a test that uses the
// global open/close functions (openSqlite, openJetDB, closeDB). Such tests must
// not call t.Parallel().
func keepGlobals(t *testing.T) {
	t.Helper()
	oldDB, oldType, oldName := db, dbtype, currentDatabase
	t.Cleanup(func() {
		db, dbtype, currentDatabase = oldDB, oldType, oldName
	})
}

// Tests openJetDB and closeDB, which use the global db: not parallel.
func TestOpenMDB(t *testing.T) {
	if !JetDBSupport {
		t.Log("MDB/JetDB not supported.")
		return
	} else {
		t.Log("MDB/JetDB supported. OK.")
	}
	keepGlobals(t)
	filename := filepath.Join(t.TempDir(), "gotest.mdb")
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

// Tests openJetDB/openSqlite and closeDB, which use the global db: not parallel.
func TestOpenDB(t *testing.T) {
	keepGlobals(t)
	filename := filepath.Join(t.TempDir(), "gotest.")

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
		closeDB() // SkapaTomDB leaves the new database open in the global db

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
