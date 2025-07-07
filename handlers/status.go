package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

// HandleStatus retrieves the current status of a job by its object ID.
//
// GET /status?object_id={id}
//
// Parameters:
//   - object_id (required): The unique identifier of the job
//
// Returns:
//   - 200: JSON with status, eta_seconds, and error_message
//   - 400: Missing object_id parameter
//   - 404: Job not found
//
// Example response:
//
//	- {"status": "running", "eta_seconds": 120, "error_message": null}
//  - {"status": "failed", eta_seconds": 0, "error_message": Failed...}
func HandleStatus(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("object_id")
	if id == "" {
		http.Error(w, "Missing object_id", http.StatusBadRequest)
		return
	}

	mutex.RLock()
	status, exists := jobStatuses[id]
	mutex.RUnlock()

	if !exists {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	eta := int(time.Until(status.ETA).Seconds())
	eta = max(eta, 0)

	json.NewEncoder(w).Encode(map[string]any{
		"status":        status.Status,
		"eta_seconds":   eta,
		"error_message": status.Error,
	})

	w.Write([]byte("/status hit"))
}
