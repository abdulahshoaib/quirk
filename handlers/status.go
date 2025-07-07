package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

// GET /status -> Check processing status by object_id
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
