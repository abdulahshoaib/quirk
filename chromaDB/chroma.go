package chromadb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// checkHealth verifies the availability of a ChromaDB instance by sending a heartbeat
// request to the tenant/database endpoint.
//
// Returns:
//   - 200 OK if the database is healthy
//   - Error and HTTP status code if health check fails
func checkHealth(req ReqParams) (int, error) {
	url := fmt.Sprintf("http://%s:%d/api/v2/tenants/%s/databases/%s/heartbeat",
		req.Host,
		req.Port,
		req.Tenant,
		req.Database,
	)

	res, err := http.Get(url)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to reach ChromaDB: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return res.StatusCode, fmt.Errorf("health check failed: received status %d", res.StatusCode)
	}

	return res.StatusCode, nil
}

// CreateNewCollection adds a new collection with embeddings to ChromaDB,
// after verifying database health.
//
// Endpoint:
//   POST /api/v2/tenants/{tenant}/databases/{database}/collections/{collection_id}/add
//
// Parameters:
//   - req: Contains connection details for ChromaDB
//   - payload: Collection payload (metadata, IDs, embeddings, etc.)
//
// Returns:
//   - 200 OK if operation succeeds
//   - Error and HTTP status code if marshaling or HTTP request fails
func CreateNewCollection(req ReqParams, payload Payload) (int, error) {
	status, err := checkHealth(req)
	if status != http.StatusOK || err != nil {
		log.Fatal("Health Check failed")
		return status, err
	}

	url := fmt.Sprintf("http://%s:%d/api/v2/tenants/%s/databases/%s/collections/%s/add",
		req.Host,
		req.Port,
		req.Tenant,
		req.Database,
		req.Collection_id,
	)

	body, err := json.Marshal(payload)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to marshal payload: %w", err)
	}

	log.Printf("Req: %s", body)

	res, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer res.Body.Close()

	return http.StatusOK, nil
}

// UpdateCollection updates an existing ChromaDB collection with new embeddings,
// after verifying database health.
//
// Endpoint:
//   POST /api/v2/tenants/{tenant}/databases/{database}/collections/{collection_id}/update
//
// Parameters:
//   - req: Contains connection details for ChromaDB
//   - payload: Updated collection payload with embeddings
//
// Returns:
//   - 200 OK if operation succeeds
//   - Error and HTTP status code if marshaling or HTTP request fails
func UpdateCollection(req ReqParams, payload Payload) (int, error) {
	status, err := checkHealth(req)
	if status != http.StatusOK || err != nil {
		log.Fatal("Health Check failed")
		return status, err
	}

	url := fmt.Sprintf("http://%s:%d/api/v2/tenants/%s/databases/%s/collections/%s/update",
		req.Host,
		req.Port,
		req.Tenant,
		req.Database,
		req.Collection_id,
	)

	body, err := json.Marshal(payload)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to marshal payload: %w", err)
	}

	res, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer res.Body.Close()

	return http.StatusOK, nil
}
