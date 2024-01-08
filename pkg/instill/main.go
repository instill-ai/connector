package instill

import (
	"crypto/tls"
	_ "embed"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"
	"github.com/instill-ai/connector/pkg/util/httpclient"

	commonPB "github.com/instill-ai/protogen-go/common/task/v1alpha"
	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1beta"
)

const (
	getModelPath = "/v1alpha/models"
	internalMode = "Internal Mode"
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

func getMode(config *structpb.Struct) string {
	return config.GetFields()["mode"].GetStringValue()
}

func getAPIKey(config *structpb.Struct) string {
	return config.GetFields()["api_token"].GetStringValue()
}
func getInstillUserUid(config *structpb.Struct) string {
	return config.GetFields()["instill_user_uid"].GetStringValue()
}

func getServerURL(config *structpb.Struct) string {
	if getMode(config) == internalMode {
		return config.GetFields()["instill_model_backend"].GetStringValue()
	}
	serverUrl := config.GetFields()["server_url"].GetStringValue()
	if strings.HasPrefix(serverUrl, "https://") {
		if len(strings.Split(serverUrl, ":")) == 2 {
			serverUrl = serverUrl + ":443"
		}
	} else if strings.HasPrefix(serverUrl, "http://") {
		if len(strings.Split(serverUrl, ":")) == 2 {
			serverUrl = serverUrl + ":80"
		}
	}
	return serverUrl
}

func (e *Execution) Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	var err error

	if len(inputs) <= 0 || inputs[0] == nil {
		return inputs, fmt.Errorf("invalid input")
	}

	gRPCCLient, gRPCCLientConn := initModelPublicServiceClient(getServerURL(e.Config))
	if gRPCCLientConn != nil {
		defer gRPCCLientConn.Close()
	}

	modelNamespace := inputs[0].GetFields()["model_namespace"].GetStringValue()
	modelId := inputs[0].GetFields()["model_id"].GetStringValue()
	modelName := fmt.Sprintf("users/%s/models/%s", modelNamespace, modelId)

	var result []*structpb.Struct
	switch e.Task {
	case commonPB.Task_TASK_UNSPECIFIED.String():
		result, err = e.executeUnspecified(gRPCCLient, modelName, inputs)
	case commonPB.Task_TASK_CLASSIFICATION.String():
		result, err = e.executeImageClassification(gRPCCLient, modelName, inputs)
	case commonPB.Task_TASK_DETECTION.String():
		result, err = e.executeObjectDetection(gRPCCLient, modelName, inputs)
	case commonPB.Task_TASK_KEYPOINT.String():
		result, err = e.executeKeyPointDetection(gRPCCLient, modelName, inputs)
	case commonPB.Task_TASK_OCR.String():
		result, err = e.executeOCR(gRPCCLient, modelName, inputs)
	case commonPB.Task_TASK_INSTANCE_SEGMENTATION.String():
		result, err = e.executeInstanceSegmentation(gRPCCLient, modelName, inputs)
	case commonPB.Task_TASK_SEMANTIC_SEGMENTATION.String():
		result, err = e.executeSemanticSegmentation(gRPCCLient, modelName, inputs)
	case commonPB.Task_TASK_TEXT_TO_IMAGE.String():
		result, err = e.executeTextToImage(gRPCCLient, modelName, inputs)
	case commonPB.Task_TASK_TEXT_GENERATION.String():
		result, err = e.executeTextGeneration(gRPCCLient, modelName, inputs)
	case commonPB.Task_TASK_TEXT_GENERATION_CHAT.String():
		result, err = e.executeTextGenerationChat(gRPCCLient, modelName, inputs)
	case commonPB.Task_TASK_VISUAL_QUESTION_ANSWERING.String():
		result, err = e.executeVisualQuestionAnswering(gRPCCLient, modelName, inputs)
	case commonPB.Task_TASK_IMAGE_TO_IMAGE.String():
		result, err = e.executeImageToImage(gRPCCLient, modelName, inputs)
	default:
		return inputs, fmt.Errorf("unsupported task: %s", e.Task)
	}

	return result, err
}

func (c *Connector) Test(_ uuid.UUID, config *structpb.Struct, logger *zap.Logger) (pipelinePB.Connector_State, error) {
	req := newHTTPClient(config, logger).R()

	path := "/model" + getModelPath
	if resp, err := req.Get(path); err != nil || resp.IsError() {
		return pipelinePB.Connector_STATE_ERROR, err
	}

	return pipelinePB.Connector_STATE_CONNECTED, nil
}

type errBody struct {
	Msg string `json:"message"`
}

func (e errBody) Message() string {
	return e.Msg
}

func newHTTPClient(config *structpb.Struct, logger *zap.Logger) *httpclient.Client {
	c := httpclient.New("Instill AI", getServerURL(config),
		httpclient.WithLogger(logger),
		httpclient.WithEndUserError(new(errBody)),
	)

	c.SetTransport(&http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives: true,
	})

	if token := getAPIKey(config); token != "" {
		c.SetAuthToken(token)
	}

	if userID := getInstillUserUid(config); userID != "" {
		c.SetHeader("Instill-User-Uid", userID)
	}

	return c
}
