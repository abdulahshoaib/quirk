package chromadb

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
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
