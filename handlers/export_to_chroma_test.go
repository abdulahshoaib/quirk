package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	chromadb "github.com/abdulahshoaib/quirk/chromaDB"
	"github.com/jarcoal/httpmock"
)

func TestHandleExportToChroma_AddSuccess(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Mock the outgoing ChromaDB POST request
	httpmock.RegisterResponder("POST", "http://localhost:8001/api/v2/tenants/quirk/databases/quirk/collections/afaa5b03-e179-4afd-bc7a-6948daf7056b/add",
		httpmock.NewStringResponder(200, `{"message": "success"}`))

	// Prepare mock jobResults
	id := "test_id"
	embedding := make([]float64, 1024)
	for i := range embedding {
		embedding[i] = float64(i) * 0.01
	}
	jobResults[id] = Result{
		Embeddings:  [][]float64{embedding},
		Filenames:   []string{"file1"},
		Filecontent: []string{"This is the file content"},
	}
	defer delete(jobResults, id)

	// Prepare request
	body := struct {
		Req     chromadb.ReqParams `json:"req"`
		Payload chromadb.Payload   `json:"payload"`
	}{
		Req: chromadb.ReqParams{
			Host:          "localhost",
			Port:          8001,
			Tenant:        "quirk",
			Database:      "quirk",
			Collection_id: "afaa5b03-e179-4afd-bc7a-6948daf7056b",
		},
		Payload: chromadb.Payload{},
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/export?object_id="+id+"&operation=add", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call handler
	HandleExportToChroma(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}
}

func TestHandleExportToChroma_MissingID(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/export?operation=add", nil)
	w := httptest.NewRecorder()

	HandleExportToChroma(w, req)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 BadRequest, got %d", w.Result().StatusCode)
	}
}

func TestHandleExportToChroma_InvalidOperation(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/export?object_id=test_id&operation=delete", nil)
	w := httptest.NewRecorder()

	HandleExportToChroma(w, req)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 BadRequest, got %d", w.Result().StatusCode)
	}
}

func TestHandleExportToChroma_BadJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/export?object_id=test_id&operation=add", bytes.NewReader([]byte("{bad json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	HandleExportToChroma(w, req)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 BadRequest, got %d", w.Result().StatusCode)
	}
}

func TestHandleExportToChroma_EmbeddingNotFound(t *testing.T) {
	reqBody := `{"req": {}, "payload": {}}`
	req := httptest.NewRequest(http.MethodPost, "/export?object_id=unknown_id&operation=add", bytes.NewReader([]byte(reqBody)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	HandleExportToChroma(w, req)

	if w.Result().StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 NotFound, got %d", w.Result().StatusCode)
	}
}
