package main

import "testing"
import "os"

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	os.Exit(m.Run())
}
func TestSum(t *testing.T) {
	t.Error("Sum not implemented.")
	//t.Log("Sum not implemented.")
}
func TestFairy(t *testing.T) {
       t.Log("Fairy is fine.")
}
