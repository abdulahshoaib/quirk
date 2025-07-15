package chromadb

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/abdulahshoaib/quirk/pipeline"
)

func TestCheckHealth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v2/tenants/test_tenant/databases/test_db/heartbeat" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	req := ReqParams{
		Host:     server.Listener.Addr().(*net.TCPAddr).IP.String(),
		Port:     server.Listener.Addr().(*net.TCPAddr).Port,
		Tenant:   "test_tenant",
		Database: "test_db",
	}

	code, err := CheckHealth(req)
	if err != nil || code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d, err: %v", code, err)
	}
}

func TestCreateNewCollection(t *testing.T) {
	var received Payload

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	req := ReqParams{
		Host:          server.Listener.Addr().(*net.TCPAddr).IP.String(),
		Port:          server.Listener.Addr().(*net.TCPAddr).Port,
		Tenant:        "t",
		Database:      "d",
		Collection_id: "c",
	}

	payload := Payload{
		Embeddings: [][]float64{{0.1, 0.2}},
		Documents:  []string{"doc"},
		IDs:        []string{"id1"},
	}

	code, err := CreateNewCollection(req, payload)
	if err != nil || code != http.StatusOK {
		t.Errorf("CreateNewCollection failed: code=%d, err=%v", code, err)
	}
}

func TestUpdateCollection(t *testing.T) {
	var received Payload

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	req := ReqParams{
		Host:          server.Listener.Addr().(*net.TCPAddr).IP.String(),
		Port:          server.Listener.Addr().(*net.TCPAddr).Port,
		Tenant:        "t",
		Database:      "d",
		Collection_id: "c",
	}

	payload := Payload{
		Embeddings: [][]float64{{0.3, 0.4}},
		Documents:  []string{"updated"},
		IDs:        []string{"id1"},
	}

	code, err := UpdateCollection(req, payload)
	if err != nil || code != http.StatusOK {
		t.Errorf("UpdateCollection failed: code=%d, err=%v", code, err)
	}
}
func TestListCollections(t *testing.T) {
	// Mock EmbeddingsAPI
	originalEmbeddingsAPI := pipeline.OverrideEmbeddingsAPI
	defer func() { pipeline.OverrideEmbeddingsAPI = originalEmbeddingsAPI }()
	pipeline.OverrideEmbeddingsAPI = func(texts []string) ([][]float64, error) {
		return [][]float64{{0.1, 0.2, 0.3}}, nil
	}

	// Mock ChromaDB server
	mockResponse := ChromaQueryResponse{
		Documents: [][]string{{"doc1"}},
		Distances: [][]float64{{0.123}},
		IDs:       [][]string{{"id1"}},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v2/tenants/t/databases/d/collections/c/query" {
			json.NewEncoder(w).Encode(mockResponse)
		} else {
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	addr := server.Listener.Addr().(*net.TCPAddr)

	req := ReqParams{
		Host:          addr.IP.String(),
		Port:          addr.Port,
		Tenant:        "t",
		Database:      "d",
		Collection_id: "c",
	}

	statusCode, err, parsed := ListCollections(req, []string{"test query"})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if statusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", statusCode)
	}
	if parsed == nil || len(parsed.Documents) != 1 {
		t.Error("Expected one document in response")
	}
}

func TestCreateNewCollection_HTTPFailure(t *testing.T) {
	req := ReqParams{
		Host:   "localhost",
		Port:   9999, // likely to fail
		Tenant: "t", Database: "d", Collection_id: "c",
	}

	payload := Payload{
		IDs: []string{"id1"},
	}

	code, err := CreateNewCollection(req, payload)
	if err == nil {
		t.Error("Expected error but got none")
	}
	if code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", code)
	}
}

func TestListCollections_InvalidJSON(t *testing.T) {
	pipeline.OverrideEmbeddingsAPI = func(texts []string) ([][]float64, error) {
		return [][]float64{{0.1}}, nil
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not-json"))
	}))
	defer server.Close()

	addr := server.Listener.Addr().(*net.TCPAddr)
	req := ReqParams{
		Host: addr.IP.String(), Port: addr.Port,
		Tenant: "t", Database: "d", Collection_id: "c",
	}

	code, err, _ := ListCollections(req, []string{"text"})
	if err == nil {
		t.Error("Expected JSON parsing error")
	}
	if code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", code)
	}
}
