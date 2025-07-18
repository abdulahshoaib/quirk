package pipeline

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"sync"

	"github.com/bbalet/stopwords"
	"github.com/ledongthuc/pdf"
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
			slog.Debug("cleaned corpus", slog.String("filename", fname), slog.String("corpus", corpus))

			mutex.Lock()
			corpusCleaned = append(corpusCleaned, corpus)
			mutex.Unlock()
		}(fname, content)
	}
	wg.Wait()
	slog.Info("processed job", slog.String("object_id", object_id), slog.Int("file_count", len(memFiles)))

	slog.Info("created tokens, sending to embedding API", slog.String("object_id", object_id))

	// for testing
	// embeddings, err := OverrideEmbeddingsAPI(corpusCleaned)

	// for production
	embeddings, err := EmbeddingsAPI(corpusCleaned)

	if err != nil {
		slog.Error("embedding API call failed", slog.String("object_id", object_id), slog.Any("error", err))
	} else {
		slog.Info("received embeddings", slog.String("object_id", object_id), slog.Int("count", len(embeddings)), slog.Int("embedding_size", len(embeddings[0])))
	}

	writeBack(object_id, embeddings, trips)
}

func PdfToText(content []byte) ([]byte, error) {
	reader := bytes.NewReader(content)
	pdfReader, err := pdf.NewReader(reader, int64(len(content)))
	if err != nil {
		return nil, fmt.Errorf("failed to create PDF reader: %v", err)
	}

	var builder strings.Builder
	numPages := pdfReader.NumPage()

	for i := 1; i <= numPages; i++ {
		page := pdfReader.Page(i)
		if page.V.IsNull() { // make sure the page exists
			continue
		}

		texts, err := page.GetPlainText(nil)
		if err != nil {
			slog.Error("failed to extract text from PDF page", slog.Int("page", i), slog.Any("error", err))
			continue
		}

		builder.WriteString(texts)
	}

	result := builder.String()
	return []byte(result), nil
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
	r.FieldsPerRecord = -1
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
