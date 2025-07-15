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
}
