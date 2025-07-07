package handlers

import (
	"encoding/json"
	"net/http"
)

func HandleResult(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("object_id")
	if id == "" {
		http.Error(w, "Missing object_id parameter", http.StatusBadRequest)
		return
	}

	mutex.RLock()
	status, exists := jobStatuses[id]
	result, hasResult := jobResults[id]
	mutex.RUnlock()

	if !exists {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if status.Status != "completed" || !hasResult {
		http.Error(w, "Result not ready", http.StatusAccepted)
		return
	}
	json.NewEncoder(w).Encode(result)
}
