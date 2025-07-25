package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"

	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestHandleResult_Completed(t *testing.T) {
	// Mock state
	id := "job1"
	jobStatuses[id] = JobStatus{Status: "completed", ETA: time.Now()}
	jobResults[id] = Result{Triples: []string{"a"}, Embeddings: [][]float64{{0.1, 0.2}}}

	req := httptest.NewRequest("GET", "/result?object_id="+id, nil)
	w := httptest.NewRecorder()

	HandleResult(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", w.Code)
	}

	var result Result
	err := json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Fatal("Invalid JSON response")
	}
	if len(result.Embeddings) != 1 {
		t.Error("Expected one embedding")
	}
}

func TestHandleExport_JSON(t *testing.T) {
	id := "job2"
	jobResults[id] = Result{Triples: []string{"a"}, Embeddings: [][]float64{{0.1, 0.2}}}

	req := httptest.NewRequest("GET", "/export?object_id="+id+"&format=json", nil)
	w := httptest.NewRecorder()

	HandleExport(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Error("Expected application/json content type")
	}
}

func TestHandleExport_CSV(t *testing.T) {
	id := "job3"
	jobResults[id] = Result{Triples: []string{"hello"}, Embeddings: [][]float64{{1.1, 2.2}}}

	req := httptest.NewRequest("GET", "/export?object_id="+id+"&format=csv", nil)
	w := httptest.NewRecorder()

	HandleExport(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "hello") {
		t.Error("Expected triple in CSV export")
	}
}

func TestHandleStatus(t *testing.T) {
	id := "job4"
	jobStatuses[id] = JobStatus{Status: "processing", ETA: time.Now().Add(10 * time.Second)}

	req := httptest.NewRequest("GET", "/status?object_id="+id, nil)
	w := httptest.NewRecorder()

	HandleStatus(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "status") {
		t.Error("Expected status in response")
	}
}

func TestHandleResult_MissingID(t *testing.T) {
	req := httptest.NewRequest("GET", "/result", nil)
	w := httptest.NewRecorder()

	HandleResult(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for missing ID, got %d", w.Code)
	}
}

func TestHandleSignup_Success(t *testing.T) {

	// mock DB
	db, mock, _ := sqlmock.New()
	defer db.Close()
	Db = db

	jwtKey = []byte("testsecret") // set your jwtKey

	// expected insert
	mock.ExpectExec("INSERT INTO user_tokens").
		WithArgs("test@example.com", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// request body
	creds := UserCredentials{
		Email: "test@example.com",
	}
	body, _ := json.Marshal(creds)

	req := httptest.NewRequest("POST", "/signup", bytes.NewReader(body))
	w := httptest.NewRecorder()

	HandleSignup(w, req)

	resp := w.Result()

	// check status
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	// check CORS headers
	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("CORS header mismatch: got %s", got)
	}

	// check response body
	var respBody map[string]string
	json.NewDecoder(resp.Body).Decode(&respBody)

	if _, ok := respBody["token"]; !ok {
		t.Errorf("Token not found in response")
	}
	if respBody["email"] != "test@example.com" {
		t.Errorf("Expected email to match, got %s", respBody["email"])
	}
}

func TestHandleSignup_InvalidMethod(t *testing.T) {
	req := httptest.NewRequest("GET", "/signup", nil)
	w := httptest.NewRecorder()

	HandleSignup(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected 405, got %d", w.Code)
	}
}

func TestHandleSignup_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest("POST", "/signup", strings.NewReader("{invalid-json}"))
	w := httptest.NewRecorder()

	HandleSignup(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for invalid JSON, got %d", w.Code)
	}
}

// Optional: Reset shared maps between tests
func TestMain(m *testing.M) {
	// clean state before tests
	mutex.Lock()
	jobStatuses = make(map[string]JobStatus)
	jobResults = make(map[string]Result)
	mutex.Unlock()

	m.Run()
}

func TestInitDB_WithMock(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	defer mockDB.Close()

	InitDB(mockDB)

	if Db != mockDB {
		t.Error("Expected db.Db to be mockDB")
	}
}
func TestInitDB(t *testing.T) {
	// Use an in-memory SQLite DB for testing
	testDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to open test DB: %v", err)
	}
	defer testDB.Close()

	InitDB(testDB)

	if Db == nil {
		t.Fatal("Expected db.Db to be set, got nil")
	}
	if Db != testDB {
		t.Error("Expected db.Db to be assigned to testDB instance")
	}
}
