package archetypeai

// summarizeParams holds the input of a summarize task.
type summarizeParams struct {
	Query   string   `json:"query"`
	FileIDs []string `json:"file_ids"`
}

// summarizeOutput is used to return the output of a summarize task execution.
type summarizeOutput struct {
	Response string `json:"response"`
}

// summarizeReq holds the params for the Archetype AI API call.
type summarizeReq struct {
	Query   string   `json:"query"`
	FileIDs []string `json:"file_ids"`
}

const (
	statusCompleted = "completed"
	statusFailed    = "failed"
)

// summarizeResp holds the response from the Archetype AI API call.
type summarizeResp struct {
	QueryID  string `json:"query_id"`
	Status   string `json:"status"`
	Response struct {
		ProcessedText string `json:"processed_text"`
	} `json:"response"`
}
