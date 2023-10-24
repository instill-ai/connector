package stabilityai

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
	err := c.sendReq(listEnginesURL, "GET", jsonMimeType, nil, &engines)
	return engines, err
}
