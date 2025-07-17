package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

func setupTestDB(t *testing.T) *sql.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"), // use a test DB
	)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("Failed to connect to DB: %v", err)
	}
	return db
}

func TestInitSchema(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	err := InitSchema(db)
	if err != nil {
		t.Fatalf("InitSchema failed: %v", err)
	}

	// Check if table exists
	var exists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_name = 'user_tokens'
		)
	`).Scan(&exists)
	if err != nil {
		t.Fatalf("Failed to check if table exists: %v", err)
	}
	if !exists {
		t.Errorf("Expected table 'user_tokens' to exist, but it doesn't")
	}

	// Check if index exists
	indexes := []string{"idx_user_tokens_token", "idx_user_tokens_email"}
	for _, index := range indexes {
		err = db.QueryRow(`
			SELECT EXISTS (
				SELECT 1 FROM pg_indexes WHERE indexname = $1
			)
		`, index).Scan(&exists)
		if err != nil {
			t.Errorf("Failed to check index %s: %v", index, err)
		}
		if !exists {
			t.Errorf("Expected index %s to exist, but it doesn't", index)
		}
	}
}

func TestRunApp_StartsServer(t *testing.T) {
	// Skip test if DB env vars aren't set
	requiredVars := []string{"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME"}
	skip := false
	for _, v := range requiredVars {
		if os.Getenv(v) == "" {
			t.Logf("Missing environment variable: %s", v)
			skip = true
		}
	}
	if skip {
		t.Skip("Skipping integration test due to missing env vars")
	}

	// Start server in goroutine
	go func() {
		err := RunApp()
		if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
			t.Errorf("RunApp failed: %v", err)
		}
	}()

	time.Sleep(1 * time.Second)

	resp, err := http.Get("http://localhost:8080/status")
	if err != nil {
		t.Fatalf("failed to connect to server: %v", err)
	}
	defer resp.Body.Close()

	// Expect 200, 404, or some valid response
	if resp.StatusCode < 200 || resp.StatusCode >= 500 {
		t.Errorf("unexpected response code: %d", resp.StatusCode)
	}
}
