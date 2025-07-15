package main

import (
<<<<<<< Updated upstream
	"os"
	"testing"
)

// TestConnect_Integration tests the connect() function by checking if environment variables
// are set and a DB connection can be established.
func TestConnect_Integration(t *testing.T) {
	requiredVars := []string{"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME"}

	// check if all env vars are set
	missing := false
	for _, v := range requiredVars {
		if os.Getenv(v) == "" {
			t.Logf("Missing environment variable: %s", v)
			missing = true
		}
	}
	if missing {
		t.Skip("Skipping DB connection test due to missing env variables")
	}

	db, err := connect()
	if err != nil {
		t.Fatalf("connect() failed: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Errorf("Ping failed: %v", err)
	}
=======
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

>>>>>>> Stashed changes
}
