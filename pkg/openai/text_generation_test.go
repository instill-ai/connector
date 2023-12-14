package openai

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"github.com/instill-ai/connector/pkg/util/mock"
)

func TestClient_GenerateTextCompletion(t *testing.T) {
	c := qt.New(t)

	req := TextCompletionReq{}
	key := "test_key"
	org := "org"

	testcases := []struct {
		name          string
		gotStatus     int
		gotBody       string
		wantLogFields []string
	}{
		{
			name:          "nok - 401 (unexpected response body)",
			gotStatus:     http.StatusUnauthorized,
			wantLogFields: []string{"url", "body", "status"},
		},
		{
			name:          "nok - 401",
			gotStatus:     http.StatusUnauthorized,
			gotBody:       `{ "error": { "message": "Incorrect API key provided." } }`,
			wantLogFields: []string{"url", "body", "status"},
		},
		{
			name:          "nok - JSON error",
			gotStatus:     http.StatusOK,
			gotBody:       "{",
			wantLogFields: []string{"url", "body"},
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			httpClient := &mock.HTTPClient{
				Output: func() (*http.Response, error) {
					return &http.Response{
						StatusCode: tc.gotStatus,
						Body:       io.NopCloser(strings.NewReader(tc.gotBody)),
					}, nil
				},
			}

			zCore, zLogs := observer.New(zap.InfoLevel)
			logger := zap.New(zCore)

			openAIClient := &Client{
				APIKey:     key,
				Org:        org,
				HTTPClient: httpClient,
				Logger:     logger,
			}
			_, err := openAIClient.GenerateTextCompletion(req)
			c.Check(err, qt.IsNotNil)

			// Check relevant information is logged.
			logs := zLogs.All()
			c.Assert(logs, qt.HasLen, 1)

			entry := logs[0].ContextMap()
			for _, k := range tc.wantLogFields {
				_, ok := entry[k]
				c.Check(ok, qt.IsTrue)
			}
		})
	}

	c.Run("nok - client error", func(c *qt.C) {
		httpErr := fmt.Errorf("boom")
		httpClient := &mock.HTTPClient{
			Output: func() (*http.Response, error) {
				return nil, httpErr
			},
		}

		zCore, zLogs := observer.New(zap.InfoLevel)
		logger := zap.New(zCore)

		openAIClient := &Client{
			APIKey:     key,
			Org:        org,
			HTTPClient: httpClient,
			Logger:     logger,
		}

		_, err := openAIClient.GenerateTextCompletion(req)
		c.Check(err, qt.ErrorMatches, ".*failed to call OpenAI.*boom.*")

		// Check relevant information is logged.
		logs := zLogs.All()
		c.Assert(logs, qt.HasLen, 1)

		entry := logs[0].ContextMap()
		c.Check(entry["error"], qt.Equals, httpErr.Error())
		_, ok := entry["url"]
		c.Check(ok, qt.IsTrue)
	})
}
