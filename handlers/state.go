package handlers

import "sync"

var (
	jwtKey      = []byte("your_secret_key")
	jobStatuses = map[string]JobStatus{}
	jobResults  = map[string]Result{}
	mutex       = sync.RWMutex{}
)
