package openai

import (
	"net/http"

	"github.com/instill-ai/connector/pkg/util/httpclient"
)

const (
	listModelsURL = host + "/v1/models"
)

// Model represents a OpenAI Model
type Model struct {
	ID         string            `json:"id"`
	Object     string            `json:"object"`
	Created    int               `json:"created"`
	OwnedBy    string            `json:"owned_by"`
	Permission []ModelPermission `json:"permission"`
	Root       string            `json:"root"`
}

type ModelPermission struct {
	ID                 string `json:"id"`
	Object             string `json:"object"`
	Created            int    `json:"created"`
	AllowCreateEngine  bool   `json:"allow_create_engine"`
	AllowSampling      bool   `json:"allow_sampling"`
	AllowLogprobs      bool   `json:"allow_logprobs"`
	AllowSearchIndices bool   `json:"allow_search_indices"`
	AllowView          bool   `json:"allow_view"`
	AllowFineTuning    bool   `json:"allow_fine_tuning"`
	Organization       string `json:"organization"`
	IsBlocking         bool   `json:"is_blocking"`
}

type ListModelsResponse struct {
	Object string  `json:"object"`
	Data   []Model `json:"data"`
}

// ListModels calls the list models endpoint and returns the available models.
// https://platform.openai.com/docs/api-reference/models/list
func (c *Client) ListModels() (resp ListModelsResponse, err error) {
	err = c.sendReqAndUnmarshal(listModelsURL, http.MethodGet, httpclient.MIMETypeJSON, nil, &resp)
	return resp, err
}
