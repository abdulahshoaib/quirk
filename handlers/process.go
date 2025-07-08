package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/abdulahshoaib/quirk/pipeline"
	"github.com/google/uuid"
)

// POST /process -> Upload files + start conversion → returns a unique object_id
func HandleProcess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "[POST] allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(50 << 20) // 50MB
	if err != nil {
		http.Error(w, "Failed to parse:"+err.Error(), http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		http.Error(w, "No files uploaded", http.StatusBadRequest)
		return
	}

	object_id := uuid.NewString()
	memFiles := map[string][]byte{}

	for _, fh := range files {
		file, err := fh.Open()
		if err != nil {
			http.Error(w, "File open error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer file.Close()

		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, file); err != nil {
			http.Error(w, "Read error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		memFiles[fh.Filename] = buf.Bytes()
		log.Printf("Uploading file: %s (%d)", fh.Filename, len(buf.Bytes()))
	}

	mutex.Lock()
	jobStatuses[object_id] = JobStatus{
		Status: "in_progress",
		ETA:    time.Now().Add(5 * time.Second),
		Error:  "",
	}
	mutex.Unlock()

	go pipeline.ProcessFiles(object_id, memFiles, func(id string, embs, trips []string) {
		mutex.Lock()
		jobResults[id] = Result{
			Embeddings: embs,
			Triples:    trips,
		}
		jobStatuses[id] = JobStatus{
			Status: "completed",
			ETA: time.Time{},
			Error: "",
		}
		mutex.Unlock()
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"object_id": object_id})

}
