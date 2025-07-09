package pipeline

import (
	"log"
	"regexp"
	"strings"
	"sync"

	"github.com/bbalet/stopwords"
)

var OverrideEmbeddingsAPI = EmbeddingsAPI

func ProcessFiles(object_id string, memFiles map[string][]byte, writeBack ResultWriter) {
	var wg = sync.WaitGroup{}
	var embs [][]float64
	var trips []string
	var corpusCleaned []string
	var mutex = sync.Mutex{}

	for fname, content := range memFiles {
		wg.Add(1)
		go func(fname string, data []byte) {
			defer wg.Done()

			re := regexp.MustCompile(`'`)
			raw := string(data)

			// removing ' (apostrophe)
			cleaned := re.ReplaceAllString(raw, "")
			corpus := stopwords.CleanString(cleaned, "en", true)

			mutex.Lock()
			corpusCleaned = append(corpusCleaned, corpus)
			mutex.Unlock()
		}(fname, content)
	}
	wg.Wait()
	log.Printf("Processed job %s: %d files", object_id, len(memFiles))

	// tokenized to send corpusData
	var tokens []string
	for _, corp := range corpusCleaned {
		tok := strings.Fields(corp)
		tokens = append(tokens, tok...)
	}

	log.Printf("Created Tokens -> sending to API")

	embeddings, err := EmbeddingsAPI(tokens)
	if err != nil {
		log.Printf("embedding failed: %v", err)
	} else {
		log.Printf("Received embedding of length: %d", len(embeddings[0]))
	}

	for _, emb := range embeddings {
		embs = append(embs, emb)
	}

	writeBack(object_id, embs, trips)
}
