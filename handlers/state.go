package handlers

import "sync"

var (
	jobStatuses = map[string]JobStatus{}
	jobResults  = map[string]Result{}
	mutex       = sync.RWMutex{}
)
