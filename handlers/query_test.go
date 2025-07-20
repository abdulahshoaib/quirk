package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/abdulahshoaib/quirk/pipeline"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func mockEmbeddingsAPI(texts []string) ([][]float64, error) {
	return [][]float64{{1.1, 2.2, 3.3}}, nil
}

func TestHandleQuery_Success(t *testing.T) {
	// Override embedding function
	pipeline.EmbeddingFn = mockEmbeddingsAPI

	// Activate HTTP mocking
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	mockResp := map[string]any{
		"documents": [][]string{{"doc1"}},
		"distances": [][]float64{{0.5}},
	}
	mockRespBytes, _ := json.Marshal(mockResp)

	// Match any outgoing HTTP POST to chroma
	httpmock.RegisterResponder("POST", "http://localhost:8000/api/v2/tenants/t1/databases/db1/collections/col1/query",
		httpmock.NewBytesResponder(200, mockRespBytes))

	// Build request
	payload := map[string]any{
		"req": map[string]any{
			"Host":          "localhost",
			"Port":          8000,
			"Tenant":        "t1",
			"Database":      "db1",
			"Collection_id": "col1",
		},
		"text": []string{"what is ai"},
	}
	jsonBytes, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/query", bytes.NewReader(jsonBytes))
	w := httptest.NewRecorder()

	// Call the actual handler
	HandleQuery(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var parsed map[string]any
	body, _ := io.ReadAll(resp.Body)
	err := json.Unmarshal(body, &parsed)
	assert.NoError(t, err)

	// check returned content
	assert.Contains(t, parsed, "documents")
	assert.Contains(t, parsed, "distances")
}
