package main

import (
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

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
