package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/abdulahshoaib/quirk/pipeline"
)

func createMultipartRequest(t *testing.T, fieldName, filename, contentType, content string) *http.Request {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile(fieldName, filename)
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	part.Write([]byte(content))

	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/process", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func TestHandleProcess_SuccessTxt(t *testing.T) {
	req := createMultipartRequest(t, "files", "example.txt", "text/plain", "Hello world!")
	w := httptest.NewRecorder()

	HandleProcess(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", res.StatusCode)
	}

	var body map[string]string
	err := json.NewDecoder(res.Body).Decode(&body)
	if err != nil {
		t.Fatalf("failed to parse JSON response: %v", err)
	}

	if _, ok := body["object_id"]; !ok {
		t.Error("expected object_id in response")
	}
}

func TestHandleProcess_NoFiles(t *testing.T) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	w.Close()

	req := httptest.NewRequest(http.MethodPost, "/process", &b)
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp := httptest.NewRecorder()
	HandleProcess(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", resp.Code)
	}
}

func TestHandleProcess_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/process", nil)
	w := httptest.NewRecorder()

	HandleProcess(w, req)

	if w.Result().StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 Method Not Allowed, got %d", w.Result().StatusCode)
	}
}

func TestHandleProcess_UnsupportedFile(t *testing.T) {
	req := createMultipartRequest(t, "files", "image.png", "image/png", "fake image content")
	w := httptest.NewRecorder()

	HandleProcess(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for unsupported file type, got %d", w.Code)
	}
}

func TestHandleProcess_ReadError(t *testing.T)  {
	req := createMultipartRequest(t, "files", "test.txt", "text/plain", "")

	oldReadAll := func(r io.Reader) ([]byte, error){
		return nil, errors.New("forced read error")
	}
	defer func (){ pipeline.ReadAll = oldReadAll }()

}
