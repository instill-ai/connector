package pinecone

type QueryInput struct {
	Namespace       string    `json:"namespace"`
	TopK            int64     `json:"top_k"`
	Vector          []float64 `json:"vector"`
	IncludeValues   bool      `json:"include_values"`
	IncludeMetadata bool      `json:"include_metadata"`
	ID              string    `json:"id"`
}

type QueryReq struct {
	Namespace       string    `json:"namespace"`
	TopK            int64     `json:"topK"`
	Vector          []float64 `json:"vector"`
	IncludeValues   bool      `json:"includeValues"`
	IncludeMetadata bool      `json:"includeMetadata"`
	ID              string    `json:"id"`
}

type QueryResp struct {
	Namespace string  `json:"namespace"`
	Matches   []Match `json:"matches"`
}

type Match struct {
	Vector
	Score float64 `json:"score"`
}

type UpsertReq struct {
	Vectors []Vector `json:"vectors"`
}

type Vector struct {
	ID       string      `json:"id"`
	Values   []float64   `json:"values"`
	Metadata interface{} `json:"metadata,omitempty"`
}

type UpsertResp struct {
	RecordsUpserted int64 `json:"upsertedCount"`
}

type UpsertOutput struct {
	RecordsUpserted int64 `json:"upserted_count"`
}
