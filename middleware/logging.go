package middleware

import (
	"net/http"
	"log/slog"
)

// fn logging(fn http.HandlerFunc) http.HandlerFunc
//
// Takes in the Handler Function for a route and logs
// the route when hit
func Logging(fn http.HandlerFunc) http.HandlerFunc{
	return func (w http.ResponseWriter, r *http.Request)  {
		slog.Info("route mounted", r.URL.Path)
		fn(w, r)
	}
}

