package chromadb

import (
	"encoding/json"
	// "fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckHealth(t *testing.T) {
	// Start a test server with 200 OK
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/tenants/test_tenant/databases/test_db/heartbeat" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	req := ReqParams{
		Host:     "localhost",
		Port:     80, // placeholder; will be ignored in test
		Tenant:   "test_tenant",
		Database: "test_db",
	}

	// Override the actual host and port to use our test server
	req.Host = ts.URL[len("http://"):]
	status, err := CheckHealth(req)
	if err != nil || status != http.StatusOK {
		t.Errorf("expected 200 OK, got %d, err: %v", status, err)
	}
}

func TestCreateNewCollection(t *testing.T) {
	payload := Payload{
		IDs:       []string{"id1"},
		Documents: []string{"Test doc"},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var p Payload
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			t.Errorf("failed to decode body: %v", err)
		}
		if p.IDs[0] != "id1" {
			t.Errorf("unexpected ID: %s", p.IDs[0])
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	req := ReqParams{
		Host:          ts.URL[len("http://"):],
		Port:          80,
		Tenant:        "tenant",
		Database:      "db",
		Collection_id: "collection",
	}

	status, err := CreateNewCollection(req, payload)
	if err != nil || status != http.StatusOK {
		t.Errorf("expected 200 OK, got %d, err: %v", status, err)
	}
}

func TestUpdateCollection(t *testing.T) {
	payload := Payload{
		IDs:       []string{"id2"},
		Documents: []string{"Updated doc"},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var p Payload
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			t.Errorf("failed to decode body: %v", err)
		}
		if p.IDs[0] != "id2" {
			t.Errorf("unexpected ID: %s", p.IDs[0])
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	req := ReqParams{
		Host:          ts.URL[len("http://"):],
		Port:          80,
		Tenant:        "tenant",
		Database:      "db",
		Collection_id: "collection",
	}

	status, err := UpdateCollection(req, payload)
	if err != nil || status != http.StatusOK {
		t.Errorf("expected 200 OK, got %d, err: %v", status, err)
	}
}

// Skipping ListCollections due to external dependency on pipeline.EmbeddingsAPI
func TestListCollections(t *testing.T) {
	t.Skip("ListCollections requires mocking pipeline.EmbeddingsAPI")
}
