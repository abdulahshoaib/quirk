package handlers

import (
	"encoding/csv"
	"encoding/json"
	"net/http"
)

// HandleExport returns the embeddings and triples in CSV or JSON format
// based on the provided object ID and export format.
//
// GET /export?object_id={id}&format={csv|json}
//
// Parameters:
//   - object_id (required): The unique identifier of the processed job
//   - format (required): Either "csv" or "json"
//
// Returns:
//   - 200: File content in requested format
//   - 400: Missing object_id or invalid format
//   - 404: If object_id is not found or job not completed
//
// Content-Type:
//   - text/csv for CSV exports
//   - application/json for JSON exports
func HandleExport(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("object_id")
	format := r.URL.Query().Get("format")

	if id == "" {
		http.Error(w, "object_id missing", http.StatusBadRequest)
		return
	}

	mutex.RLock()
	result, ok := jobResults[id]
	mutex.RUnlock()

	if !ok {
		http.Error(w, "Result not found", http.StatusNotFound)
		return
	}

	switch format {

	case "csv":
		w.Header().Set("Content-Type", "text/csv")
		writer := csv.NewWriter(w)
		writer.Write([]string{"Embeddings", "Triple"})
		for i := range result.Embeddings {
			triple := ""
			if i < len(result.Triples) {
				triple = result.Triples[i]
			}
			writer.Write([]string{result.Embeddings[i], triple})
		}
		writer.Flush()
		return

	case "json":
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
		return

	default:
		http.Error(w, "Format unrecognized", http.StatusBadRequest)
		return
	}
}
