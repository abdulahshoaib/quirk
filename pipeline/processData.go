package pipeline

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
	"sync"
	"encoding/json"
	"encoding/csv"
	"strings"

	"rsc.io/pdf"

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

			re := regexp.MustCompile(`['\n]`)
			raw := string(data)

			// removing ' (apostrophe) and '\n'
			cleaned := re.ReplaceAllString(raw, "")
			corpus := stopwords.CleanString(cleaned, "en", true)
			log.Println(corpus)

			mutex.Lock()
			corpusCleaned = append(corpusCleaned, corpus)
			mutex.Unlock()
		}(fname, content)
	}
	wg.Wait()
	log.Printf("Processed job %s: %d files", object_id, len(memFiles))

	log.Printf("Created Tokens -> sending to API")

	embeddings, err := EmbeddingsAPI(corpusCleaned)
	if err != nil {
		log.Printf("embedding failed: %v", err)
	} else {
		log.Printf("Received %d embedding of length: %d", len(embeddings), len(embeddings[0]))
	}

	writeBack(object_id, embeddings, trips)
}

func PdfToText(content []byte) ([]byte, error) {
	reader := bytes.NewReader(content)
	pdfReader, err := pdf.NewReader(reader, int64(len(content)))
	if err != nil {
		return nil, fmt.Errorf("Failed to create PDF reader: %v", err)
	}

	var text string

	numPages := pdfReader.NumPage()
	for i := 1; i <= numPages; i++ {
		page := pdfReader.Page(i)

		content := page.Content()
		for _, textObj := range content.Text {
			text += textObj.S
		}
	}
	return []byte(text), nil
}

func JsonToText(content []byte) ([]byte, error) {
	var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, content, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}
	return prettyJSON.Bytes(), nil
}

func CsvToText(content []byte) ([]byte, error) {
	r := csv.NewReader(bytes.NewReader(content))
	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %v", err)
	}

	var builder strings.Builder
	for _, record := range records {
		builder.WriteString(strings.Join(record, "\t")) // Tab-separated
		builder.WriteString("\n")
	}

	return []byte(builder.String()), nil
}
