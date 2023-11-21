package openai

import (
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"

	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1alpha"
)

const (
	venderName            = "openAI"
	host                  = "https://api.openai.com"
	jsonMimeType          = "application/json"
	reqTimeout            = time.Second * 60 * 5
	textGenerationTask    = "TASK_TEXT_GENERATION"
	textEmbeddingsTask    = "TASK_TEXT_EMBEDDINGS"
	speechRecognitionTask = "TASK_SPEECH_RECOGNITION"
	textToSpeechTask      = "TASK_TEXT_TO_SPEECH"
	textToImageTask       = "TASK_TEXT_TO_IMAGE"
)

var (
	//go:embed config/definitions.json
	definitionsJSON []byte
	//go:embed config/tasks.json
	tasksJSON []byte
	//go:embed config/openai.json
	openAIJSON []byte

	once      sync.Once
	connector base.IConnector
)

type Connector struct {
	base.Connector
}

type Execution struct {
	base.Execution
}

// Client represents a OpenAI client
type Client struct {
	APIKey     string
	Org        string
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
		err := connector.LoadConnectorDefinitions(definitionsJSON, tasksJSON, map[string][]byte{"openai.json": openAIJSON})
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

// NewClient initializes a new OpenAI client
func NewClient(apiKey, org string) Client {
	tr := &http.Transport{
		DisableKeepAlives: true,
	}
	return Client{APIKey: apiKey, Org: org, HTTPClient: &http.Client{Timeout: reqTimeout, Transport: tr}}
}

// sendReq is responsible for making the http request with to given URL, method, and params
func (c *Client) sendReq(reqURL, method, contentType string, data io.Reader) ([]byte, error) {
	req, _ := http.NewRequest(method, reqURL, data)
	req.Header.Add("Content-Type", contentType)
	req.Header.Add("Accept", jsonMimeType)
	req.Header.Add("Authorization", "Bearer "+c.APIKey)
	if c.Org != "" {
		req.Header.Add("OpenAI-Organization", c.Org)
	}
	http.DefaultClient.Timeout = reqTimeout
	res, err := c.HTTPClient.Do(req)
	if res != nil && res.Body != nil {
		defer res.Body.Close()
	}
	if err != nil || res == nil {
		err = fmt.Errorf("error occurred: %v, while calling URL: %s", err, reqURL)
		return nil, err
	}
	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("non-200 status code: %d, while calling URL: %s, response body: %s", res.StatusCode, reqURL, respBody)
		return nil, err
	}
	return respBody, nil
}

// sendReqAndUnmarshal is responsible for making the http request with to given URL, method, and params and unmarshalling the response into given object.
func (c *Client) sendReqAndUnmarshal(reqURL, method, contentType string, data io.Reader, respObj interface{}) error {
	respBody, err := c.sendReq(reqURL, method, contentType, data)
	if err != nil {
		return err
	}
	err = json.Unmarshal(respBody, &respObj)
	if err != nil {
		return fmt.Errorf("error in json decode: %s, while calling URL: %s, response body: %s", err, reqURL, respBody)
	}
	return nil
}

func getAPIKey(config *structpb.Struct) string {
	return config.GetFields()["api_key"].GetStringValue()
}

func getOrg(config *structpb.Struct) string {
	val, ok := config.GetFields()["organization"]
	if !ok {
		return ""
	}
	return val.GetStringValue()
}

