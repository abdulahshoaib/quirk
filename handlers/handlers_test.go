package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	// "sync"
	"testing"
	"time"
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

// Optional: Reset shared maps between tests
func TestMain(m *testing.M) {
	// clean state before tests
	mutex.Lock()
	jobStatuses = make(map[string]JobStatus)
	jobResults = make(map[string]Result)
	mutex.Unlock()

	m.Run()
}

