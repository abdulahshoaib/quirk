package chromadb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/abdulahshoaib/quirk/pipeline"
)

// checkHealth verifies the availability of a ChromaDB instance by sending a heartbeat
// request to the tenant/database endpoint.
//
// Returns:
//   - 200 OK if the database is healthy
//   - Error and HTTP status code if health check fails
func CheckHealth(req ReqParams) (int, error) {
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
//
//	POST /api/v2/tenants/{tenant}/databases/{database}/collections/{collection_id}/add
//
// Parameters:
//   - req: Contains connection details for ChromaDB
//   - payload: Collection payload (metadata, IDs, embeddings, etc.)
//
// Returns:
//   - 200 OK if operation succeeds
//   - Error and HTTP status code if marshaling or HTTP request fails
func CreateNewCollection(req ReqParams, payload Payload) (int, error) {
	url := fmt.Sprintf("http://%s:%d/api/v2/tenants/%s/databases/%s/collections/%s/add",
		req.Host,
		req.Port,
		req.Tenant,
		req.Database,
		req.Collection_id,
	)

	log.Println(url)

	body, err := json.Marshal(payload)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to marshal payload: %w", err)
	}

	res, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer res.Body.Close()
	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to read response: %w", err)
	}

	var prettyJSON bytes.Buffer
	json.Indent(&prettyJSON, respBody, "", "  ")
	if err != nil {
		log.Printf("Failed to format JSON: %v", err)
	} else {
		log.Printf("Res \n%s", prettyJSON.String())
	}

	return http.StatusOK, nil
}

// UpdateCollection updates an existing ChromaDB collection with new embeddings,
// after verifying database health.
//
// Endpoint:
//
//	POST /api/v2/tenants/{tenant}/databases/{database}/collections/{collection_id}/update
//
// Parameters:
//   - req: Contains connection details for ChromaDB
//   - payload: Updated collection payload with embeddings
//
// Returns:
//   - 200 OK if operation succeeds
//   - Error and HTTP status code if marshaling or HTTP request fails
func UpdateCollection(req ReqParams, payload Payload) (int, error) {
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
	log.Println(string(body))

	res, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer res.Body.Close()

	return http.StatusOK, nil
}

func ListCollections(req ReqParams, query_text []string) (int, error, *ChromaQueryResponse) {
	url := fmt.Sprintf("http://%s:%d/api/v2/tenants/%s/databases/%s/collections/%s/query",
		req.Host,
		req.Port,
		req.Tenant,
		req.Database,
		req.Collection_id,
	)

	query_embeddings, err := pipeline.EmbeddingsAPI(query_text)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("embedding failed: %s", err), nil
	}

	if len(query_embeddings) == 0 || len(query_embeddings[0]) == 0 {
		return http.StatusBadRequest, fmt.Errorf("no valid embeddings returned"), nil
	}

	payload := map[string]any{
		"include":          []string{"distances", "documents"},
		"n_results":        10,
		"query_embeddings": query_embeddings,
		// "where":         "optional-if-needed",
		// "where_document":"optional-if-needed",
		// "ids":           []string{"optional-id"},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to marshal payload: %w", err), nil
	}

	res, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("HTTP request failed: %w", err), nil
	}
	defer res.Body.Close()
	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to read response: %w", err), nil
	}

	if res.StatusCode < 200 && res.StatusCode >= 300 {
		return res.StatusCode, fmt.Errorf(string(respBody)), nil
	}

	var parsed ChromaQueryResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("invalid JSON format: %w", err), nil
	}

	return res.StatusCode, nil, &parsed
}
