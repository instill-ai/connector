package httpclient

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"

	"github.com/instill-ai/x/errmsg"
)

const (
	reqTimeout = time.Second * 60 * 5

	// MIMETypeJSON defines the MIME type for JSON documents.
	MIMETypeJSON = "application/json"
)

// Client performs HTTP requests for connectors, implementing error handling
// and logging in a consistent way.
type Client struct {
	*resty.Client

	name string
}

// Option provides configuration options for a client.
type Option func(*Client)

// WithLogger will use the provider logger to log the request and response
// information.
func WithLogger(logger *zap.Logger) Option {
	return func(c *Client) {
		logger := logger.With(zap.String("name", c.name))

		c.SetLogger(logger.Sugar()).OnError(func(req *resty.Request, err error) {
			logger := logger.With(zap.String("url", req.URL))

			if v, ok := err.(*resty.ResponseError); ok {
				logger = logger.With(
					zap.Int("status", v.Response.StatusCode()),
					zap.ByteString("body", v.Response.Body()),
				)
			}

			logger.Warn("HTTP request failed", zap.Error(err))
		})
	}
}

// ErrBody allows Client to extract an error message from the API.
type ErrBody interface {
	Message() string
}

func wrapWithErrMessage(apiName string) func(*resty.Client, *resty.Response) error {
	return func(_ *resty.Client, resp *resty.Response) error {
		if !resp.IsError() {
			return nil
		}

		var issue string

		if v, ok := resp.Error().(ErrBody); ok && v.Message() != "" {
			issue = v.Message()
		}

		// Certain errors are returned as text/plain, e.g. incorrect API key
		// (401) vs invalid /query request (400) in Pinecone.
		// This is also a fallback if the error format is unexpected. It's
		// better to pass the error response to the user than displaying
		// nothing.
		if issue == "" {
			issue = resp.String()
		}

		if issue == "" {
			issue = fmt.Sprintf("Please refer to %s's API reference for more information.", apiName)
		}

		msg := fmt.Sprintf("%s responded with a %d status code. %s", apiName, resp.StatusCode(), issue)
		return errmsg.AddMessage(fmt.Errorf("unsuccessful HTTP response"), msg)
	}
}

// WithEndUserError will unmarshal error response bodies as the error struct
// and will use their message as an end-user error.
func WithEndUserError(e ErrBody) Option {
	return func(c *Client) {
		c.SetError(e).OnAfterResponse(wrapWithErrMessage(c.name))
	}
}

// New returns an httpclient configured to call a remote host.
func New(name, host string, options ...Option) *Client {
	r := resty.New().
		SetBaseURL(host).
		SetHeader("Accept", MIMETypeJSON).
		SetTimeout(reqTimeout).
		SetTransport(&http.Transport{
			DisableKeepAlives: true,
		})

	c := &Client{
		Client: r,
		name:   name,
	}

	for _, o := range options {
		o(c)
	}

	return c
}

// WrapURLError is a helper to add an end-user message to trasnport errros.
//
// Resty doesn't provide a hook for errors in `http.Client.Do`, e.g. if the
// connector configuration contains an invalid URL. This wrapper offers
// clients a way to handle such cases:
//
//	if _, err := httpclient.New(name, host).R().Post(url); err != nil {
//	    return nil, httpclient.WrapURLError(err)
//	}
func WrapURLError(err error) error {
	uerr := new(url.Error)
	if errors.As(err, &uerr) {
		err = errmsg.AddMessage(
			err,
			fmt.Sprintf("Failed to call %s. Please check that the connector configuration is correct.", uerr.URL),
		)
	}

	return err
}
