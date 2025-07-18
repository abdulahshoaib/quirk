package main

import (
	"database/sql"
	"fmt"
	"log/slog"
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
		slog.Info("database does not exist, creating", slog.String("db_name", dbName))
		_, err = adminDB.Exec(fmt.Sprintf("CREATE DATABASE \"%s\"", dbName)) // quotes for safety
		if err != nil {
			return fmt.Errorf("failed to create database: %v", err)
		}
		slog.Info("created database", slog.String("db_name", dbName))
	} else {
		slog.Info("database already exists", slog.String("db_name", dbName))
	}

	// Now connect to the actual app database
	db, err := connect()
	if err != nil {
		return fmt.Errorf("failed to connect to DB: %v", err)
	}
	defer db.Close()

	slog.Info("connected to database successfully", slog.String("db_name", dbName))

	middleware.InitDB(db)
	handlers.InitDB(db)

	if err := InitSchema(db); err != nil {
		return fmt.Errorf("failed to sync db: %v", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/signup", middleware.Logging(handlers.HandleSignup))

	mux.HandleFunc("/process", middleware.Logging(handlers.AuthenticateJWT(handlers.HandleProcess)))
	mux.HandleFunc("/status", middleware.Logging(handlers.AuthenticateJWT(handlers.HandleStatus)))
	mux.HandleFunc("/result", middleware.Logging(handlers.AuthenticateJWT(handlers.HandleResult)))
	mux.HandleFunc("/export", middleware.Logging(handlers.AuthenticateJWT(handlers.HandleExport)))
	mux.HandleFunc("/export-chroma", middleware.Logging(handlers.AuthenticateJWT(handlers.HandleExportToChroma)))
	mux.HandleFunc("/query", middleware.Logging(handlers.AuthenticateJWT(handlers.HandleQuery)))
	// following command was used to check authentication
	// mux.HandleFunc("/protected", handlers.AuthenticateJWT(handleProtectedRoute))
	//
	//
	//	mux.HandleFunc("/process", middleware.Logging(middleware.Auth(handlers.HandleProcess)))
	//	mux.HandleFunc("/status", middleware.Logging(middleware.Auth(handlers.HandleStatus)))
	//	mux.HandleFunc("/result", middleware.Logging(middleware.Auth(handlers.HandleResult)))
	//	mux.HandleFunc("/export", middleware.Logging(middleware.Auth(handlers.HandleExport)))

	slog.Info("server started", slog.String("port", "8080"))
	return http.ListenAndServe(":8080", mux)
}

func main() {
	if err := RunApp(); err != nil {
		slog.Error("server", slog.Any("error", err))
	}
}