func (e *Execution) Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error) {

	client := NewClient(getAPIKey(e.Config), getOrg(e.Config))

	outputs := []*structpb.Struct{}

	for _, input := range inputs {
		switch e.Task {
		case textGenerationTask:

			inputStruct := TextCompletionInput{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}

			messages := []Message{}

			// If chat history is provided, add it to the messages, and ignore the system message
			if inputStruct.ChatHistory != nil {
				for _, textMessage := range inputStruct.ChatHistory {
					messages = append(messages, Message{Role: textMessage.Role, Content: []Content{{Type: "text", Text: &textMessage.Content}}})
				}
			} else {
				// If chat history is not provided, add the system message to the messages
				if inputStruct.SystemMessage != nil {
					messages = append(messages, Message{Role: "system", Content: []Content{{Type: "text", Text: inputStruct.SystemMessage}}})
				}
			}
			userContents := []Content{}
			userContents = append(userContents, Content{Type: "text", Text: &inputStruct.Prompt})
			for _, image := range inputStruct.Images {
				b, err := base64.StdEncoding.DecodeString(base.TrimBase64Mime(image))
				if err != nil {
					return nil, err
				}
				url := fmt.Sprintf("data:%s;base64,%s", mimetype.Detect(b).String(), base.TrimBase64Mime(image))
				userContents = append(userContents, Content{Type: "image_url", ImageUrl: &ImageUrl{Url: url}})
			}
			messages = append(messages, Message{Role: "user", Content: userContents})

			req := TextCompletionReq{
				Messages:         messages,
				Model:            inputStruct.Model,
				MaxTokens:        inputStruct.MaxTokens,
				Temperature:      inputStruct.Temperature,
				N:                inputStruct.N,
				TopP:             inputStruct.TopP,
				PresencePenalty:  inputStruct.PresencePenalty,
				FrequencyPenalty: inputStruct.FrequencyPenalty,
			}

			// workaround, the OpenAI service can not accept this param
			if inputStruct.Model != "gpt-4-vision-preview" {
				req.ResponseFormat = inputStruct.ResponseFormat
			}

			resp, err := client.GenerateTextCompletion(req)
			if err != nil {
				return inputs, err
			}
			outputStruct := TextCompletionOutput{
				Texts: []string{},
			}
			for _, c := range resp.Choices {
				outputStruct.Texts = append(outputStruct.Texts, c.Message.Content)
			}

			outputJson, err := json.Marshal(outputStruct)
			if err != nil {
				return nil, err
			}
			output := structpb.Struct{}
			err = protojson.Unmarshal(outputJson, &output)
			if err != nil {
				return nil, err
			}
			outputs = append(outputs, &output)

		case textEmbeddingsTask:

			inputStruct := TextEmbeddingsInput{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}

			req := TextEmbeddingsReq{
				Model: inputStruct.Model,
				Input: []string{inputStruct.Text},
			}
			resp, err := client.GenerateTextEmbeddings(req)
			if err != nil {
				return inputs, err
			}

			outputStruct := TextEmbeddingsOutput{
				Embedding: resp.Data[0].Embedding,
			}

			output, err := base.ConvertToStructpb(outputStruct)
			if err != nil {
				return nil, err
			}
			outputs = append(outputs, output)

		case speechRecognitionTask:

			inputStruct := AudioTranscriptionInput{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}

			audioBytes, err := base64.StdEncoding.DecodeString(base.TrimBase64Mime(inputStruct.Audio))
			if err != nil {
				return nil, err
			}
			req := AudioTranscriptionReq{
				File:        audioBytes,
				Model:       inputStruct.Model,
				Prompt:      inputStruct.Prompt,
				Language:    inputStruct.Prompt,
				Temperature: inputStruct.Temperature,
			}

			resp, err := client.GenerateAudioTranscriptions(req)
			if err != nil {
				return inputs, err
			}

			output, err := base.ConvertToStructpb(resp)
			if err != nil {
				return nil, err
			}
			outputs = append(outputs, output)

		case textToSpeechTask:

			inputStruct := TextToSpeechInput{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}

			req := TextToSpeechReq{
				Input:          inputStruct.Text,
				Model:          inputStruct.Model,
				Voice:          inputStruct.Voice,
				ResponseFormat: inputStruct.ResponseFormat,
				Speed:          inputStruct.Speed,
			}

			outputStruct, err := client.CreateSpeech(req)
			if err != nil {
				return inputs, err
			}

			output, err := base.ConvertToStructpb(outputStruct)
			if err != nil {
				return nil, err
			}
			outputs = append(outputs, output)

		case textToImageTask:

			inputStruct := ImagesGenerationInput{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}

			req := ImageGenerationsReq{
				Model:          inputStruct.Model,
				Prompt:         inputStruct.Prompt,
				Quality:        inputStruct.Quality,
				Size:           inputStruct.Size,
				Style:          inputStruct.Style,
				N:              inputStruct.N,
				ResponseFormat: "b64_json",
			}

			resp, err := client.GenerateImagesGenerations(req)
			if err != nil {
				return inputs, err
			}

			results := []ImageGenerationsOutputResult{}
			for _, data := range resp.Data {
				results = append(results, ImageGenerationsOutputResult(data))
			}
			outputStruct := ImageGenerationsOutput{
				Results: results,
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
	client := NewClient(getAPIKey(config), getOrg(config))
	models, err := client.ListModels()
	if err != nil {
		return pipelinePB.Connector_STATE_ERROR, err
	}
	if len(models.Data) == 0 {
		return pipelinePB.Connector_STATE_DISCONNECTED, nil
	}
	return pipelinePB.Connector_STATE_CONNECTED, nil
}
