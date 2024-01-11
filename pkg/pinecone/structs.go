package pinecone

type queryInput struct {
	Namespace       string      `json:"namespace"`
	TopK            int64       `json:"top_k"`
	Vector          []float64   `json:"vector"`
	IncludeValues   bool        `json:"include_values"`
	IncludeMetadata bool        `json:"include_metadata"`
	ID              string      `json:"id"`
	Filter          interface{} `json:"filter"`
}

type queryReq struct {
	Namespace       string      `json:"namespace"`
	TopK            int64       `json:"topK"`
	Vector          []float64   `json:"vector,omitempty"`
	IncludeValues   bool        `json:"includeValues"`
	IncludeMetadata bool        `json:"includeMetadata"`
	ID              string      `json:"id,omitempty"`
	Filter          interface{} `json:"filter,omitempty"`
}

type queryResp struct {
	Namespace string  `json:"namespace"`
	Matches   []match `json:"matches"`
}

type match struct {
	vector
	Score float64 `json:"score"`
}

type upsertReq struct {
	Vectors []vector `json:"vectors"`
}

type vector struct {
	ID       string      `json:"id"`
	Values   []float64   `json:"values,omitempty"`
	Metadata interface{} `json:"metadata,omitempty"`
}

type upsertResp struct {
	RecordsUpserted int64 `json:"upsertedCount"`
}

type upsertOutput struct {
	RecordsUpserted int64 `json:"upserted_count"`
}

type errBody struct {
	Msg string `json:"message"`
}

func (e errBody) Message() string {
	return e.Msg
}
