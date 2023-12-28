package pinecone

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"
	"github.com/instill-ai/connector/pkg/util"
	"github.com/instill-ai/x/errmsg"

	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1beta"
)

const (
	reqTimeout = time.Second * 60 * 5
	taskQuery  = "TASK_QUERY"
	taskUpsert = "TASK_UPSERT"
)

//go:embed config/definitions.json
var definitionsJSON []byte

//go:embed config/tasks.json
var tasksJSON []byte

var once sync.Once
var connector base.IConnector

type Connector struct {
	base.Connector
}

type Execution struct {
	base.Execution
}

type Client struct {
	APIKey     string
	HTTPClient util.HTTPClient
	Logger     *zap.Logger
}

func Init(logger *zap.Logger) base.IConnector {
	once.Do(func() {
		connector = &Connector{
			Connector: base.Connector{
				Component: base.Component{Logger: logger},
			},
		}
		err := connector.LoadConnectorDefinitions(definitionsJSON, tasksJSON, nil)
		if err != nil {
			logger.Fatal(err.Error())
		}
	})
	return connector
}

func (c *Connector) CreateExecution(defUID uuid.UUID, task string, config *structpb.Struct, logger *zap.Logger) (base.IExecution, error) {
	e := &Execution{}
	e.Execution = base.CreateExecutionHelper(e, c, defUID, task, config, logger)
	return e, nil
}

// NewClient initializes a new Pinecone client.
func NewClient(apiKey string, logger *zap.Logger) Client {
	tr := &http.Transport{
		DisableKeepAlives: true,
	}
	return Client{
		APIKey:     apiKey,
		HTTPClient: &http.Client{Timeout: reqTimeout, Transport: tr},
		Logger:     logger,
	}
}

func getAPIKey(config *structpb.Struct) string {
	return config.GetFields()["api_key"].GetStringValue()
}

func getURL(config *structpb.Struct) string {
	return config.GetFields()["url"].GetStringValue()
}

// sendReqAndUnmarshal makes an HTTP request to Pinecone's API, unmarshalling
// the response into the provided object.
func (c *Client) sendReqAndUnmarshal(reqURL, method string, params, respObj any) error {
	logger := c.Logger.With(zap.String("url", reqURL))

	data, err := json.Marshal(params)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(method, reqURL, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", util.MIMETypeJSON)
	req.Header.Add("Accept", util.MIMETypeJSON)
	req.Header.Add("Api-Key", c.APIKey)
	http.DefaultClient.Timeout = reqTimeout

	res, err := c.HTTPClient.Do(req)
	if res != nil && res.Body != nil {
		defer res.Body.Close()
	}
	if err != nil || res == nil {
		logger.Warn("Failed to call Pinecone", zap.Error(err))
		return errmsg.AddMessage(
			fmt.Errorf("failed to call pinecone: %w", err),
			"Failed to call Pinecone's API.",
		)
	}
	respBody, _ := io.ReadAll(res.Body)
	logger = logger.With(zap.ByteString("body", respBody))

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		err := fmt.Errorf("unsuccessful response from pinecone")
		logger = logger.With(zap.Int("status", res.StatusCode))

		var errBody struct {
			Message string `json:"message"`
		}

		// We want to provide a useful error message so we don't return an
		// error here.
		if jsonErr := json.Unmarshal(respBody, &errBody); jsonErr != nil {
			logger = logger.With(zap.NamedError("json_error", jsonErr))
		}

		msg := errBody.Message
		if msg == "" {
			msg = "Please refer to Pinecone's API reference for more information."
		}
		issue := fmt.Sprintf("Pinecone responded with a %d status code. %s", res.StatusCode, msg)

		logger.Warn("Unsuccessful response from Pinecone")
		return errmsg.AddMessage(err, issue)
	}

	if err := json.Unmarshal(respBody, &respObj); err != nil {
		c.Logger.Warn("Failed to decode response from Pinecone",
			zap.Error(err),
		)

		return errmsg.AddMessage(
			fmt.Errorf("failed to decode response from pinecone: %w", err),
			"Failed to decode response from Pinecone's API.",
		)
	}

	return nil
}

func (e *Execution) Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	client := NewClient(getAPIKey(e.Config), e.Logger)
	outputs := []*structpb.Struct{}

	for _, input := range inputs {
		var output *structpb.Struct
		switch e.Task {
		case taskQuery:
			inputStruct := QueryInput{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}

			// Each query request can contain only one of the parameters
			// vector, or id.
			// Ref: https://docs.pinecone.io/reference/query
			if inputStruct.ID != "" {
				inputStruct.Vector = nil
			}

			url := getURL(e.Config) + "/query"
			resp := QueryResp{}
			err = client.sendReqAndUnmarshal(url, http.MethodPost, QueryReq(inputStruct), &resp)
			if err != nil {
				return nil, err
			}
			output, err = base.ConvertToStructpb(resp)
			if err != nil {
				return nil, err
			}
		case taskUpsert:
			vector := Vector{}
			err := base.ConvertFromStructpb(input, &vector)
			if err != nil {
				return nil, err
			}
			url := getURL(e.Config) + "/vectors/upsert"
			resp := UpsertResp{}
			inputStruct := UpsertReq{
				Vectors: []Vector{vector},
			}
			err = client.sendReqAndUnmarshal(url, http.MethodPost, inputStruct, &resp)
			if err != nil {
				return nil, err
			}

			output, err = base.ConvertToStructpb(UpsertOutput(resp))
			if err != nil {
				return nil, err
			}
		}
		outputs = append(outputs, output)
	}
	return outputs, nil
}

func (c *Connector) Test(defUid uuid.UUID, config *structpb.Struct, logger *zap.Logger) (pipelinePB.Connector_State, error) {
	//TODO: change this
	return pipelinePB.Connector_STATE_CONNECTED, nil
}
