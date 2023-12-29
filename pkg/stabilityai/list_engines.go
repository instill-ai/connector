package stabilityai

import "github.com/instill-ai/connector/pkg/util/httpclient"

const (
	listEnginesURL = host + "/v1/engines/list"
)

// Engine represents a Stability AI Engine
type Engine struct {
	Description string `json:"description"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
}

// ListEngines calls the list engine endpoint and returns the available engines.
// https://platform.stability.ai/rest-api#tag/v1engines/operation/listEngines
func (c *Client) ListEngines() ([]Engine, error) {
	var engines []Engine
	err := c.sendReq(listEnginesURL, "GET", httpclient.MIMETypeJSON, nil, &engines)
	return engines, err
}
