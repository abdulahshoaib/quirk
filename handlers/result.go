package handlers

import (
	"net/http"
)

// GET /result -> Return embeddings + triples if conversion is complete
func HandleResult(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("/result hit"))
}
