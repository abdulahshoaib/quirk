package middleware

import (
	"database/sql"
	"net/http"
	"strings"
)

var db *sql.DB

func InitDB(database *sql.DB) {
	db = database
}

// Auth authenticates incoming HTTP requests
// using a bearer token in the Authorization header
//
//   - Error: responds with HTTP 401 Unauthorized
//
// This middleware requires the database connection to be initialized via InitDB.
func Auth(nx http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Missing or invalid Authorization Header", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		token = strings.TrimSpace(token)

		if token == "" {
			http.Error(w, "Empty bearer token", http.StatusUnauthorized)
			return
		}

		var exists bool
		err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM api_tokens WHERE token = $1)", token).Scan(&exists)
		if err != nil || !exists {
			http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
			return
		}

		nx(w, r)
	}
}
