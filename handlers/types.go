package handlers

import (
	"database/sql"
	"time"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/lib/pq"
)

type JobStatus struct {
	Status string
	ETA    time.Time
	Error  string
}

type Result struct {
	Embeddings  [][]float64
	Triples     []string
	Filenames   []string
	Filecontent []string
}

type Name struct {
	filename string
}

type UserCredentials struct {
	Email string `json:"email"`
}

type Claims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

var Db *sql.DB

func InitDB(database *sql.DB) {
	Db = database
}
