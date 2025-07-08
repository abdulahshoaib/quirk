package pipeline

import (
	"log"
	"sync"
)

type ResultWriter func(object_id string, embeddings, triples []string)

func ProcessFiles(object_id string, memFiles map[string][]byte, writeBack ResultWriter)  {
		var wg = sync.WaitGroup{}
		var embs []string
		var trips []string
		var mutex = sync.Mutex{}

		for name, content := range memFiles {
			wg.Add(1)
			go func(name string, data []byte) {
				defer wg.Done()

				// embedding logic
				embeddings := ""
				// triple logic
				triple := ""

				mutex.Lock()
				embs = append(embs, embeddings)
				trips = append(trips, triple)
				mutex.Unlock()
			}(name, content)
		}
		wg.Wait()
		log.Printf("Processed job %s: %d files", object_id, len(memFiles))

		writeBack(object_id, embs, trips)
}
