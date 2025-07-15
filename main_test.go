package main

import (
	"log"
	"testing"
)

func TestMain(t *testing.T) {

	var err error
	db, err = connect()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Basic test to verify package compilation
	t.Log("Main package compiled successfully")

}
