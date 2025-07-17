package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/abdulahshoaib/quirk/pipeline"
	"github.com/google/uuid"
)

// For testing
var ReadAll = io.ReadAll

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With")
	(*w).Header().Set("Access-Control-Allow-Credentials", "true")
}

// HandleProcess handles the uploading of multiple files, starts asynchronous
// processing using the embedding model, and returns a unique object ID.
//
// POST /process
//
// Description:
//   - Accepts multipart form uploads under the field "files".
//   - Reads all uploaded files into memory.
//   - Triggers asynchronous processing for embeddings and triples.
//   - Tracks job status using an internal job ID.
//
// Request:
//   - Content-Type: multipart/form-data
//   - Form Field: files (one or more files)
//
// Returns:
//   - 200: JSON object with { "object_id": string }
//   - 400: If no files are uploaded or request is malformed
//   - 405: If method is not POST
//
// Example:
//
//	POST /process
//	Form field "files": [file1.txt, file2.txt]
//
// Response:
//
//	{ "object_id": "e45c1c20-b7a1-4d65-b7e1-a9fa73c0e217" }
//
// The returned object ID can be used to:
//   - Check status via /status?object_id=...
//   - Get results via /results?object_id=...
//   - Export results via /export?object_id=...&format=csv|json
func HandleProcess(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	if r.Method != http.MethodPost {
		log.Fatal("Method not allowed")
		http.Error(w, "[POST] allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(50 << 20) // 50MB
	if err != nil {
		log.Fatalf("Failed to parse: %v", err.Error())
		http.Error(w, "Failed to parse:"+err.Error(), http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		log.Fatal("No files uploaded")
		http.Error(w, "No files uploaded", http.StatusBadRequest)
		return
	}

	object_id := uuid.NewString()
	memFiles := map[string][]byte{}
	filenames := []string{}
	filecontent := []string{}

	for _, fh := range files {
		file, err := fh.Open()
		if err != nil {
			log.Fatalf("File open error: %v", err.Error())
			http.Error(w, "File open error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer file.Close()

		contentBytes, err := ReadAll(file)
		if err != nil {
			log.Fatalf("Read error: %v", err.Error())
			http.Error(w, "Read error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		fileExt := strings.ToLower(filepath.Ext(fh.Filename))
		if fileExt != ".pdf" &&
			fileExt != ".csv" &&
			fileExt != ".txt" &&
			fileExt != ".json" &&
			fileExt != ".md" &&
			fileExt != ".yml" &&
			fileExt != ".xml" {
			log.Fatalf("Unsupported file extension: %v", err.Error())
			http.Error(w, "Unsupported file extension: "+fileExt, http.StatusBadRequest)
			return
		}

		contentType := http.DetectContentType(contentBytes)
		if idx := strings.Index(contentType, ";"); idx != -1 {
			contentType = strings.TrimSpace(contentType[:idx])
		}

		var textBytes []byte

		switch contentType {
		case "application/pdf":
			textBytes, err = pipeline.PdfToText(contentBytes)
		case "text/csv":
			textBytes, err = pipeline.CsvToText(contentBytes)
		case "application/json":
			textBytes, err = pipeline.JsonToText(contentBytes)
		case "text/plain", "text/markdown", "text/x-log", "text/x-yaml", "text/x-markdown":
			textBytes = contentBytes
		case "application/xml", "text/xml":
			textBytes = contentBytes
		default:
			log.Fatalf("Unsupported file type: %s", contentType)
			http.Error(w, "Unsupported file type: "+contentType, http.StatusBadRequest)
			return
		}

		if err != nil {
			log.Fatal("Failed to process file: %v", err.Error())
			http.Error(w, "Failed to process file: "+err.Error(), http.StatusInternalServerError)
			return
		}

		memFiles[fh.Filename] = textBytes
		filenames = append(filenames, fh.Filename)
		filecontent = append(filecontent, string(textBytes))

		log.Printf("Processed file: %s (%d bytes)", fh.Filename, len(contentBytes))
	}

	mutex.Lock()
	jobStatuses[object_id] = JobStatus{
		Status: "processing",
		ETA:    time.Now().Add(5 * time.Second),
		Error:  "",
	}
	mutex.Unlock()

	// asynchronusly writes back whenever the embeddings are created
	go pipeline.ProcessFiles(object_id, memFiles, func(id string, embs [][]float64, trips []string) {
		mutex.Lock()
		jobResults[id] = Result{
			Embeddings:  embs,
			Triples:     trips,
			Filenames:   filenames,
			Filecontent: filecontent,
		}
		jobStatuses[id] = JobStatus{
			Status: "completed",
			ETA:    time.Time{},
			Error:  "",
		}
		mutex.Unlock()
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"object_id": object_id})
}
