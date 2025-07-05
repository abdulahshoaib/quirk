package test

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/abdulahshoaib/quirk/handlers"
)

func TestHandleUpload(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	files := []string{"testdata/test.pdf", "testdata/test.csv"}

	for _, f := range files {
		part, err := writer.CreateFormFile("files", filepath.Base(f))
		if err != nil {
			t.Fatal(err)
		}
		file, err := os.Open(f)
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()
		io.Copy(part, file)
	}
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rr := httptest.NewRecorder()
	handlers.HandleUpload(rr, req)

	if rr.Code != http.StatusCreated && rr.Code != http.StatusOK {
		t.Errorf("expected status 200 or 201, got %d", rr.Code)
	}

	t.Log("response: ", rr.Body.String())

}
