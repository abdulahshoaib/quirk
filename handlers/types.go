package handlers

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JobStatus struct {
	Status string
	ETA    time.Time
	Error  string
}

type Result struct {
	Embeddings [][]float64
	Triples    []string
}

type UserCredentials struct {
	Email string `json:"email"`
}
