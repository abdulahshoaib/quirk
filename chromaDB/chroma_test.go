package chromadb

import (
	"encoding/json"
	"fmt"
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

	t.Run("non-200 status code", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
		defer server.Close()

		req := ReqParams{
			Host:     server.Listener.Addr().(*net.TCPAddr).IP.String(),
			Port:     server.Listener.Addr().(*net.TCPAddr).Port,
			Tenant:   "test_tenant",
			Database: "test_db",
		}

		code, err := CheckHealth(req)
		if err == nil || code != http.StatusServiceUnavailable {
			t.Errorf("expected 503 error, got %d, err: %v", code, err)
		}
	})

	t.Run("http get failure", func(t *testing.T) {
		req := ReqParams{
			Host:     "127.0.0.1",
			Port:     0, // Invalid port to simulate failure
			Tenant:   "test_tenant",
			Database: "test_db",
		}

		code, err := CheckHealth(req)
		if err == nil || code != http.StatusInternalServerError {
			t.Errorf("expected internal server error, got %d, err: %v", code, err)
		}
	})

	req := ReqParams{
		Host:     server.Listener.Addr().(*net.TCPAddr).IP.String(),
		Port:     server.Listener.Addr().(*net.TCPAddr).Port,
		Tenant:   "test_tenant",
		Database: "test_db",
	}

	t.Run("malformed URL", func(t *testing.T) {
		req := ReqParams{
			Host:     "", // Missing host
			Port:     8080,
			Tenant:   "t",
			Database: "d",
		}

		_, err := CheckHealth(req)
		if err == nil {
			t.Error("expected error due to malformed URL, got nil")
		}
	})

	code, err := CheckHealth(req)
	if err != nil || code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d, err: %v", code, err)
	}
}

func TestCreateNewCollection(t *testing.T) {
	makePayload := func() Payload {
		return Payload{
			Embeddings: [][]float64{{0.1, 0.2}},
			Documents:  []string{"doc"},
			IDs:        []string{"id1"},
		}
	}

	t.Run("successful request", func(t *testing.T) {
		var received Payload

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
				t.Errorf("failed to decode payload: %v", err)
			}
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

		payload := makePayload()

		code, err := CreateNewCollection(req, payload)
		if err != nil || code != http.StatusOK {
			t.Errorf("CreateNewCollection failed: code=%d, err=%v", code, err)
		}
	})

	t.Run("http post failure", func(t *testing.T) {
		req := ReqParams{
			Host:          "127.0.0.1",
			Port:          0, // invalid port
			Tenant:        "t",
			Database:      "d",
			Collection_id: "c",
		}

		payload := makePayload()

		code, err := CreateNewCollection(req, payload)
		if err == nil || code != http.StatusInternalServerError {
			t.Errorf("expected http post error, got code=%d, err=%v", code, err)
		}
	})

	t.Run("read response failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hj, ok := w.(http.Hijacker)
			if ok {
				conn, _, _ := hj.Hijack()
				conn.Close() // Close connection before writing response
			}
		}))
		defer server.Close()

		req := ReqParams{
			Host:          server.Listener.Addr().(*net.TCPAddr).IP.String(),
			Port:          server.Listener.Addr().(*net.TCPAddr).Port,
			Tenant:        "t",
			Database:      "d",
			Collection_id: "c",
		}

		payload := makePayload()

		code, err := CreateNewCollection(req, payload)
		if err == nil || code != http.StatusInternalServerError {
			t.Errorf("expected response read error, got code=%d, err=%v", code, err)
		}
	})
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
	// Backup and restore EmbeddingFn
	original := pipeline.EmbeddingFn
	defer func() { pipeline.EmbeddingFn = original }()

	t.Run("success", func(t *testing.T) {
		pipeline.EmbeddingFn = func(texts []string) ([][]float64, error) {
			return [][]float64{{0.1, 0.2, 0.3}}, nil
		}

		mockResponse := ChromaQueryResponse{
			Documents: [][]string{{"doc1"}},
			Distances: [][]float64{{0.123}},
			IDs:       [][]string{{"id1"}},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(mockResponse)
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

		code, err, parsed := ListCollections(req, []string{"test query"})
		if err != nil || code != http.StatusOK || parsed == nil || len(parsed.Documents) != 1 {
			t.Errorf("expected success, got code=%d, err=%v", code, err)
		}
	})

	t.Run("embedding API fails", func(t *testing.T) {
		pipeline.EmbeddingFn = func(_ []string) ([][]float64, error) {
			return nil, fmt.Errorf("embedding error")
		}

		req := ReqParams{}
		code, err, parsed := ListCollections(req, []string{"fail"})
		if err == nil || code != http.StatusInternalServerError || parsed != nil {
			t.Errorf("expected embedding failure, got code=%d, err=%v", code, err)
		}
	})

	t.Run("empty embeddings", func(t *testing.T) {
		pipeline.EmbeddingFn = func(_ []string) ([][]float64, error) {
			return [][]float64{}, nil
		}

		req := ReqParams{}
		code, err, parsed := ListCollections(req, []string{"empty"})
		if err == nil || code != http.StatusBadRequest || parsed != nil {
			t.Errorf("expected empty embedding error, got code=%d, err=%v", code, err)
		}
	})

	t.Run("http post fails", func(t *testing.T) {
		pipeline.EmbeddingFn = func(_ []string) ([][]float64, error) {
			return [][]float64{{0.1}}, nil
		}

		req := ReqParams{
			Host:          "127.0.0.1",
			Port:          0, // invalid port
			Tenant:        "t",
			Database:      "d",
			Collection_id: "c",
		}

		code, err, parsed := ListCollections(req, []string{"fail"})
		if err == nil || code != http.StatusInternalServerError || parsed != nil {
			t.Errorf("expected http post failure, got code=%d, err=%v", code, err)
		}
	})

	t.Run("http returns error status", func(t *testing.T) {
		pipeline.EmbeddingFn = func(_ []string) ([][]float64, error) {
			return [][]float64{{0.1}}, nil
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "server error", http.StatusInternalServerError)
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

		code, err, parsed := ListCollections(req, []string{"fail"})
		if err == nil || code != http.StatusInternalServerError || parsed != nil {
			t.Errorf("expected server error, got code=%d, err=%v", code, err)
		}
	})

	t.Run("invalid JSON response", func(t *testing.T) {
		pipeline.EmbeddingFn = func(_ []string) ([][]float64, error) {
			return [][]float64{{0.1}}, nil
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "not a json")
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

		code, err, parsed := ListCollections(req, []string{"fail"})
		if err == nil || code != http.StatusInternalServerError || parsed != nil {
			t.Errorf("expected JSON unmarshal error, got code=%d, err=%v", code, err)
		}
	})
}
