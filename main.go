package main

import (
	"log"
	"net/http"

	"github.com/abdulahshoaib/quirk/handlers"
	"github.com/abdulahshoaib/quirk/middleware"
)

func main() {

	mux := http.NewServeMux()

	mux.HandleFunc("/process", middleware.Logging(middleware.Auth(handlers.HandleProcess)))
	mux.HandleFunc("/status", middleware.Logging(middleware.Auth(handlers.HandleStatus)))
	mux.HandleFunc("/result", middleware.Logging(middleware.Auth(handlers.HandleResult)))
	mux.HandleFunc("/export", middleware.Logging(middleware.Auth(handlers.HandleExport)))

	log.Print("serving on :8080")
	http.ListenAndServe(":8080", mux)
}
