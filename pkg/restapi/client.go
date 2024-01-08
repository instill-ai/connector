package restapi

import (
	"github.com/instill-ai/connector/pkg/util/httpclient"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

type TaskInput struct {
	EndpointPath *string                `json:"endpoint_path,omitempty"`
	Body         map[string]interface{} `json:"body,omitempty"`
}

type TaskOutput struct {
	StatusCode int                    `json:"status_code"`
	Body       map[string]interface{} `json:"body"`
	Header     map[string][]string    `json:"header"`
}

func newClient(config *structpb.Struct, logger *zap.Logger) (*httpclient.Client, error) {
	c := httpclient.New("REST API", getBaseURL(config),
		httpclient.WithLogger(logger),
	)

	auth, err := getAuthentication(config)
	if err != nil {
		return nil, err
	}

	if err := auth.setAuthInClient(c); err != nil {
		return nil, err
	}

	return c, nil
}
