package pipeline

import (
	"bytes"
	"encoding/json"
	"fmt"
	// "log"
	"net/http"
	"os"
)

func EmbeddingsAPI(texts []string) ([][]float64, error) {
	account_id := os.Getenv("CLOUDFLARE_ACCOUNT_ID")
	apiToken := os.Getenv("CLOUDFLARE_API_TOKEN")

	if account_id == "" || apiToken == "" {
		return nil, fmt.Errorf("missing CLOUDFLARE_ACC or CLOUDFLARE_TOKEN")
	}

	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/ai/run/@cf/baai/bge-large-en-v1.5", account_id)

	body, err := json.Marshal(BGEReq{Text: texts})
	if err != nil {
		return nil, err
	}

	 //log.Printf("%s", string(body))

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	var result BGERes
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Result.Data, nil
}
