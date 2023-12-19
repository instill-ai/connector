package mock

import (
	"fmt"
	"net/http"
)

// HTTPClient is a mocked implementation of HTTPClient. By setting the Output
// field, clients have control over the Do method response.
type HTTPClient struct {
	Output func() (*http.Response, error)
}

// Do implements the HTTPClient interface.
func (c *HTTPClient) Do(_ *http.Request) (*http.Response, error) {
	if c.Output == nil {
		return nil, fmt.Errorf("http client not initialized")
	}

	return c.Output()
}
