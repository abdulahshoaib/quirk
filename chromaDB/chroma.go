package chromadb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func checkHealth(req ReqParams) (int, error) {
	url := fmt.Sprintf("http://%s:%d/api/v2/tenants/%s/databases/%s/heartbeat",
		req.Host,
		req.Port,
		req.Tenant,
		req.Database,
	)

	res, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to reach ChromaDB: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return res.StatusCode, fmt.Errorf("health check failed: received status %d", res.StatusCode)
	}

	return res.StatusCode, nil
}

// http://$host:$port/api/v2/tenants/$tenants/databases/$databas}/collections/$collection_id/add
func CreateNewCollection(req ReqParams, payload Payload) (int, error) {
	// status, err := checkHealth(req)
	// if status != http.StatusOK || err != nil {
	// 	log.Fatal("Health Check failed")
	// 	return status, err
	// }

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

// http://$host:$port/api/v2/tenants/$tenants/databases/$databas}/collections/$collection_id/update
func UpdateCollection(req ReqParams, payload Payload) (int, error) {
	// status, err := checkHealth(req)
	// if status != http.StatusOK || err != nil {
	// 	log.Fatal("Health Check failed")
	// 	return status, err
	// }

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
