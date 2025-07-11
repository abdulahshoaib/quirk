package pipeline

import (
	"log"
	"regexp"
	"sync"

	"github.com/bbalet/stopwords"
)

var OverrideEmbeddingsAPI = EmbeddingsAPI

func ProcessFiles(object_id string, memFiles map[string][]byte, writeBack ResultWriter) {
	var wg = sync.WaitGroup{}
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

	for _, corpus := range corpusCleaned {
		log.Println(corpus)
	}

	log.Printf("Created Tokens -> sending to API")

	embeddings, err := EmbeddingsAPI(corpusCleaned)
	if err != nil {
		log.Printf("embedding failed: %v", err)
	} else {
		log.Printf("Received %d embedding of length: %d", len(embeddings), len(embeddings[0]))
	}

	writeBack(object_id, embeddings, trips)
}
