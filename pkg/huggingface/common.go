package huggingface

import (
	"bytes"
	"errors"
	"io"
	"net/http"

	connectorPB "github.com/instill-ai/protogen-go/vdp/connector/v1alpha"
)

const (
	AuthHeaderKey     = "Authorization"
	AuthHeaderPrefix  = "Bearer "
	ContentTypeHeader = "Content-Type"
	modelsPath        = "/models/"
)

// MakeHFAPIRequest builds and sends an HTTP POST request to the given model
// using the provided JSON body. If the request is successful, returns the
// response JSON and a nil error. If the request fails, returns an empty slice
// and an error describing the failure.
func (c *Client) MakeHFAPIRequest(body []byte, model string) ([]byte, error) {
	url := c.BaseURL
	if !c.IsCustomEndpoint {
		url += modelsPath + model
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, errors.New("nil request created")
	}
	req.Header.Set(ContentTypeHeader, http.DetectContentType(body))
	req.Header.Set(AuthHeaderKey, AuthHeaderPrefix+c.APIKey)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return respBody, nil
}

func (c *Client) GetConnectionState() (connectorPB.ConnectorResource_State, error) {
	req, _ := http.NewRequest(http.MethodGet, c.BaseURL, nil)
	req.Header.Set(AuthHeaderKey, AuthHeaderPrefix+c.APIKey)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return connectorPB.ConnectorResource_STATE_ERROR, err
	}
	if resp != nil && resp.StatusCode == http.StatusOK {
		return connectorPB.ConnectorResource_STATE_CONNECTED, nil
	}
	return connectorPB.ConnectorResource_STATE_DISCONNECTED, nil
}
