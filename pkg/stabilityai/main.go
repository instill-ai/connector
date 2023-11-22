package stabilityai

import (
	_ "embed"
	"encoding/base64"
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

	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1alpha"
)

const (
	host             = "https://api.stability.ai"
	jsonMimeType     = "application/json"
	reqTimeout       = time.Second * 60 * 5
	textToImageTask  = "TASK_TEXT_TO_IMAGE"
	imageToImageTask = "TASK_IMAGE_TO_IMAGE"
)

var (
	//go:embed config/definitions.json
	definitionsJSON []byte
	//go:embed config/tasks.json
	tasksJSON []byte
	//go:embed config/stabilityai.json
	stabilityaiJSON []byte
	once            sync.Once
	connector       base.IConnector
)

type Connector struct {
	base.Connector
}

type Execution struct {
	base.Execution
}

// Client represents a Stability AI client
type Client struct {
	APIKey     string
	HTTPClient HTTPClient
}

// HTTPClient interface
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func Init(logger *zap.Logger) base.IConnector {
	once.Do(func() {
		connector = &Connector{
			Connector: base.Connector{
				Component: base.Component{Logger: logger},
			},
		}
		err := connector.LoadConnectorDefinitions(definitionsJSON, tasksJSON, map[string][]byte{"stabilityai.json": stabilityaiJSON})
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

// NewClient initializes a new Stability AI client
func NewClient(apiKey string) Client {
	tr := &http.Transport{
		DisableKeepAlives: true,
	}
	return Client{APIKey: apiKey, HTTPClient: &http.Client{Timeout: reqTimeout, Transport: tr}}
}

// sendReq is responsible for making the http request with to given URL, method, and params and unmarshalling the response into given object.
func (c *Client) sendReq(reqURL, method, contentType string, data io.Reader, respObj interface{}) (err error) {
	req, _ := http.NewRequest(method, reqURL, data)
	req.Header.Add("Content-Type", contentType)
	req.Header.Add("Accept", jsonMimeType)
	req.Header.Add("Authorization", "Bearer "+c.APIKey)
	http.DefaultClient.Timeout = reqTimeout
	res, err := c.HTTPClient.Do(req)
	if res != nil && res.Body != nil {
		defer res.Body.Close()
	}
	if err != nil || res == nil {
		err = fmt.Errorf("error occurred: %v, while calling URL: %s, request body: %s", err, reqURL, data)
		return
	}
	bytes, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("non-200 status code: %d, while calling URL: %s, response body: %s", res.StatusCode, reqURL, bytes)
		return
	}
	if err = json.Unmarshal(bytes, &respObj); err != nil {
		err = fmt.Errorf("error in json decode: %s, while calling URL: %s, response body: %s", err, reqURL, bytes)
	}
	return
}

func getAPIKey(config *structpb.Struct) string {
	return config.GetFields()["api_key"].GetStringValue()
}

func (e *Execution) Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error) {

	client := NewClient(getAPIKey(e.Config))

	outputs := []*structpb.Struct{}

	for _, input := range inputs {
		switch e.Task {
		case textToImageTask:

			inputStruct := TextToImageInput{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}

			noOfPrompts := len(inputStruct.Prompts)
			if noOfPrompts <= 0 {
				return inputs, fmt.Errorf("no text promts given")
			}
			req := TextToImageReq{
				CFGScale:           inputStruct.CfgScale,
				ClipGuidancePreset: inputStruct.ClipGuidancePreset,
				Sampler:            inputStruct.Sampler,
				Samples:            inputStruct.Samples,
				Seed:               inputStruct.Seed,
				Steps:              inputStruct.Steps,
				StylePreset:        inputStruct.StylePreset,
				Height:             inputStruct.Height,
				Width:              inputStruct.Width,
			}

			req.TextPrompts = make([]TextPrompt, 0, noOfPrompts)
			var w float64
			for index, t := range inputStruct.Prompts {
				if inputStruct.Weights != nil && len(*inputStruct.Weights) > index {
					w = (*inputStruct.Weights)[index]
				} else {
					// If weights are not provided, set all weights to 1.0
					w = 1.0
				}
				req.TextPrompts = append(req.TextPrompts, TextPrompt{Text: t, Weight: &w})
			}
			images, err := client.GenerateImageFromText(req, inputStruct.Engine)
			if err != nil {
				return inputs, err
			}

			outputStruct := TextToImageOutput{
				Images: []string{},
				Seeds:  []uint32{},
			}

			for _, image := range images {
				outputStruct.Images = append(outputStruct.Images, image.Base64)
				outputStruct.Seeds = append(outputStruct.Seeds, image.Seed)
			}
			output, err := base.ConvertToStructpb(outputStruct)
			if err != nil {
				return nil, err
			}

			outputs = append(outputs, output)

		case imageToImageTask:

			inputStruct := ImageToImageInput{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}

			noOfPrompts := len(inputStruct.Prompts)
			if noOfPrompts <= 0 {
				return inputs, fmt.Errorf("no text promts given")
			}

			req := ImageToImageReq{
				InitImage:          inputStruct.InitImage,
				InitImageMode:      inputStruct.InitImageMode,
				ImageStrength:      inputStruct.ImageStrength,
				StepScheduleStart:  inputStruct.StepScheduleStart,
				StepScheduleEnd:    inputStruct.StepScheduleEnd,
				CFGScale:           inputStruct.CfgScale,
				ClipGuidancePreset: inputStruct.ClipGuidancePreset,
				Sampler:            inputStruct.Sampler,
				Samples:            inputStruct.Samples,
				Seed:               inputStruct.Seed,
				Steps:              inputStruct.Steps,
				StylePreset:        inputStruct.StylePreset,
			}

			req.TextPrompts = make([]TextPrompt, 0, noOfPrompts)
			var w float64
			for index, t := range inputStruct.Prompts {
				if inputStruct.Weights != nil && len(*inputStruct.Weights) > index {
					w = (*inputStruct.Weights)[index]
				}
				req.TextPrompts = append(req.TextPrompts, TextPrompt{Text: t, Weight: &w})
			}
			images, err := client.GenerateImageFromImage(req, inputStruct.Engine)
			if err != nil {
				return inputs, err
			}
			outputStruct := TextToImageOutput{
				Images: []string{},
				Seeds:  []uint32{},
			}

			for _, image := range images {
				outputStruct.Images = append(outputStruct.Images, image.Base64)
				outputStruct.Seeds = append(outputStruct.Seeds, image.Seed)
			}
			output, err := base.ConvertToStructpb(outputStruct)
			if err != nil {
				return nil, err
			}
			outputs = append(outputs, output)

		default:
			return nil, fmt.Errorf("not supported task: %s", e.Task)
		}
	}
	return outputs, nil
}

func (c *Connector) Test(defUid uuid.UUID, config *structpb.Struct, logger *zap.Logger) (pipelinePB.Connector_State, error) {
	client := NewClient(getAPIKey(config))
	engines, err := client.ListEngines()
	if err != nil {
		return pipelinePB.Connector_STATE_ERROR, err
	}
	if len(engines) == 0 {
		return pipelinePB.Connector_STATE_DISCONNECTED, nil
	}
	return pipelinePB.Connector_STATE_CONNECTED, nil
}

// decode if the string is base64 encoded
func DecodeBase64(input string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(base.TrimBase64Mime(input))
}
