package archetypeai

import (
	"github.com/instill-ai/connector/pkg/util/httpclient"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	host          = "https://api.archetypeai.dev"
	summarizePath = "/v0.3/summarize"
)

func newClient(config *structpb.Struct, logger *zap.Logger) *httpclient.Client {
	c := httpclient.New("Archetype AI", getBasePath(config),
		httpclient.WithLogger(logger),
		httpclient.WithEndUserError(new(errBody)),
	)

	c.SetAuthToken(getAPIKey(config))

	return c
}

type errBody struct {
	Error string `json:"error"`
}

func (e errBody) Message() string {
	return e.Error
}
