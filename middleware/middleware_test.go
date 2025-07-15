package middleware

import (
	// "database/sql"
	"database/sql/driver"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestInitDB tests the database initialization
func TestInitDB(t *testing.T) {
	// Create a mock database
	mockDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	// Test initialization
	InitDB(mockDB)

	// Verify that the global db variable is set
	if db == nil {
		t.Error("Database was not initialized properly")
	}

	if db != mockDB {
		t.Error("Database instance does not match the initialized one")
	}
}

// TestLogging tests the logging middleware
func TestLogging(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		method   string
		expected string
	}{
		{
			name:     "GET request to root",
			path:     "/",
			method:   "GET",
			expected: "/",
		},
		{
			name:     "POST request to API endpoint",
			path:     "/api/users",
			method:   "POST",
			expected: "/api/users",
		},
		{
			name:     "PUT request with query parameters",
			path:     "/api/users/123?update=true",
			method:   "PUT",
			expected: "/api/users/123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture log output
			logOutput := captureLogOutput(func() {
				// Create a test handler that does nothing
				testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				})

				// Wrap with logging middleware
				loggedHandler := Logging(testHandler)

				// Create test request
				req := httptest.NewRequest(tt.method, tt.path, nil)
				rr := httptest.NewRecorder()

				// Execute the handler
				loggedHandler(rr, req)

				// Check that the handler was called successfully
				if rr.Code != http.StatusOK {
					t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
				}
			})

			// Check that the path was logged
			if !strings.Contains(logOutput, tt.expected) {
				t.Errorf("Expected log to contain '%s', got '%s'", tt.expected, logOutput)
			}
		})
	}
}

// TestAuth tests the authentication middleware
func TestAuth(t *testing.T) {
	tests := []struct {
		name           string
		authHeader     string
		mockQuery      bool
		mockResult     bool
		mockError      error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "missing authorization header",
			authHeader:     "",
			mockQuery:      false,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Missing or invalid Authorization Header",
		},
		{
			name:           "invalid authorization header format",
			authHeader:     "Basic dXNlcjpwYXNz",
			mockQuery:      false,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Missing or invalid Authorization Header",
		},
		{
			name:           "empty bearer token",
			authHeader:     "Bearer ",
			mockQuery:      false,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Empty bearer token",
		},
		{
			name:           "whitespace only bearer token",
			authHeader:     "Bearer   ",
			mockQuery:      false,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Empty bearer token",
		},
		{
			name:           "valid token - database query succeeds",
			authHeader:     "Bearer valid-token-123",
			mockQuery:      true,
			mockResult:     true,
			expectedStatus: http.StatusOK,
			expectedBody:   "Success",
		},
		{
			name:           "valid token - but not found in database",
			authHeader:     "Bearer invalid-token-456",
			mockQuery:      true,
			mockResult:     false,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized: invalid token",
		},
		{
			name:           "valid token - database query error",
			authHeader:     "Bearer valid-token-789",
			mockQuery:      true,
			mockResult:     false,
			mockError:      fmt.Errorf("database connection error"),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized: invalid token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock database
			mockDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to create mock database: %v", err)
			}
			defer mockDB.Close()

			// Initialize the middleware with mock database
			InitDB(mockDB)

			// Set up database mock expectations
			if tt.mockQuery {
				token := strings.TrimSpace(strings.TrimPrefix(tt.authHeader, "Bearer "))
				if tt.mockError != nil {
					mock.ExpectQuery("SELECT EXISTS").
						WithArgs(token).
						WillReturnError(tt.mockError)
				} else {
					rows := sqlmock.NewRows([]string{"exists"}).AddRow(tt.mockResult)
					mock.ExpectQuery("SELECT EXISTS").
						WithArgs(token).
						WillReturnRows(rows)
				}
			}

			// Create test handler
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Success"))
			})

			// Wrap with auth middleware
			authHandler := Auth(testHandler)

			// Create test request
			req := httptest.NewRequest("GET", "/protected", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rr := httptest.NewRecorder()

			// Execute the handler
			authHandler(rr, req)

			// Check status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			// Check response body
			body := strings.TrimSpace(rr.Body.String())
			if !strings.Contains(body, tt.expectedBody) {
				t.Errorf("Expected body to contain '%s', got '%s'", tt.expectedBody, body)
			}

			// Verify all mock expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Mock expectations were not met: %v", err)
			}
		})
	}
}

