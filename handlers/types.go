package handlers

import "time"

type JobStatus struct {
	Status string
	ETA    time.Time
	Error  string
}

type Result struct {
	Embeddings []string
	Triples    []string
}
