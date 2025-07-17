package handlers

import (
	"encoding/json"
	"fmt"
	"log"
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
		log.Fatal("invalid request body")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Call ListCollections with extracted data
	status, err, res := chromadb.ListCollections(input.Req, input.Text)
	if err != nil {
		log.Fatal(err)
		http.Error(w, fmt.Sprintf("query failed: %v", err), status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
