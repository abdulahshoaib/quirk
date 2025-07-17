package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/abdulahshoaib/quirk/handlers"
	"github.com/abdulahshoaib/quirk/middleware"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
)

func RunApp() error {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	connect()
	log.Println("connected to DB")

	Db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open db: %v", err)
	}

	if err := Db.Ping(); err != nil {
		return fmt.Errorf("failed to ping db: %v", err)
	}

	middleware.InitDB(Db)
	handlers.InitDB(Db)

	mux := http.NewServeMux()

	mux.HandleFunc("/login", middleware.Logging((handlers.HandleSignup)))
	mux.HandleFunc("/process", middleware.Logging((handlers.HandleProcess)))
	mux.HandleFunc("/status", middleware.Logging((handlers.HandleStatus)))
	mux.HandleFunc("/result", middleware.Logging((handlers.HandleResult)))
	mux.HandleFunc("/export", middleware.Logging((handlers.HandleExport)))
	mux.HandleFunc("/signup", middleware.Logging(handlers.HandleSignup))
	mux.HandleFunc("/export-chroma", middleware.Logging(handlers.HandleExportToChroma))
	mux.HandleFunc("/query", middleware.Logging(handlers.HandleQuery))
	// following command was used to check authentication
	// mux.HandleFunc("/protected", handlers.AuthenticateJWT(handleProtectedRoute))
	//
	//
	//	mux.HandleFunc("/process", middleware.Logging(middleware.Auth(handlers.HandleProcess)))
	//	mux.HandleFunc("/status", middleware.Logging(middleware.Auth(handlers.HandleStatus)))
	//	mux.HandleFunc("/result", middleware.Logging(middleware.Auth(handlers.HandleResult)))
	//	mux.HandleFunc("/export", middleware.Logging(middleware.Auth(handlers.HandleExport)))

	return http.ListenAndServe(":8080", mux)
}

func main() {
	if err := RunApp(); err != nil {
		log.Fatalf("app failed: %v", err)
	}
}
