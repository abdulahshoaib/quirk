package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	chromadb "github.com/abdulahshoaib/quirk/chromaDB"
)

func HandleQuery(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	var input struct {
		Req  chromadb.ReqParams `json:"req"`
		Text []string           `json:"text"`
	}

	// Parse and decode JSON
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		slog.Error("invalid request body", slog.Any("error", err), slog.String("handler", "HandleQuery"))
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Call ListCollections with extracted data
	status, err, res := chromadb.ListCollections(input.Req, input.Text)
	if err != nil {
		slog.Error("chroma query failed", slog.Any("error", err), slog.Int("status", status), slog.String("handler", "HandleQuery"))
		http.Error(w, fmt.Sprintf("query failed: %v", err), status)
		return
	}

	slog.Info("Chroma query succeeded", slog.Int("status", status), slog.Any("response", res), slog.String("handler", "HandleQuery"))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
