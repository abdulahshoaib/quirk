<<<<<<< Updated upstream
package pipeline

import (
	"reflect"
	"testing"
)

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
func TestPdfToText_Empty(t *testing.T) {
	// empty PDF byte content should fail gracefully
	_, err := PdfToText([]byte{})
	if err == nil {
		t.Error("Expected error from PdfToText when passing empty input, got nil")
	}
=======
// pipeline_test.go
package pipeline

import (
	"testing"
)

// MockEmbeddingsAPI mocks the embedding API
func MockEmbeddingsAPI(texts []string) ([][]float64, error) {
	mockEmbeddings := [][]float64{
		{0.1, 0.2, 0.3},
		{0.4, 0.5, 0.6},
	}
	return mockEmbeddings, nil
}

// Mock writer to capture output
func mockWriter(object_id string, embeddings [][]float64, triples []string) {
	if object_id != "test-object" {
		panic("Object ID mismatch")
	}
	if len(embeddings) != 2 {
		panic("Expected 2 embeddings")
	}
}

func TestProcessFiles(t *testing.T) {
	// Override the EmbeddingsAPI with the mock
	OverrideEmbeddingsAPI = MockEmbeddingsAPI

	// Prepare dummy file data
	memFiles := map[string][]byte{
		"file1.txt": []byte("This is an example text."),
		"file2.txt": []byte("Another example's text."),
	}

	// Run the function
	ProcessFiles("test-object", memFiles, mockWriter)
>>>>>>> Stashed changes
}
