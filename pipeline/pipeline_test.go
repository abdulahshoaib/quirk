package pipeline

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
)

// ------------------------------------
// ------------------------------------
// ----- Testing File Processing ------
// ------------------------------------
// ------------------------------------

// mock EmbeddingsAPI override
func mockEmbeddingsAPI(texts []string) ([][]float64, error) {
	return [][]float64{
		{0.1, 0.2, 0.3},
		{0.4, 0.5, 0.6},
	}, nil
}

// mock ResultWriter to capture results
type captureResult struct {
	objectID   string
	embeddings [][]float64
	triples    []string
}
func TestProcessFiles(t *testing.T) {
	// override API
	OverrideEmbeddingsAPI = mockEmbeddingsAPI

	memFiles := map[string][]byte{
		"file1.txt": []byte("Hello, this is test 1.\nIt's a file."),
		"file2.txt": []byte("Second file's content."),
	}

	var captured captureResult

	writeBack := func(object_id string, embeddings [][]float64, triples []string) {
		captured = captureResult{object_id, embeddings, triples}
	}

	ProcessFiles("obj123", memFiles, writeBack)

	if captured.objectID != "obj123" {
		t.Errorf("Expected object ID 'obj123', got '%s'", captured.objectID)
	}

	expectedEmbeddings := [][]float64{
		{0.1, 0.2, 0.3},
		{0.4, 0.5, 0.6},
	}

	if !reflect.DeepEqual(captured.embeddings, expectedEmbeddings) {
		t.Errorf("Embeddings mismatch.\nExpected: %v\nGot: %v", expectedEmbeddings, captured.embeddings)
	}

	if len(captured.triples) != 0 {
		t.Errorf("Expected no triples, got: %v", captured.triples)
	}
}

// ------------------------------------
// ------------------------------------
// ------- Testing CSV => Text --------
// ------------------------------------
// ------------------------------------
func TestCsvToText(t *testing.T) {
	input := []byte("name,age\nEesa,21")
	expected := "name\tage\nEesa\t21\n"

	output, err := CsvToText(input)
	if err != nil {
		t.Errorf("CsvToText returned error: %v", err)
	}
	if string(output) != expected {
		t.Errorf("CsvToText output mismatch.\nExpected:\n%s\nGot:\n%s", expected, output)
	}
}
func TestCsvToText_Empty(t *testing.T) {
	input := []byte("")

	output, err := CsvToText(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := ""
	if string(output) != expected {
		t.Errorf("expected %q, got %q", expected, string(output))
	}
}
func TestCsvToText_Malformed(t *testing.T) {
	input := []byte(`name,age\n"Eesa,21`) // unclosed quote

	_, err := CsvToText(input)
	if err == nil {
		t.Fatal("expected error for malformed CSV, got nil")
	}
}
func TestCsvToText_UnevenRows(t *testing.T) {
	input := []byte("name,age\nEesa")
	expected := "name\tage\nEesa\n"

	output, err := CsvToText(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(output) != expected {
		t.Errorf("expected:\n%q\ngot:\n%q", expected, output)
	}
}

// -------------------------------------
// -------------------------------------
// ---- Testing PDF => Text (empty) ----
// -------------------------------------
// -------------------------------------
func TestPdfToText(t *testing.T) {
	pdfData, err := os.ReadFile("../test-data/week9.pdf")

	text, err := PdfToText(pdfData)
	if err != nil {
		t.Fatalf("PdfToText failed: %v", err)
	}

	result := string(text)
	expectedSubstring := "departments given"

	if !strings.Contains(result, expectedSubstring) {
		t.Errorf("Expected text to contain %q, got %q", expectedSubstring, result)
	}
}
func TestPdfToText_Empty(t *testing.T) {
	_, err := PdfToText([]byte{})
	if err == nil {
		t.Error("Expected error from PdfToText when passing empty input, got nil")
	}
}
func TestPdfToText_InvalidPDF(t *testing.T) {
	_, err := PdfToText([]byte("%not-a-real-pdf"))
	if err == nil {
		t.Error("expected error when reading invalid PDF, got nil")
	}
}

// --------------------------------
// --------------------------------
// --- Testing Backend API res ----
// --------------------------------
// --------------------------------
func TestEmbeddingsAPI(t *testing.T) {
	os.Setenv("CLOUDFLARE_ACCOUNT_ID", "testid")
	os.Setenv("CLOUDFLARE_API_TOKEN", "testtoken")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req BGEReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if len(req.Text) != 2 || req.Text[0] != "hello" {
			t.Errorf("unexpected input: %v", req.Text)
		}

		response := BGERes{}
		response.Result.Data = [][]float64{
			{0.1, 0.2, 0.3},
			{0.4, 0.5, 0.6},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	EmbeddingsAPIURL = server.URL + "/%s"

	result, err := EmbeddingsAPI([]string{"hello", "world"})
	if err != nil {
		t.Fatalf("EmbeddingsAPI failed: %v", err)
	}

	expected := [][]float64{
		{0.1, 0.2, 0.3},
		{0.4, 0.5, 0.6},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestEmbeddingsAPI_MissingEnvVars(t *testing.T) {
	os.Unsetenv("CLOUDFLARE_ACCOUNT_ID")
	os.Unsetenv("CLOUDFLARE_API_TOKEN")

	_, err := EmbeddingsAPI([]string{"test"})
	if err == nil || !strings.Contains(err.Error(), "missing CLOUDFLARE_ACC") {
		t.Errorf("Expected error for missing env vars, got: %v", err)
	}
}

func TestEmbeddingsAPI_BadJSONResponse(t *testing.T) {
	os.Setenv("CLOUDFLARE_ACCOUNT_ID", "id")
	os.Setenv("CLOUDFLARE_API_TOKEN", "token")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	EmbeddingsAPIURL = server.URL + "/%s"

	_, err := EmbeddingsAPI([]string{"x"})
	if err == nil {
		t.Fatal("Expected JSON decode error, got nil")
	}
}

func TestEmbeddingsAPI_RequestError(t *testing.T) {
	os.Setenv("CLOUDFLARE_ACCOUNT_ID", "id")
	os.Setenv("CLOUDFLARE_API_TOKEN", "token")

	// Point to an invalid server (e.g., closed port)
	EmbeddingsAPIURL = "http://localhost:12345/%s"

	_, err := EmbeddingsAPI([]string{"fail"})
	if err == nil {
		t.Fatal("Expected request error, got nil")
	}
}

// ---- Testing JSON => Text ----
func TestJsonToText(t *testing.T) {
	input := []byte(`{"name":"Eesa","age":21}`)
	expected := "{\n  \"name\": \"Eesa\",\n  \"age\": 21\n}"

	output, err := JsonToText(input)
	if err != nil {
		t.Errorf("JsonToText returned error: %v", err)
	}
	if string(output) != expected {
		t.Errorf("JsonToText output mismatch.\nExpected:\n%s\nGot:\n%s", expected, output)
	}
}

func TestJsonToText_Empty(t *testing.T) {
	input := []byte("")

	result, err := JsonToText(input)

	if err == nil {
		t.Fatal("expected error for empty JSON, got nil")
	}

	if result != nil {
		t.Errorf("expected empty result, got: %s", string(result))
	}
}
