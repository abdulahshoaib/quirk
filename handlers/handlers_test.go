package handlers

import (
	"bytes"
	// "context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	// "sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	// _ "github.com/mattn/go-sqlite3" // SQLite driver for testing
)

// Test setup and teardown
func TestMain(m *testing.M) {
	// Setup test database
	setupTestDB()

	// Setup test routes
	mux := http.NewServeMux()
	mux.HandleFunc("/login", logging(HandleSignup))
	mux.HandleFunc("/process", logging(HandleProcess))
	mux.HandleFunc("/status", logging(HandleStatus))
	mux.HandleFunc("/result", logging(HandleResult))
	mux.HandleFunc("/export", logging(HandleExport))
	mux.HandleFunc("/signup", logging(HandleSignup))
	mux.HandleFunc("/export-chroma", logging(HandleExportToChroma))
	mux.HandleFunc("/query", logging(HandleQuery))

	// Run tests
	code := m.Run()

	// Cleanup
	teardownTestDB()
	os.Exit(code)
}

func logging(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.URL.Path)
		fn(w, r)
	}
}

// Mock database setup
func setupTestDB() {
	// Create in-memory SQLite database for testing
	db, err := sql.Open("postgres", "file::memory:?cache=shared")
	if err != nil {
		log.Fatal("Failed to open test database:", err)
	}

	// Create test table
	_, err = db.Exec(`
		CREATE TABLE user_tokens (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT NOT NULL,
			token TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Fatal("Failed to create test table:", err)
	}

	InitDB(db)
}

func teardownTestDB() {
	if Db != nil {
		Db.Close()
	}
}

// Reset global state between tests
func resetGlobalState() {
	mutex.Lock()
	defer mutex.Unlock()
	jobStatuses = make(map[string]JobStatus)
	jobResults = make(map[string]Result)
}

// Test HandleSignup
func TestHandleSignup(t *testing.T) {
	resetGlobalState()

	tests := []struct {
		name           string
		method         string
		body           interface{}
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "Valid signup",
			method:         http.MethodPost,
			body:           UserCredentials{Email: "test@example.com"},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.NewDecoder(rr.Body).Decode(&response)
				if err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}
				if response["email"] != "test@example.com" {
					t.Errorf("Expected email test@example.com, got %s", response["email"])
				}
				if response["token"] == "" {
					t.Error("Expected token to be present")
				}
			},
		},
		{
			name:           "Invalid method",
			method:         http.MethodGet,
			body:           UserCredentials{Email: "test@example.com"},
			expectedStatus: http.StatusMethodNotAllowed,
			checkResponse:  nil,
		},
		{
			name:           "Invalid JSON",
			method:         http.MethodPost,
			body:           "invalid json",
			expectedStatus: http.StatusBadRequest,
			checkResponse:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body io.Reader
			if tt.body != nil {
				if str, ok := tt.body.(string); ok {
					body = strings.NewReader(str)
				} else {
					jsonBody, _ := json.Marshal(tt.body)
					body = bytes.NewBuffer(jsonBody)
				}
			}

			req, err := http.NewRequest(tt.method, "/signup", body)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			HandleSignup(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, status)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
		})
	}
}

// Test AuthenticateJWT
func TestAuthenticateJWT(t *testing.T) {
	resetGlobalState()

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		email := r.Context().Value("email").(string)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("Hello %s", email)))
	})

	// Create a valid token
	claims := &Claims{
		Email: "test@example.com",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString(jwtKey)

	// Store token in database
	_, err := Db.Exec(`INSERT INTO user_tokens (email, token) VALUES ($1, $2)`,
		"test@example.com", tokenStr)
	if err != nil {
		t.Fatal("Failed to store test token:", err)
	}

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "Valid token",
			authHeader:     "Bearer " + tokenStr,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing token",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid token",
			authHeader:     "Bearer invalid_token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Malformed header",
			authHeader:     "InvalidFormat",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "/test", nil)
			if err != nil {
				t.Fatal(err)
			}

			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rr := httptest.NewRecorder()
			AuthenticateJWT(testHandler)(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, status)
			}
		})
	}
}

// Test HandleProcess
func TestHandleProcess(t *testing.T) {
	resetGlobalState()

	tests := []struct {
		name           string
		method         string
		setupFiles     func() *bytes.Buffer
		contentType    string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "Invalid method",
			method:         http.MethodGet,
			setupFiles:     nil,
			expectedStatus: http.StatusMethodNotAllowed,
			checkResponse:  nil,
		},
		{
			name:   "No files uploaded",
			method: http.MethodPost,
			setupFiles: func() *bytes.Buffer {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				writer.Close()
				return body
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse:  nil,
		},
		{
			name:   "Valid file upload",
			method: http.MethodPost,
			setupFiles: func() *bytes.Buffer {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)

				// Create a test file
				part, err := writer.CreateFormFile("files", "test.txt")
				if err != nil {
					t.Fatal(err)
				}
				part.Write([]byte("test content"))
				writer.Close()
				return body
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.NewDecoder(rr.Body).Decode(&response)
				if err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}
				if response["object_id"] == "" {
					t.Error("Expected object_id to be present")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			var err error

			if tt.setupFiles != nil {
				body := tt.setupFiles()
				req, err = http.NewRequest(tt.method, "/process", body)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Set("Content-Type", "multipart/form-data")
			} else {
				req, err = http.NewRequest(tt.method, "/process", nil)
				if err != nil {
					t.Fatal(err)
				}
			}

			rr := httptest.NewRecorder()
			HandleProcess(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, status)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
		})
	}
}

// Test HandleStatus
func TestHandleStatus(t *testing.T) {
	resetGlobalState()

	// Setup test data
	testID := uuid.NewString()
	mutex.Lock()
	jobStatuses[testID] = JobStatus{
		Status: "in_progress",
		ETA:    time.Now().Add(5 * time.Second),
		Error:  "",
	}
	mutex.Unlock()

	tests := []struct {
		name           string
		objectID       string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "Missing object_id",
			objectID:       "",
			expectedStatus: http.StatusBadRequest,
			checkResponse:  nil,
		},
		{
			name:           "Valid object_id",
			objectID:       testID,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.NewDecoder(rr.Body).Decode(&response)
				if err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}
				if response["status"] != "in_progress" {
					t.Errorf("Expected status in_progress, got %v", response["status"])
				}
			},
		},
		{
			name:           "Non-existent object_id",
			objectID:       "nonexistent",
			expectedStatus: http.StatusNotFound,
			checkResponse:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/status"
			if tt.objectID != "" {
				url += "?object_id=" + tt.objectID
			}

			req, err := http.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			HandleStatus(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, status)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
		})
	}
}

// Test HandleResult
func TestHandleResult(t *testing.T) {
	resetGlobalState()

	// Setup test data
	testID := uuid.NewString()
	mutex.Lock()
	jobStatuses[testID] = JobStatus{
		Status: "completed",
		ETA:    time.Time{},
		Error:  "",
	}
	jobResults[testID] = Result{
		Embeddings:  [][]float64{{1.0, 2.0}, {3.0, 4.0}},
		Triples:     []string{"triple1", "triple2"},
		Filenames:   []string{"file1.txt", "file2.txt"},
		Filecontent: []string{"content1", "content2"},
	}
	mutex.Unlock()

	tests := []struct {
		name           string
		objectID       string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "Missing object_id",
			objectID:       "",
			expectedStatus: http.StatusBadRequest,
			checkResponse:  nil,
		},
		{
			name:           "Valid completed result",
			objectID:       testID,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response Result
				err := json.NewDecoder(rr.Body).Decode(&response)
				if err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}
				if len(response.Embeddings) != 2 {
					t.Errorf("Expected 2 embeddings, got %d", len(response.Embeddings))
				}
			},
		},
		{
			name:           "Non-existent object_id",
			objectID:       "nonexistent",
			expectedStatus: http.StatusNotFound,
			checkResponse:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/result"
			if tt.objectID != "" {
				url += "?object_id=" + tt.objectID
			}

			req, err := http.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			HandleResult(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, status)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
		})
	}
}

// Test HandleExport
func TestHandleExport(t *testing.T) {
	resetGlobalState()

	// Setup test data
	testID := uuid.NewString()
	mutex.Lock()
	jobResults[testID] = Result{
		Embeddings:  [][]float64{{1.0, 2.0}, {3.0, 4.0}},
		Triples:     []string{"triple1", "triple2"},
		Filenames:   []string{"file1.txt", "file2.txt"},
		Filecontent: []string{"content1", "content2"},
	}
	mutex.Unlock()

	tests := []struct {
		name           string
		objectID       string
		format         string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "Missing object_id",
			objectID:       "",
			format:         "json",
			expectedStatus: http.StatusBadRequest,
			checkResponse:  nil,
		},
		{
			name:           "Valid JSON export",
			objectID:       testID,
			format:         "json",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				contentType := rr.Header().Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("Expected content-type application/json, got %s", contentType)
				}
			},
		},
		{
			name:           "Valid CSV export",
			objectID:       testID,
			format:         "csv",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				contentType := rr.Header().Get("Content-Type")
				if contentType != "text/csv" {
					t.Errorf("Expected content-type text/csv, got %s", contentType)
				}
			},
		},
		{
			name:           "Invalid format",
			objectID:       testID,
			format:         "xml",
			expectedStatus: http.StatusBadRequest,
			checkResponse:  nil,
		},
		{
			name:           "Non-existent object_id",
			objectID:       "nonexistent",
			format:         "json",
			expectedStatus: http.StatusNotFound,
			checkResponse:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/export"
			params := []string{}
			if tt.objectID != "" {
				params = append(params, "object_id="+tt.objectID)
			}
			if tt.format != "" {
				params = append(params, "format="+tt.format)
			}
			if len(params) > 0 {
				url += "?" + strings.Join(params, "&")
			}

			req, err := http.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			HandleExport(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, status)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
		})
	}
}

// Test embeddingsToString helper function
func TestEmbeddingsToString(t *testing.T) {
	tests := []struct {
		name     string
		input    []float64
		expected []string
	}{
		{
			name:     "Empty slice",
			input:    []float64{},
			expected: []string{},
		},
		{
			name:     "Single value",
			input:    []float64{1.5},
			expected: []string{"1.5"},
		},
		{
			name:     "Multiple values",
			input:    []float64{1.0, 2.5, 3.14159},
			expected: []string{"1", "2.5", "3.14159"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := embeddingsToString(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected length %d, got %d", len(tt.expected), len(result))
			}
			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("Expected %s at index %d, got %s", expected, i, result[i])
				}
			}
		})
	}
}

// Test enableCors helper function
func TestEnableCors(t *testing.T) {
	rr := httptest.NewRecorder()
	var w http.ResponseWriter = rr
	enableCors(&w)

	expectedHeaders := map[string]string{
		"Access-Control-Allow-Origin":      "*",
		"Access-Control-Allow-Methods":     "GET, POST, OPTIONS, PUT, DELETE",
		"Access-Control-Allow-Headers":     "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With",
		"Access-Control-Allow-Credentials": "true",
	}

	for key, expected := range expectedHeaders {
		actual := rr.Header().Get(key)
		if actual != expected {
			t.Errorf("Expected header %s to be %s, got %s", key, expected, actual)
		}
	}
}

// Benchmark tests
func BenchmarkHandleSignup(b *testing.B) {
	resetGlobalState()

	creds := UserCredentials{Email: "bench@example.com"}
	jsonBody, _ := json.Marshal(creds)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest(http.MethodPost, "/signup", bytes.NewBuffer(jsonBody))
		rr := httptest.NewRecorder()
		HandleSignup(rr, req)
	}
}

func BenchmarkHandleStatus(b *testing.B) {
	resetGlobalState()

	testID := uuid.NewString()
	mutex.Lock()
	jobStatuses[testID] = JobStatus{
		Status: "in_progress",
		ETA:    time.Now().Add(5 * time.Second),
		Error:  "",
	}
	mutex.Unlock()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest(http.MethodGet, "/status?object_id="+testID, nil)
		rr := httptest.NewRecorder()
		HandleStatus(rr, req)
	}
}
