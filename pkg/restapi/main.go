package restapi

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/instill-ai/component/pkg/base"
	"github.com/instill-ai/x/errmsg"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"

	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1beta"
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

var (
	//go:embed config/definitions.json
	definitionsJSON []byte

	//go:embed config/tasks.json
	tasksJSON []byte

	once      sync.Once
	connector base.IConnector

	taskMethod = map[string]string{
		taskGet:     http.MethodGet,
		taskPost:    http.MethodPost,
		taskPatch:   http.MethodPatch,
		taskPut:     http.MethodPut,
		taskDelete:  http.MethodDelete,
		taskHead:    http.MethodHead,
		taskOptions: http.MethodOptions,
	}
)

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

func getAuthentication(config *structpb.Struct) (authentication, error) {
	auth := config.GetFields()["authentication"].GetStructValue()
	authType := auth.GetFields()["auth_type"].GetStringValue()

	switch authType {
	case string(noAuthType):
		authStruct := noAuth{}
		err := base.ConvertFromStructpb(auth, &authStruct)
		if err != nil {
			return nil, err
		}
		return authStruct, nil
	case string(basicAuthType):
		authStruct := basicAuth{}
		err := base.ConvertFromStructpb(auth, &authStruct)
		if err != nil {
			return nil, err
		}
		return authStruct, nil
	case string(apiKeyType):
		authStruct := apiKeyAuth{}
		err := base.ConvertFromStructpb(auth, &authStruct)
		if err != nil {
			return nil, err
		}
		return authStruct, nil
	case string(bearerTokenType):
		authStruct := bearerTokenAuth{}
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
	client, err := newClient(e.Config, e.Logger)
	if err != nil {
		return nil, err
	}

	method, ok := taskMethod[e.Task]
	if !ok {
		return nil, errmsg.AddMessage(
			fmt.Errorf("not supported task: %s", e.Task),
			fmt.Sprintf("%s task is not supported.", e.Task),
		)
	}

	outputs := []*structpb.Struct{}
	for _, input := range inputs {
		taskIn := TaskInput{}
		taskOut := TaskOutput{}
		path := ""

		if err := base.ConvertFromStructpb(input, &taskIn); err != nil {
			return nil, err
		}

		if taskIn.EndpointPath != nil {
			path = *taskIn.EndpointPath
		}

		// An API error is a valid output in this connector.
		req := client.R().SetResult(&taskOut.Body).SetError(&taskOut.Body)
		if taskIn.Body != nil {
			req.SetBody(taskIn.Body)
		}

		resp, err := req.Execute(method, path)
		if err != nil {
			return nil, err
		}

		taskOut.StatusCode = resp.StatusCode()
		taskOut.Header = resp.Header()

		output, err := base.ConvertToStructpb(taskOut)
		if err != nil {
			return nil, err
		}

		outputs = append(outputs, output)
	}
	return outputs, nil
}

func (c *Connector) Test(defUid uuid.UUID, config *structpb.Struct, logger *zap.Logger) (pipelinePB.Connector_State, error) {
	client, err := newClient(config, logger)
	if err != nil {
		return pipelinePB.Connector_STATE_ERROR, err
	}

	if _, err := client.R().Get(""); err != nil {
		return pipelinePB.Connector_STATE_ERROR, err
	}

	return pipelinePB.Connector_STATE_CONNECTED, nil
}

func (c *Connector) GetConnectorDefinitionByID(defID string, resourceConfig *structpb.Struct, componentConfig *structpb.Struct) (*pipelinePB.ConnectorDefinition, error) {
	def, err := c.Connector.GetConnectorDefinitionByID(defID, resourceConfig, componentConfig)
	if err != nil {
		return nil, err
	}

	return c.GetConnectorDefinitionByUID(uuid.FromStringOrNil(def.Uid), resourceConfig, componentConfig)
}

// Generate the model_name enum based on the task
func (c *Connector) GetConnectorDefinitionByUID(defUID uuid.UUID, resourceConfig *structpb.Struct, componentConfig *structpb.Struct) (*pipelinePB.ConnectorDefinition, error) {
	oriDef, err := c.Connector.GetConnectorDefinitionByUID(defUID, resourceConfig, componentConfig)
	if err != nil {
		return nil, err
	}

	def := proto.Clone(oriDef).(*pipelinePB.ConnectorDefinition)
	if componentConfig == nil {
		return def, nil
	}
	if _, ok := componentConfig.Fields["task"]; !ok {
		return def, nil
	}
	if _, ok := componentConfig.Fields["input"]; !ok {
		return def, nil
	}
	if _, ok := componentConfig.Fields["input"].GetStructValue().Fields["output_body_schema"]; !ok {
		return def, nil
	}

	task := componentConfig.Fields["task"].GetStringValue()
	schStr := componentConfig.Fields["input"].GetStructValue().Fields["output_body_schema"].GetStringValue()
	sch := &structpb.Struct{}
	_ = json.Unmarshal([]byte(schStr), sch)
	spec := def.Spec.OpenapiSpecifications
	walk := spec.Fields[task]
	for _, key := range []string{"paths", "/execute", "post", "responses", "200", "content", "application/json", "schema", "properties", "outputs", "items", "properties", "body"} {
		if _, ok := walk.GetStructValue().Fields[key]; !ok {
			return def, nil
		}
		walk = walk.GetStructValue().Fields[key]
	}
	*walk = *structpb.NewStructValue(sch)
	return def, nil
}
