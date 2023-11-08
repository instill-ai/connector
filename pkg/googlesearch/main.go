package googlesearch

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/api/customsearch/v1"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"

	connectorPB "github.com/instill-ai/protogen-go/vdp/connector/v1alpha"
)

const (
	taskSearch = "TASK_SEARCH"
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

// NewService creates a Google custom search service
func NewService(apiKey string) (*customsearch.Service, error) {
	return customsearch.NewService(context.Background(), option.WithAPIKey(apiKey))
}

func getAPIKey(config *structpb.Struct) string {
	return config.GetFields()["api_key"].GetStringValue()
}

func getSearchEngineID(config *structpb.Struct) string {
	return config.GetFields()["cse_id"].GetStringValue()
}

func (e *Execution) Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error) {

	service, err := NewService(getAPIKey(e.Config))
	if err != nil || service == nil {
		return nil, fmt.Errorf("error creating Google custom search service: %v", err)
	}
	cseID := getSearchEngineID(e.Config)

	outputs := []*structpb.Struct{}

	for _, input := range inputs {
		switch e.Task {
		case taskSearch:

			inputStruct := SearchInput{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}

			// Make the search request
			results, err := search(service, cseID, inputStruct.Query, int64(*inputStruct.TopK), *inputStruct.IncludeLinkText, *inputStruct.IncludeLinkHtml)

			if err != nil {
				return nil, err
			}

			outputStruct := SearchOutput{
				Results: results,
			}

			outputJson, err := json.Marshal(outputStruct)
			if err != nil {
				return nil, err
			}
			output := structpb.Struct{}
			err = json.Unmarshal(outputJson, &output)
			if err != nil {
				return nil, err
			}
			outputs = append(outputs, &output)

		default:
			return nil, fmt.Errorf("not supported task: %s", e.Task)
		}
	}

	return outputs, nil
}

func (c *Connector) Test(defUid uuid.UUID, config *structpb.Struct, logger *zap.Logger) (connectorPB.ConnectorResource_State, error) {

	service, err := NewService(getAPIKey(config))
	if err != nil || service == nil {
		return connectorPB.ConnectorResource_STATE_ERROR, fmt.Errorf("error creating Google custom search service: %v", err)
	}
	if service == nil {
		return connectorPB.ConnectorResource_STATE_ERROR, fmt.Errorf("error creating Google custom search service: %v", err)
	}
	return connectorPB.ConnectorResource_STATE_CONNECTED, nil
}
