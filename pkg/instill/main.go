package instill

import (
	"crypto/tls"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"
	"github.com/instill-ai/connector/pkg/util/httpclient"

	commonPB "github.com/instill-ai/protogen-go/common/task/v1alpha"
	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1beta"
)

const (
	venderName   = "instillModel"
	getModelPath = "/v1alpha/models"
	internalMode = "Internal Mode"
	reqTimeout   = time.Second * 60
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
	client *Client
}

// Client represents an Instill Model client
type Client struct {
	APIKey         string
	InstillUserUid string
	HTTPClient     httpclient.Doer
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

// NewClient initializes a new Instill model client
func NewClient(config *structpb.Struct) (*Client, error) {
	tr := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives: true,
	}
	return &Client{APIKey: getAPIKey(config), InstillUserUid: getInstillUserUid(config), HTTPClient: &http.Client{Timeout: reqTimeout, Transport: tr}}, nil
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

func getModels(config *structpb.Struct) (err error) {
	serverURL := getServerURL(config) + "/model"
	client, err := NewClient(config)
	if err != nil {
		return err
	}
	reqURL := serverURL + getModelPath
	err = client.sendReq(reqURL, http.MethodGet, nil)
	return err
}

// sendReq is responsible for making the http request with to given URL, method, and params and unmarshalling the response into given object.
func (c *Client) sendReq(reqURL, method string, params interface{}) (err error) {
	req, _ := http.NewRequest(method, reqURL, nil)
	if c.APIKey != "" {
		req.Header.Add("Authorization", "Bearer "+c.APIKey)
	}
	if c.InstillUserUid != "" {
		req.Header.Add("Instill-User-Uid", c.InstillUserUid)
	}

	http.DefaultClient.Timeout = reqTimeout
	res, err := c.HTTPClient.Do(req)

	if err != nil || res == nil {
		err = fmt.Errorf("error occurred: %v, while calling URL: %s", err, reqURL)
		return
	}
	defer res.Body.Close()
	bytes, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("non-200 status code: %d, while calling URL: %s, response body: %s", res.StatusCode, reqURL, bytes)
		return
	}
	return
}

func (e *Execution) Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error) {

	var err error
	e.client, err = NewClient(e.Config)
	if err != nil {
		return nil, err
	}

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

func (c *Connector) Test(defUid uuid.UUID, config *structpb.Struct, logger *zap.Logger) (pipelinePB.Connector_State, error) {
	err := getModels(config)
	if err != nil {
		return pipelinePB.Connector_STATE_ERROR, err
	}
	return pipelinePB.Connector_STATE_CONNECTED, nil
}
