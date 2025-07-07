package handlers

import (
	"net/http"
)

// GET /export -> Export result to CSV or JSON by object_id (via ?format=csv or json)
func HandleExport(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("/export hit"))
}
