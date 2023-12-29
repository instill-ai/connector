package mock

import (
	"fmt"
	"net/http"

	"github.com/instill-ai/connector/pkg/util/httpclient"
)

// Check interface is implemented.
var _ httpclient.Doer = (*HTTPDoer)(nil)

// HTTPDoer is a mocked implementation of httpclient.Doer. By setting the
// Output field, clients have control over the Do method response.
type HTTPDoer struct {
	Output func() (*http.Response, error)
}

// Do implements the httpclient.Doer interface.
func (c *HTTPDoer) Do(_ *http.Request) (*http.Response, error) {
	if c.Output == nil {
		return nil, fmt.Errorf("http client not initialized")
	}

	return c.Output()
}
