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
	adminDSN := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
	)

	adminDB, err := sql.Open("postgres", adminDSN)
	if err != nil {
		return fmt.Errorf("failed to connect to admin db: %v", err)
	}
	defer adminDB.Close()

	// Check if DB exists
	dbName := os.Getenv("DB_NAME")
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM pg_database WHERE LOWER(datname) = LOWER($1))`
	err = adminDB.QueryRow(query, dbName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check database existence: %v", err)
	}

	if !exists {
		log.Printf("Database %s does not exist, creating...", dbName)
		_, err = adminDB.Exec(fmt.Sprintf("CREATE DATABASE \"%s\"", dbName)) // quotes for safety
		if err != nil {
			return fmt.Errorf("failed to create database: %v", err)
		}
		log.Printf("Database %s created successfully", dbName)
	} else {
		log.Printf("Database %s already exists", dbName)
	}

	// Now connect to the actual app database
	db, err := connect()
	if err != nil {
		return fmt.Errorf("failed to connect to DB: %v", err)
	}
	defer db.Close()

	log.Println("Connected to DB successfully")

	middleware.InitDB(db)
	handlers.InitDB(db)

	if err := InitSchema(db); err != nil {
		return fmt.Errorf("failed to sync db: %v", err)
	}

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

	log.Println("Server Started on Port 8080")
	return http.ListenAndServe(":8080", mux)
}

func main() {
	if err := RunApp(); err != nil {
		log.Fatalf("app failed: %v", err)
	}
}
