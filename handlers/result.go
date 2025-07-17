package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

// HandleResult returns the processed embeddings and knowledge triples
// for a given object ID, if the job has been completed.
//
// GET /result?object_id={id}
//
// Query Parameters:
//   - object_id (required): Unique identifier for the processing job
//
// Response Codes:
//   - 200 OK: Job completed; returns JSON with embeddings and triples
//   - 202 Accepted: Job is still in progress or incomplete
//   - 400 Bad Request: Missing object_id
//   - 404 Not Found: Unknown or invalid object_id
//
// Example JSON Response:
//
//	{
//	  "embeddings": [
//	    { "text": "Some text", "vector": [0.12, 0.98, ...] }
//	  ],
//	  "triples": [
//	    { "subject": "Earth", "predicate": "is", "object": "planet" }
//	  ]
//	}
func HandleResult(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	id := r.URL.Query().Get("object_id")
	if id == "" {
		log.Fatal("Missing object_id parameter")
		http.Error(w, "Missing object_id parameter", http.StatusBadRequest)
		return
	}

	mutex.RLock()
	status, exists := jobStatuses[id]
	result, hasResult := jobResults[id]
	mutex.RUnlock()

	if !exists {
		log.Fatal("Not found")
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if status.Status != "completed" || !hasResult {
		log.Fatal("Result no ready")
		http.Error(w, "Result not ready", http.StatusAccepted)
		return
	}
	json.NewEncoder(w).Encode(result)
}
