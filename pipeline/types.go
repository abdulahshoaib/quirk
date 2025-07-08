package pipeline

// call back for the embedding data recived to be written to the handler package
type ResultWriter func(object_id string, embeddings [][]float64, triples []string)

type BGEReq struct {
	Text    []string `json:"text"`
	Pooling string   `json:"pooling,omitempty"` // cls or mean
}

type BGERes struct {
	Result struct {
		Data [][]float64 `json:"data"`
	} `json:"result"`
}
