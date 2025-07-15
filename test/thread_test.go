package test

import (
	"fmt"
	"math/rand/v2"
	"sync"
	"testing"
	"time"

	"github.com/abdulahshoaib/quirk/pipeline"
)

func TestThreading(t *testing.T) {
	originalAPI := pipeline.OverrideEmbeddingsAPI
	defer func() { pipeline.OverrideEmbeddingsAPI = originalAPI }()

	pipeline.OverrideEmbeddingsAPI = func(texts []string) ([][]float64, error) {
		// mock latency
		time.Sleep(time.Duration(rand.IntN(10)) * time.Millisecond)
		return [][]float64{{0.1, 0.2, 0.3}}, nil
	}

	memFiles := map[string][]byte{}
	for i := 1; i <= 100; i++ {
		filename := fmt.Sprintf("file%d.txt", i)
		content := []byte(fmt.Sprintf("content of file %d", i))
		memFiles[filename] = content
	}

	var mu sync.Mutex
	var capturedID string
	var capturedEmbs [][]float64
	var capturedTrips []string

	done := make(chan bool, 1)

	writer := func(id string, embs [][]float64, trips []string) {
		mu.Lock()
		capturedID = id
		capturedEmbs = embs
		capturedTrips = trips
		mu.Unlock()
		done <- true
	}

	go pipeline.ProcessFiles("abc123", memFiles, writer)

	select {
	case <-done:
		mu.Lock()
		defer mu.Unlock()
		if capturedID != "abc123" {
			t.Errorf("expected object_id 'abc123', got %s", capturedID)
		}
		if len(capturedEmbs) != 100 {
			t.Errorf("expected 2 embeddings, got %d", len(capturedEmbs))
		}
		if len(capturedTrips) != 100 {
			t.Errorf("expected 2 triples, got %d", len(capturedTrips))
		}
	case <-time.After(10 * time.Second):
		t.Fatal("timeout waiting for goroutines to finish")
	}
}

func TestThreading_ZeroFiles(t *testing.T) {
	memFiles := map[string][]byte{}

	var called bool
	writer := func(id string, embs [][]float64, trips []string) {
		called = true
	}

	pipeline.ProcessFiles("empty", memFiles, writer)

	if !called {
		t.Error("writeBack not called on empty input")
	}
}

func TestEmbeddingAPIReturnsHTML(t *testing.T) {
	originalAPI := pipeline.OverrideEmbeddingsAPI
	defer func() { pipeline.OverrideEmbeddingsAPI = originalAPI }()

	pipeline.OverrideEmbeddingsAPI = func(texts []string) ([][]float64, error) {
		return nil, fmt.Errorf("<html><body>500 Internal Server Error</body></html>")
	}

	memFiles := map[string][]byte{
		"file1.txt": []byte("hello"),
	}

	called := false

	writer := func(id string, embs [][]float64, trips []string) {
		called = true

		if len(embs) != 0 {
			t.Errorf("expected 0 embeddings due to error, got %d", len(embs))
		}
		if len(trips) != 1 {
			t.Errorf("expected 1 triple (even if blank), got %d", len(trips))
		}
	}

	pipeline.ProcessFiles("html-test", memFiles, writer)

	if !called {
		t.Error("writeBack not called when HTML error occurred")
	}
}
