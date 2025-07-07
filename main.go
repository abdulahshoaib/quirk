package main

import (
	"log"
	"net/http"

	"github.com/abdulahshoaib/quirk/handlers"
	"github.com/abdulahshoaib/quirk/middleware"
)

func main() {

	mux := http.NewServeMux()

	mux.HandleFunc("/process", middleware.Logging(handlers.HandleProcess))
	mux.HandleFunc("/status", middleware.Logging(handlers.HandleStatus))
	mux.HandleFunc("/result", middleware.Logging(handlers.HandleResult))
	mux.HandleFunc("/export", middleware.Logging(handlers.HandleExport))

	log.Print("serving on :8080")
	http.ListenAndServe(":8080", mux)
}
