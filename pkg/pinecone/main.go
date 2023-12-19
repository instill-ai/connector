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

	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1beta"
)

const (
	reqTimeout   = time.Second * 60 * 5
	taskQuery    = "TASK_QUERY"
	taskUpsert   = "TASK_UPSERT"
	jsonMimeType = "application/json"
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

// NewClient initializes a new Pinecone client
func NewClient(apiKey string) Client {
	tr := &http.Transport{
		DisableKeepAlives: true,
	}
	return Client{APIKey: apiKey, HTTPClient: &http.Client{Timeout: reqTimeout, Transport: tr}}
}

func getAPIKey(config *structpb.Struct) string {
	return config.GetFields()["api_key"].GetStringValue()
}

func getURL(config *structpb.Struct) string {
	return config.GetFields()["url"].GetStringValue()
}

// sendReq is responsible for making the http request with to given URL, method, and params and unmarshalling the response into given object.
func (c *Client) sendReq(reqURL, method string, params interface{}, respObj interface{}) error {
	var req *http.Request
	data, err := json.Marshal(params)
	if err != nil {
		return err
	}
	req, err = http.NewRequest(method, reqURL, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", jsonMimeType)
	req.Header.Add("Accept", jsonMimeType)
	req.Header.Add("Api-Key", c.APIKey)
	http.DefaultClient.Timeout = reqTimeout
	res, err := c.HTTPClient.Do(req)
	if res != nil && res.Body != nil {
		defer res.Body.Close()
	}
	if err != nil || res == nil {
		return fmt.Errorf("error occurred: %v, while calling URL: %s, request body: %s", err, reqURL, data)
	}
	bytes, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("non-200 status code: %d, while calling URL: %s, response body: %s", res.StatusCode, reqURL, bytes)
	}
	if err = json.Unmarshal(bytes, &respObj); err != nil {
		err = fmt.Errorf("error in json decode: %s, while calling URL: %s, response body: %s", err, reqURL, bytes)
	}
	return err
}

func (e *Execution) Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error) {

	client := NewClient(getAPIKey(e.Config))
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
			url := getURL(e.Config) + "/query"
			resp := QueryResp{}
			err = client.sendReq(url, http.MethodPost, QueryReq(inputStruct), &resp)
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
			err = client.sendReq(url, http.MethodPost, inputStruct, &resp)
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
