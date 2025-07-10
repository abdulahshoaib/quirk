package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/abdulahshoaib/quirk/handlers"
	"github.com/abdulahshoaib/quirk/middleware"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
)

var db *sql.DB

func main() {

	var err error
	db, err = connect()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	log.Println("Database connection successful")

	middleware.InitDB(db)
	handlers.InitDB(db)

	mux := http.NewServeMux()

	mux.HandleFunc("/login", middleware.Logging((handlers.HandleSignup)))
	mux.HandleFunc("/process", middleware.Logging((handlers.HandleProcess)))
	mux.HandleFunc("/status", middleware.Logging((handlers.HandleStatus)))
	mux.HandleFunc("/result", middleware.Logging((handlers.HandleResult)))
	mux.HandleFunc("/export", middleware.Logging((handlers.HandleExport)))
	mux.HandleFunc("/signup", middleware.Logging(handlers.HandleSignup))
	mux.HandleFunc("/export-chroma", middleware.Logging(handlers.HandleExportToChroma))

	//	mux.HandleFunc("/process", middleware.Logging(middleware.Auth(handlers.HandleProcess)))
	//	mux.HandleFunc("/status", middleware.Logging(middleware.Auth(handlers.HandleStatus)))
	//	mux.HandleFunc("/result", middleware.Logging(middleware.Auth(handlers.HandleResult)))
	//	mux.HandleFunc("/export", middleware.Logging(middleware.Auth(handlers.HandleExport)))

	log.Print("serving on :8080")
	http.ListenAndServe(":8080", mux)
}
