package redis

import (
	"context"
	_ "embed"
	"fmt"
	"sync"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	goredis "github.com/redis/go-redis/v9"

	"github.com/instill-ai/component/pkg/base"

	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1alpha"
)

const (
	taskChatMessageWrite    = "TASK_CHAT_MESSAGE_WRITE"
	taskChatHistoryRetrieve = "TASK_CHAT_HISTORY_RETRIEVE"
)

var (
	//go:embed config/definitions.json
	definitionsJSON []byte
	//go:embed config/tasks.json
	tasksJSON []byte

	once      sync.Once
	connector base.IConnector
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

func NewClient(host string, port int, username, password string) *goredis.Client {

	op := &goredis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: password,
		DB:       0,
	}
	if username != "" {
		op.Username = username
	}

	return goredis.NewClient(op)

	// TODO - add TLS support
	// TODO - add SSL support
}

func getHost(config *structpb.Struct) string {
	return config.GetFields()["host"].GetStringValue()
}
func getPort(config *structpb.Struct) int {
	return int(config.GetFields()["port"].GetNumberValue())
}
func getPassword(config *structpb.Struct) string {
	val, ok := config.GetFields()["password"]
	if !ok {
		return ""
	}
	return val.GetStringValue()
}
func getUsername(config *structpb.Struct) string {
	val, ok := config.GetFields()["username"]
	if !ok {
		return ""
	}
	return val.GetStringValue()
}

func (e *Execution) Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	outputs := []*structpb.Struct{}

	client := NewClient(getHost(e.Config), getPort(e.Config), getUsername(e.Config), getPassword(e.Config))
	defer client.Close()

	for _, input := range inputs {
		var output *structpb.Struct
		switch e.Task {
		case taskChatMessageWrite:
			inputStruct := ChatMessageWriteInput{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}
			outputStruct := WriteMessage(client, inputStruct)
			output, err = base.ConvertToStructpb(outputStruct)
			if err != nil {
				return nil, err
			}
		case taskChatHistoryRetrieve:
			inputStruct := ChatHistoryRetrieveInput{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}
			outputStruct := RetrieveSessionMessages(client, inputStruct)
			output, err = base.ConvertToStructpb(outputStruct)
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("unsupported task: %s", e.Task)
		}
		outputs = append(outputs, output)
	}
	return outputs, nil
}

func (c *Connector) Test(defUid uuid.UUID, config *structpb.Struct, logger *zap.Logger) (pipelinePB.Connector_State, error) {
	client := NewClient(getHost(config), getPort(config), getUsername(config), getPassword(config))
	defer client.Close()

	// Ping the Redis server to check the connection
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return pipelinePB.Connector_STATE_DISCONNECTED, err
	}
	return pipelinePB.Connector_STATE_CONNECTED, nil
}
