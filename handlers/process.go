package handlers

import (
	"io"
	"log"
	"net/http"
	"os"

	"path/filepath"
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

	for _, fh := range files {
		file, err := fh.Open()
		if err != nil {
			http.Error(w, "File open error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer file.Close()

		log.Printf("Uploading file: %s (%d)", fh.Filename, fh.Size)

		dst, err := os.Create(filepath.Join("/tmp", fh.Filename))
		if err != nil {
			http.Error(w, "Save Error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		io.Copy(dst, file)
	}
}
