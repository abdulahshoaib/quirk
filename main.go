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

func main() {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	log.Println("DSN =", dsn)

	Db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}

	if err := Db.Ping(); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}

	middleware.InitDB(Db)
	handlers.InitDB(Db)

	mux := http.NewServeMux()

	mux.HandleFunc("/login", middleware.Logging((handlers.HandleLogin)))
	mux.HandleFunc("/process", middleware.Logging((handlers.HandleProcess)))
	mux.HandleFunc("/status", middleware.Logging((handlers.HandleStatus)))
	mux.HandleFunc("/result", middleware.Logging((handlers.HandleResult)))
	mux.HandleFunc("/export", middleware.Logging((handlers.HandleExport)))

	mux.HandleFunc("/protected", handlers.AuthenticateJWT(handleProtectedRoute))

	// mux.HandleFunc("/signup", middleware.Logging(handlers.HandleSignUp))
	log.Print("serving on :8080")
	http.ListenAndServe(":8080", mux)
}

func handleProtectedRoute(w http.ResponseWriter, r *http.Request) {
	// Get email from context
	email, ok := r.Context().Value("email").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Use the email...
	fmt.Fprintf(w, "Welcome %s!", email)
}
