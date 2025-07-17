package handlers

import (
	"bytes"
	"encoding/json"
	chromadb "github.com/abdulahshoaib/quirk/chromaDB"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleExportToChroma_AddSuccess(t *testing.T) {
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

	body := struct {
		Req     chromadb.ReqParams `json:"req"`
		Payload chromadb.Payload   `json:"payload"`
	}{
		Req: chromadb.ReqParams{
			Host:          "localhost",
			Port:          8001,
			Tenant:        "quirk",
			Database:      "quirk",
			Collection_id: "53fe1c09-3202-482a-bbe5-590cbd7ba0cc",
		},
		Payload: chromadb.Payload{},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/export?object_id="+id+"&operation=add", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	HandleExportToChroma(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", resp.StatusCode)
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
