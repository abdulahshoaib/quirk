package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	chromadb "github.com/abdulahshoaib/quirk/chromaDB"
)

// HandleExportToChroma handles exporting embeddings to a ChromaDB collection,
// either by adding a new collection or updating an existing one.
//
// POST /export?object_id={id}&operation={add|update}
//
// Query Parameters:
//   - object_id (required): Unique identifier corresponding to pre-computed embeddings
//   - operation (required): Operation type; must be either "add" or "update"
//
// Request Body (JSON):
//
//	{
//	  "req": { ... },       // chromadb.ReqParams object for Chroma configuration
//	  "payload": { ... }    // chromadb.Payload with metadata; embeddings will be injected
//	}
//
// Response Codes:
//   - 200 OK: Operation completed successfully
//   - 400 Bad Request: Missing or invalid parameters, or malformed JSON body
//   - 404 Not Found: No embeddings found for the given object_id
//   - 5xx Error: Internal error during Chroma operation
//
// Behavior:
//   - Reads the object_id and operation from query parameters
//   - Validates presence and correctness of inputs
//   - Extracts precomputed embeddings from in-memory jobResults map
//   - Injects embeddings into the payload and calls ChromaDB API (add/update)
func HandleExportToChroma(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	id := r.URL.Query().Get("object_id")
	operation := r.URL.Query().Get("operation")

	if id == "" {
		slog.Error("missing object_id", slog.String("handler", "HandleExportToChroma"))
		http.Error(w, "Missing object_id", http.StatusBadRequest)
		return
	}
	if operation == "" {
		slog.Error("missing operation", slog.String("object_id", id), slog.String("handler", "HandleExportToChroma"))
		http.Error(w, "operation missing", http.StatusBadRequest)
		return
	}

	if operation != "update" && operation != "add" {
		slog.Error("invalid operation param", slog.String("operation", operation), slog.String("object_id", id))
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
		slog.Error("invalid JSON body", slog.Any("error", err), slog.String("handler", "HandleExportToChroma"))
		http.Error(w, "invalid JSON body"+err.Error(), http.StatusBadRequest)
		return
	}

	results, ok := jobResults[id]
	if !ok {
		slog.Error("embedding not found", slog.String("object_id", id), slog.String("handler", "HandleExportToChroma"))
		http.Error(w, "embedding not found for object_id", http.StatusNotFound)
		return
	}

	req = body.Req
	payload = body.Payload
	payload.Embeddings = results.Embeddings
	payload.IDs = results.Filenames
	payload.Documents = results.Filecontent

	slog.Info("embedding export payload size",
		slog.Int("ids", len(payload.IDs)),
		slog.Int("docs", len(payload.Documents)),
		slog.Int("embeds", len(payload.Embeddings)),
		slog.Int("metas", len(payload.Metadatas)),
	)

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
		slog.Error("chroma operation failed", slog.String("operation", operation), slog.Any("error", err))
		http.Error(w, err.Error(), status)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Chroma operation succeeded"))
}
