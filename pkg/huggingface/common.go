package huggingface

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1beta"
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
	if resp.StatusCode != http.StatusOK {
		err = checkRespForError(respBody)
		if err != nil {
			return nil, err
		}
	}

	return respBody, nil
}

type apiError struct {
	Error string `json:"error,omitempty"`
}

type apiErrors struct {
	Errors []string `json:"error,omitempty"`
}

// Checks for errors in the API response and returns them if
// found.
func checkRespForError(respJSON []byte) error {
	// Check for single error
	{
		buf := make([]byte, len(respJSON))
		copy(buf, respJSON)
		apiErr := apiError{}
		err := json.Unmarshal(buf, &apiErr)
		if err != nil {
			return err
		}
		if apiErr.Error != "" {
			return errors.New(string(respJSON))
		}
	}
	// Check for multiple errors
	{
		buf := make([]byte, len(respJSON))
		copy(buf, respJSON)
		apiErrs := apiErrors{}
		err := json.Unmarshal(buf, &apiErrs)
		if err != nil {
			return err
		}
		if apiErrs.Errors != nil {
			return errors.New(string(respJSON))
		}
	}
	return nil
}

func (c *Client) GetConnectionState() (pipelinePB.Connector_State, error) {
	req, _ := http.NewRequest(http.MethodGet, c.BaseURL, nil)
	req.Header.Set(AuthHeaderKey, AuthHeaderPrefix+c.APIKey)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return pipelinePB.Connector_STATE_ERROR, err
	}
	if resp != nil && resp.StatusCode == http.StatusOK {
		return pipelinePB.Connector_STATE_CONNECTED, nil
	}
	return pipelinePB.Connector_STATE_DISCONNECTED, nil
}
