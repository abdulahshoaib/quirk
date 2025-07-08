package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/abdulahshoaib/quirk/handlers"
	"github.com/abdulahshoaib/quirk/middleware"
)

func main() {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}

	middleware.InitDB(db)

	mux := http.NewServeMux()

<<<<<<< HEAD
	mux.HandleFunc("/process", middleware.Logging(middleware.Auth(handlers.HandleProcess)))
	mux.HandleFunc("/status", middleware.Logging(middleware.Auth(handlers.HandleStatus)))
	mux.HandleFunc("/result", middleware.Logging(middleware.Auth(handlers.HandleResult)))
	mux.HandleFunc("/export", middleware.Logging(middleware.Auth(handlers.HandleExport)))

	// mux.HandleFunc("/signup", middleware.Logging(handlers.HandleSignUp))
=======
	mux.HandleFunc("/process", middleware.Logging(handlers.HandleProcess))
	mux.HandleFunc("/status", middleware.Logging(handlers.HandleStatus))
	mux.HandleFunc("/result", middleware.Logging(handlers.HandleResult))
	mux.HandleFunc("/export", middleware.Logging(handlers.HandleExport))
	mux.HandleFunc("/signup", middleware.Logging(handlers.HandleLogin))
>>>>>>> 665f30f (feat: added signup handling)

	log.Print("serving on :8080")
	http.ListenAndServe(":8080", mux)
}
