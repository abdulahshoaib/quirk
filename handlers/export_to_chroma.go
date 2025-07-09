package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	chromadb "github.com/abdulahshoaib/quirk/chromaDB"
)

func HandleExportToChroma(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("object_id")
	operation := r.URL.Query().Get("operation")

	if id == "" {
		http.Error(w, "Missing object_id", http.StatusBadRequest)
		return
	}
	if operation == "" {
		http.Error(w, "operation missing", http.StatusBadRequest)
		return
	}

	if operation != "update" && operation != "add" {
		http.Error(w, "invalid operation param", http.StatusBadRequest)
		return
	}

	var (
		req     chromadb.ReqParams
		payload chromadb.Payload
	)

	var body struct {
		Req     chromadb.ReqParams `json:"req"`
		Payload chromadb.Payload   `json:"payload"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	results, ok := jobResults[id]
	if !ok {
		http.Error(w, "embedding not found for object_id", http.StatusNotFound)
		return
	}

	req = body.Req
	payload = body.Payload
	payload.Embeddings = results.Embeddings

	log.Printf("req %s", req)
	log.Printf("payload %s", payload)

	var (
		status int
		err    error
	)

	switch operation {
	case "update":
		status, err = chromadb.UpdateCollection(req, payload)
	case "add":
		status, err = chromadb.CreateNewCollection(req, payload)
	}

	if err != nil {
		log.Printf("Chroma operation failed: %v", err)
		http.Error(w, err.Error(), status)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Chroma operation succeeded"))
}
