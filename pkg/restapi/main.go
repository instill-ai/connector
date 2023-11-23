package restapi

import (
	_ "embed"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/instill-ai/component/pkg/base"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1alpha"
)

const (
	taskGet     = "TASK_GET"
	taskPost    = "TASK_POST"
	taskPatch   = "TASK_PATCH"
	taskPut     = "TASK_PUT"
	taskDelete  = "TASK_DELETE"
	taskHead    = "TASK_HEAD"
	taskOptions = "TASK_OPTIONS"
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

func getBaseURL(config *structpb.Struct) string {
	return config.GetFields()["base_url"].GetStringValue()
}

func getAuthentication(config *structpb.Struct) (Authentication, error) {
	auth := config.GetFields()["authentication"].GetStructValue()
	authType := auth.GetFields()["auth_type"].GetStringValue()

	switch authType {
	case string(NoAuthType):
		authStruct := NoAuth{}
		err := base.ConvertFromStructpb(auth, &authStruct)
		if err != nil {
			return nil, err
		}
		return authStruct, nil
	case string(BasicAuthType):
		authStruct := BasicAuth{}
		err := base.ConvertFromStructpb(auth, &authStruct)
		if err != nil {
			return nil, err
		}
		return authStruct, nil
	case string(APIKeyType):
		authStruct := APIKeyAuth{}
		err := base.ConvertFromStructpb(auth, &authStruct)
		if err != nil {
			return nil, err
		}
		return authStruct, nil
	case string(BearerTokenType):
		authStruct := BearerTokenAuth{}
		err := base.ConvertFromStructpb(auth, &authStruct)
		if err != nil {
			return nil, err
		}
		return authStruct, nil
	default:
		return nil, errors.New("invalid authentication type")
	}
}

func (e *Execution) Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error) {

	client, err := NewClient(e.Config)
	if err != nil {
		return nil, err
	}

	outputs := []*structpb.Struct{}
	for _, input := range inputs {
		var method string
		inputStruct := TaskInput{}
		switch e.Task {
		case taskGet:
			method = http.MethodGet
		case taskPost:
			method = http.MethodPost
		case taskPatch:
			method = http.MethodPatch
		case taskPut:
			method = http.MethodPut
		case taskDelete:
			method = http.MethodDelete
		case taskHead:
			method = http.MethodHead
		case taskOptions:
			method = http.MethodOptions
		default:
			return nil, fmt.Errorf("not supported task: %s", e.Task)
		}

		err := base.ConvertFromStructpb(input, &inputStruct)
		if err != nil {
			return nil, err
		}
		outputStruct, err := client.sendRequest(method, inputStruct)
		if err != nil {
			return nil, err
		}
		output, err := base.ConvertToStructpb(outputStruct)
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, output)
	}
	return outputs, nil
}

func (c *Connector) Test(defUid uuid.UUID, config *structpb.Struct, logger *zap.Logger) (pipelinePB.Connector_State, error) {
	client, err := NewClient(config)
	if err != nil {
		return pipelinePB.Connector_STATE_ERROR, err
	}
	_, err = client.sendRequest(http.MethodGet, TaskInput{})
	if err != nil {
		return pipelinePB.Connector_STATE_ERROR, err
	}
	return pipelinePB.Connector_STATE_CONNECTED, nil

}
