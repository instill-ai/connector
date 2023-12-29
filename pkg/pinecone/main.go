package pinecone

import (
	_ "embed"
	"errors"
	"fmt"
	"net/url"
	"sync"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"
	"github.com/instill-ai/connector/pkg/util/httpclient"
	"github.com/instill-ai/x/errmsg"

	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1beta"
)

const (
	taskQuery  = "TASK_QUERY"
	taskUpsert = "TASK_UPSERT"

	upsertPath = "/vectors/upsert"
	queryPath  = "/query"
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

func newClient(config *structpb.Struct, logger *zap.Logger) *httpclient.Client {
	c := httpclient.New("Pinecone", getURL(config),
		httpclient.WithLogger(logger),
		httpclient.WithEndUserError(new(errBody)),
	)

	c.SetHeader("Api-Key", getAPIKey(config))

	return c
}

func getAPIKey(config *structpb.Struct) string {
	return config.GetFields()["api_key"].GetStringValue()
}

func getURL(config *structpb.Struct) string {
	return config.GetFields()["url"].GetStringValue()
}

// The HTTP client doesn't provide a hook for errors in `http.Client.Do`, e.g.
// if the connector configuration contains an invalid URL. This wrapper adds an
// end-user error in such cases.
func wrapURLError(err error) error {
	uerr := new(url.Error)
	if errors.As(err, &uerr) {
		err = errmsg.AddMessage(
			err,
			fmt.Sprintf("Failed to call %s. Please check that the connector configuration is correct.", uerr.URL),
		)
	}

	return err
}

func (e *Execution) Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	req := newClient(e.Config, e.Logger).R()
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

			resp := QueryResp{}
			req.SetResult(&resp).SetBody(QueryReq(inputStruct))

			if _, err := req.Post(queryPath); err != nil {
				return nil, wrapURLError(err)
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

			resp := UpsertResp{}
			req.SetResult(&resp).SetBody(UpsertReq{
				Vectors: []Vector{vector},
			})

			if _, err := req.Post(upsertPath); err != nil {
				return nil, wrapURLError(err)
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
