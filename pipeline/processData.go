package pipeline

import (
	"log"
	"sync"
)

func ProcessFiles(object_id string, memFiles map[string][]byte, writeBack ResultWriter) {
	var wg = sync.WaitGroup{}
	var embs [][]float64
	var trips []string
	var mutex = sync.Mutex{}

	for fname, content := range memFiles {
		wg.Add(1)
		go func(fname string, data []byte) {
			defer wg.Done()

			text := string(data)
			embeddings, err := embeddingsAPI([]string{text})
			if err != nil {
				log.Printf("embedding failed: %v", err)
			} else {
				log.Printf("Received embedding of length: %d", len(embeddings[0]))
			}

			triples := ""

			mutex.Lock()
			embs = append(embs, embeddings[0])
			trips = append(trips, triples)
			mutex.Unlock()

		}(fname, content)
	}
	wg.Wait()
	log.Printf("Processed job %s: %d files", object_id, len(memFiles))

	writeBack(object_id, embs, trips)
}
