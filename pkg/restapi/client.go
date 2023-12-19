package restapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/instill-ai/connector/pkg/util"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	reqTimeout = time.Second * 60 * 5
)

type TaskInput struct {
	EndpointPath *string `json:"endpoint_path,omitempty"`
	Body         Body    `json:"body,omitempty"`
}

type TaskOutput struct {
	StatusCode int                    `json:"status_code"`
	Body       map[string]interface{} `json:"body"`
	Header     map[string][]string    `json:"header"`
}

type Client struct {
	BaseURL        string         `json:"base_url"`
	Authentication Authentication `json:"authentication"`
	HTTPClient     util.HTTPClient
}

func NewClient(config *structpb.Struct) (Client, error) {
	baseURL := getBaseURL(config)
	auth, err := getAuthentication(config)
	if err != nil {
		return Client{}, err
	}

	tr := &http.Transport{
		DisableKeepAlives: true,
	}

	return Client{
		BaseURL:        baseURL,
		Authentication: auth,
		HTTPClient: &http.Client{
			Timeout:   reqTimeout,
			Transport: tr,
		},
	}, nil
}

func (c *Client) sendRequest(method string, input TaskInput) (TaskOutput, error) {
	resp := TaskOutput{}

	http.DefaultClient.Timeout = reqTimeout

	reqURL := c.BaseURL
	if input.EndpointPath != nil {
		reqURL += *input.EndpointPath
	}

	var req *http.Request
	var err error

	switch method {
	case http.MethodGet, http.MethodHead:
		// Ignore body
		req, err = http.NewRequest(method, reqURL, nil)
		if err != nil {
			return resp, err
		}
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		if input.Body == nil || input.Body.GetBodyType() == NoneBodyType || input.Body.GetBody() == nil || len(input.Body.GetBody()) == 0 {
			req, err = http.NewRequest(method, reqURL, nil)
			if err != nil {
				return resp, err
			}
		} else {
			// Convert body to JSON
			jsonBody, err := json.Marshal(input.Body.GetBody())
			if err != nil {
				return resp, err
			}
			req, err = http.NewRequest(method, reqURL, bytes.NewBuffer(jsonBody))
			if err != nil {
				return resp, err
			}
		}
	default:
		return resp, fmt.Errorf("not supported method: %s", method)
	}

	// Add authentication
	switch c.Authentication.GetAuthLocation() {
	case Header:
		key, value, err := c.Authentication.GenAuthHeader()
		if err != nil {
			return resp, err
		}
		if key != "" {
			req.Header.Add(key, value)
		}
	case Query:
		key, value, err := c.Authentication.GenAuthQuery()
		if err != nil {
			return resp, err
		}
		if key != "" {
			// Append the query parameter to the URL
			extraParams := url.Values{}
			extraParams.Add(key, value)
			if req.URL.RawQuery != "" {
				req.URL.RawQuery += "&" + extraParams.Encode()
			} else {
				req.URL.RawQuery = extraParams.Encode()
			}
		}
	default:
		return resp, nil
	}

	// Send request
	res, err := c.HTTPClient.Do(req)
	if res != nil && res.Body != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return resp, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return resp, err
	}

	resp.StatusCode = res.StatusCode
	resp.Header = res.Header
	err = json.Unmarshal(body, &resp.Body)
	if err != nil {
		return resp, err
	}
	return resp, nil
}
