package main

import "testing"
import "os"

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags

	// The tests get their own database connections from newTestDB and run in
	// parallel, so the global dbtype is set once here instead of by each test.
	// It controls decimal comma and Windows-1252 conversion for Jet/mdb.
	if JetDBSupport {
		dbtype = 1
	}
	os.Exit(m.Run())
}
