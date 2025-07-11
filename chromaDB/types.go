package chromadb

type MetadataVal any

type Payload struct {
	Embeddings [][]float64              `json:"embeddings,omitempty"`
	Documents  []string                 `json:"documents,omitempty"`
	IDs        []string                 `json:"ids"`
	Metadatas  []map[string]MetadataVal `json:"metadatas,omitempty"`
	URI        []string                 `json:"uris,omitempty"`
}

type ReqParams struct {
	Host          string
	Port          int
	Tenant        string
	Database      string
	Collection_id string
}
