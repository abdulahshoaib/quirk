package chromadb

type MetadataVal any

type Payload struct {
	Documents  []string                 `json:"documents,omitempty"`
	Embeddings [][]float64              `json:"embeddings,omitempty"`
	IDs        []string                 `json:"ids"`
	Metadata   []map[string]MetadataVal `json:"metadatas,omitempty"`
	URI        []string                 `json:"uris,omitempty"`
}

type ReqParams struct {
	Host          string
	Port          int
	Tenant       string
	Database      string
	Collection_id string
}