// TestAuthMiddlewareChaining tests that the auth middleware can be chained with other middleware
func TestAuthMiddlewareChaining(t *testing.T) {
	// Create mock database
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	// Initialize the middleware with mock database
	InitDB(mockDB)

	// Mock successful token validation
	rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs("valid-token").
		WillReturnRows(rows)

	// Test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Protected resource"))
	})

	// Chain middleware: Logging -> Auth -> Handler
	handler := Logging(Auth(testHandler))

	// Create test request
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()

	// Capture log output
	logOutput := captureLogOutput(func() {
		handler(rr, req)
	})

	// Check response
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	if !strings.Contains(rr.Body.String(), "Protected resource") {
		t.Error("Expected response body to contain 'Protected resource'")
	}

	// Check that logging occurred
	if !strings.Contains(logOutput, "/protected") {
		t.Error("Expected log output to contain '/protected'")
	}

	// Verify mock expectations
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Mock expectations were not met: %v", err)
	}
}

// TestAuthWithoutDatabaseInit tests auth middleware behavior when database is not initialized
func TestAuthWithoutDatabaseInit(t *testing.T) {
	// Reset the global db variable
	db = nil

	// Create test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Success"))
	})

	// Wrap with auth middleware
	authHandler := Auth(testHandler)

	// Create test request
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()

	// This should panic or return an error because db is nil
	defer func() {
		if r := recover(); r == nil {
			// If it doesn't panic, check that it returns unauthorized
			if rr.Code != http.StatusUnauthorized {
				t.Error("Expected unauthorized status when database is not initialized")
			}
		}
	}()

	authHandler(rr, req)
}

// BenchmarkLogging benchmarks the logging middleware
func BenchmarkLogging(b *testing.B) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	loggedHandler := Logging(testHandler)
	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		loggedHandler(rr, req)
	}
}

// BenchmarkAuth benchmarks the auth middleware
func BenchmarkAuth(b *testing.B) {
	// Create mock database
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	// Initialize the middleware with mock database
	InitDB(mockDB)

	// Mock successful token validation for all benchmark iterations
	for i := 0; i < b.N; i++ {
		rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
		mock.ExpectQuery("SELECT EXISTS").
			WithArgs("benchmark-token").
			WillReturnRows(rows)
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	authHandler := Auth(testHandler)
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer benchmark-token")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		authHandler(rr, req)
	}
}

// Example test demonstrating middleware usage
func ExampleLogging() {
	// Create a simple handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello World"))
	})

	// Wrap with logging middleware
	loggedHandler := Logging(handler)

	// Create a test request
	req := httptest.NewRequest("GET", "/example", nil)
	rr := httptest.NewRecorder()

	// Execute the handler
	loggedHandler(rr, req)

	fmt.Printf("Status: %d\n", rr.Code)
	// Output: Status: 200
}

// Helper function to capture log output
func captureLogOutput(fn func()) string {
	// Create a temporary file to capture log output
	tmpFile, err := os.CreateTemp("", "test_log")
	if err != nil {
		return ""
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Redirect log output to the temporary file
	originalOutput := log.Writer()
	log.SetOutput(tmpFile)
	defer log.SetOutput(originalOutput)

	// Execute the function
	fn()

	// Read the captured output
	tmpFile.Seek(0, 0)
	content := make([]byte, 1024)
	n, _ := tmpFile.Read(content)
	return string(content[:n])
}

// Custom matcher for SQL arguments to handle different types
type anyArg struct{}

func (a anyArg) Match(v driver.Value) bool {
	return true
}

// Example of how to use the middleware in a real application
func ExampleAuth() {
	// Initialize with a real database connection
	// db, _ := sql.Open("postgres", "connection_string")
	// InitDB(db)

	// Create your handler
	protectedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Protected resource"))
	})

	// Apply middleware
	_ = Logging(Auth(protectedHandler))

	// Use in your HTTP server
	// http.Handle("/api/protected", handler)
	// http.ListenAndServe(":8080", nil)

	fmt.Println("Middleware setup complete")
	// Output: Middleware setup complete
}
