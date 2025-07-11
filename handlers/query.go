package handlers

import (
	"encoding/json"
	"net/http"
	"fmt"

	chromadb "github.com/abdulahshoaib/quirk/chromaDB"
)

func HandleQuery(w http.ResponseWriter, r *http.Request) {
	// Define anonymous struct for decoding request body
	var input struct {
		Req  chromadb.ReqParams `json:"req"`
		Text []string           `json:"text"`
	}

	// Parse and decode JSON
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Call ListCollections with extracted data
	status, err := chromadb.ListCollections(input.Req, input.Text)
	if err != nil {
		http.Error(w, fmt.Sprintf("query failed: %v", err), status)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"query successful"}`))
}
