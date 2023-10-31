package huggingface

import (
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"

	connectorPB "github.com/instill-ai/protogen-go/vdp/connector/v1alpha"
)

const (
	venderName = "huggingface"
	reqTimeout = time.Second * 60 * 5
	//tasks
	textGenerationTask         = "TASK_TEXT_GENERATION"
	textToImageTask            = "TASK_TEXT_TO_IMAGE"
	fillMaskTask               = "TASK_FILL_MASK"
	summarizationTask          = "TASK_SUMMARIZATION"
	textClassificationTask     = "TASK_TEXT_CLASSIFICATION"
	tokenClassificationTask    = "TASK_TOKEN_CLASSIFICATION"
	translationTask            = "TASK_TRANSLATION"
	zeroShotClassificationTask = "TASK_ZERO_SHOT_CLASSIFICATION"
	featureExtractionTask      = "TASK_FEATURE_EXTRACTION"
	questionAnsweringTask      = "TASK_QUESTION_ANSWERING"
	tableQuestionAnsweringTask = "TASK_TABLE_QUESTION_ANSWERING"
	sentenceSimilarityTask     = "TASK_SENTENCE_SIMILARITY"
	conversationalTask         = "TASK_CONVERSATIONAL"
	imageClassificationTask    = "TASK_IMAGE_CLASSIFICATION"
	imageSegmentationTask      = "TASK_IMAGE_SEGMENTATION"
	objectDetectionTask        = "TASK_OBJECT_DETECTION"
	imageToTextTask            = "TASK_IMAGE_TO_TEXT"
	speechRecognitionTask      = "TASK_SPEECH_RECOGNITION"
	audioClassificationTask    = "TASK_AUDIO_CLASSIFICATION"
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

// Client represents a OpenAI client
type Client struct {
	APIKey           string
	BaseURL          string
	IsCustomEndpoint bool
	HTTPClient       HTTPClient
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

// NewClient initializes a new Hugging Face client
func NewClient(apiKey, baseURL string, isCustomEndpoint bool) Client {
	tr := &http.Transport{DisableKeepAlives: true}
	return Client{APIKey: apiKey, BaseURL: baseURL, IsCustomEndpoint: isCustomEndpoint, HTTPClient: &http.Client{Timeout: reqTimeout, Transport: tr}}
}

func getAPIKey(config *structpb.Struct) string {
	return config.GetFields()["api_key"].GetStringValue()
}

func getBaseURL(config *structpb.Struct) string {
	return config.GetFields()["base_url"].GetStringValue()
}

func isCustomEndpoint(config *structpb.Struct) bool {
	return config.GetFields()["is_custom_endpoint"].GetBoolValue()
}

func (e *Execution) Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	client := NewClient(getAPIKey(e.Config), getBaseURL(e.Config), isCustomEndpoint(e.Config))
	outputs := []*structpb.Struct{}
	model := inputs[0].GetFields()["model"].GetStringValue()

	for _, input := range inputs {
		switch e.Task {
		case textGenerationTask:
			inputStruct := TextGenerationRequest{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}
			jsonBody, _ := json.Marshal(inputStruct)
			resp, err := client.MakeHFAPIRequest(jsonBody, model)
			if err != nil {
				return nil, err
			}
			outputArr := []TextGenerationResponse{}
			err = json.Unmarshal(resp, &outputArr)
			if err != nil {
				return nil, err
			}
			output := structpb.Struct{Fields: make(map[string]*structpb.Value)}
			output.Fields["generated_text"] = structpb.NewStringValue(outputArr[0].GeneratedText)
			outputs = append(outputs, &output)
		case textToImageTask:
			inputStruct := TextToImageRequest{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}
			jsonBody, _ := json.Marshal(inputStruct)
			resp, err := client.MakeHFAPIRequest(jsonBody, model)
			if err != nil {
				return nil, err
			}
			outputStruct := TextToImageResponse{Image: base64.StdEncoding.EncodeToString(resp)}
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
		case fillMaskTask:
			inputStruct := FillMaskRequest{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}
			jsonBody, _ := json.Marshal(inputStruct)
			resp, err := client.MakeHFAPIRequest(jsonBody, model)
			if err != nil {
				return nil, err
			}
			outputArr := []FillMaskResponseEntry{}
			err = json.Unmarshal(resp, &outputArr)
			if err != nil {
				return nil, err
			}
			results := structpb.ListValue{}
			results.Values = make([]*structpb.Value, len(outputArr))
			for i := range outputArr {
				results.Values[i] = &structpb.Value{Kind: &structpb.Value_StructValue{
					StructValue: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"sequence":  {Kind: &structpb.Value_StringValue{StringValue: outputArr[i].Sequence}},
							"score":     {Kind: &structpb.Value_NumberValue{NumberValue: outputArr[i].Score}},
							"token":     {Kind: &structpb.Value_NumberValue{NumberValue: float64(outputArr[i].Token)}},
							"token_str": {Kind: &structpb.Value_StringValue{StringValue: outputArr[i].TokenStr}},
						},
					},
				}}
			}
			output := structpb.Struct{
				Fields: map[string]*structpb.Value{"results": {Kind: &structpb.Value_ListValue{ListValue: &results}}},
			}
			outputs = append(outputs, &output)
		case summarizationTask:
			inputStruct := SummarizationRequest{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}
			jsonBody, _ := json.Marshal(inputStruct)
			resp, err := client.MakeHFAPIRequest(jsonBody, model)
			if err != nil {
				return nil, err
			}
			outputArr := []SummarizationResponse{}
			err = json.Unmarshal(resp, &outputArr)
			if err != nil {
				return nil, err
			}
			output := structpb.Struct{Fields: make(map[string]*structpb.Value)}
			output.Fields["summary_text"] = structpb.NewStringValue(outputArr[0].SummaryText)
			outputs = append(outputs, &output)
		case textClassificationTask:
			inputStruct := TextClassificationRequest{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}
			jsonBody, _ := json.Marshal(inputStruct)
			resp, err := client.MakeHFAPIRequest(jsonBody, model)
			if err != nil {
				return nil, err
			}
			nestedArr := [][]ClassificationResponse{}
			err = json.Unmarshal(resp, &nestedArr)
			if err != nil {
				return nil, err
			}
			if len(nestedArr) <= 0 {
				return nil, errors.New("invalid response")
			}
			outputArr := nestedArr[0]
			results := structpb.ListValue{}
			results.Values = make([]*structpb.Value, len(outputArr))
			for i := range outputArr {
				results.Values[i] = &structpb.Value{Kind: &structpb.Value_StructValue{
					StructValue: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"label": {Kind: &structpb.Value_StringValue{StringValue: outputArr[i].Label}},
							"score": {Kind: &structpb.Value_NumberValue{NumberValue: outputArr[i].Score}},
						},
					},
				}}
			}
			output := structpb.Struct{
				Fields: map[string]*structpb.Value{"results": {Kind: &structpb.Value_ListValue{ListValue: &results}}},
			}
			outputs = append(outputs, &output)
		case tokenClassificationTask:
			inputStruct := TokenClassificationRequest{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}
			jsonBody, _ := json.Marshal(inputStruct)
			resp, err := client.MakeHFAPIRequest(jsonBody, model)
			if err != nil {
				return nil, err
			}
			outputArr := []TokenClassificationResponseEntity{}
			err = json.Unmarshal(resp, &outputArr)
			if err != nil {
				return nil, err
			}
			classes := structpb.ListValue{}
			classes.Values = make([]*structpb.Value, len(outputArr))
			for i := range outputArr {
				classes.Values[i] = &structpb.Value{Kind: &structpb.Value_StructValue{
					StructValue: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"entity_group": {Kind: &structpb.Value_StringValue{StringValue: outputArr[i].EntityGroup}},
							"score":        {Kind: &structpb.Value_NumberValue{NumberValue: outputArr[i].Score}},
							"word":         {Kind: &structpb.Value_StringValue{StringValue: outputArr[i].Word}},
							"start":        {Kind: &structpb.Value_NumberValue{NumberValue: float64(outputArr[i].Start)}},
							"end":          {Kind: &structpb.Value_NumberValue{NumberValue: float64(outputArr[i].End)}},
						},
					},
				}}
			}
			output := structpb.Struct{
				Fields: map[string]*structpb.Value{"results": {Kind: &structpb.Value_ListValue{ListValue: &classes}}},
			}
			outputs = append(outputs, &output)
		case translationTask:
			inputStruct := TranslationRequest{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}
			jsonBody, _ := json.Marshal(inputStruct)
			resp, err := client.MakeHFAPIRequest(jsonBody, model)
			if err != nil {
				return nil, err
			}
			outputArr := []TranslationResponse{}
			err = json.Unmarshal(resp, &outputArr)
			if err != nil {
				return nil, err
			}
			output := structpb.Struct{Fields: make(map[string]*structpb.Value)}
			output.Fields["translation_text"] = structpb.NewStringValue(outputArr[0].TranslationText)
			outputs = append(outputs, &output)
		case zeroShotClassificationTask:
			inputStruct := ZeroShotRequest{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}
			jsonBody, _ := json.Marshal(inputStruct)
			resp, err := client.MakeHFAPIRequest(jsonBody, model)
			if err != nil {
				return nil, err
			}
			output := structpb.Struct{}
			err = protojson.Unmarshal(resp, &output)
			if err != nil {
				return nil, err
			}
			outputs = append(outputs, &output)
		// TODO: fix this task
		// case featureExtractionTask:
		// 	inputStruct := FeatureExtractionRequest{}
		// 	err := base.ConvertFromStructpb(input, &inputStruct)
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// 	jsonBody, _ := json.Marshal(inputStruct)
		// 	resp, err := client.MakeHFAPIRequest(jsonBody, model)
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// 	threeDArr := [][][]float64{}
		// 	err = json.Unmarshal(resp, &threeDArr)
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// 	if len(threeDArr) <= 0 {
		// 		return nil, errors.New("invalid response")
		// 	}
		// 	nestedArr := threeDArr[0]
		// 	features := structpb.ListValue{}
		// 	features.Values = make([]*structpb.Value, len(nestedArr))
		// 	for i, innerArr := range nestedArr {
		// 		innerValues := make([]*structpb.Value, len(innerArr))
		// 		for j := range innerArr {
		// 			innerValues[j] = &structpb.Value{Kind: &structpb.Value_NumberValue{NumberValue: innerArr[j]}}
		// 		}
		// 		features.Values[i] = &structpb.Value{Kind: &structpb.Value_ListValue{ListValue: &structpb.ListValue{Values: innerValues}}}
		// 	}
		// 	output := structpb.Struct{
		// 		Fields: map[string]*structpb.Value{"feature": {Kind: &structpb.Value_ListValue{ListValue: &features}}},
		// 	}
		// 	outputs = append(outputs, &output)
		case questionAnsweringTask:
			inputStruct := QuestionAnsweringRequest{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}
			jsonBody, _ := json.Marshal(inputStruct)
			resp, err := client.MakeHFAPIRequest(jsonBody, model)
			if err != nil {
				return nil, err
			}
			output := structpb.Struct{}
			err = protojson.Unmarshal(resp, &output)
			if err != nil {
				return nil, err
			}
			outputs = append(outputs, &output)
		case tableQuestionAnsweringTask:
			inputStruct := TableQuestionAnsweringRequest{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}
			jsonBody, _ := json.Marshal(inputStruct)
			resp, err := client.MakeHFAPIRequest(jsonBody, model)
			if err != nil {
				return nil, err
			}
			output := structpb.Struct{}
			err = protojson.Unmarshal(resp, &output)
			if err != nil {
				return nil, err
			}
			outputs = append(outputs, &output)
		case sentenceSimilarityTask:
			inputStruct := SentenceSimilarityRequest{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}
			jsonBody, _ := json.Marshal(inputStruct)
			resp, err := client.MakeHFAPIRequest(jsonBody, model)
			if err != nil {
				return nil, err
			}
			outputArr := []float64{}
			err = json.Unmarshal(resp, &outputArr)
			if err != nil {
				return nil, err
			}
			scores := structpb.ListValue{}
			scores.Values = make([]*structpb.Value, len(outputArr))
			for i := range outputArr {
				scores.Values[i] = &structpb.Value{Kind: &structpb.Value_NumberValue{NumberValue: outputArr[i]}}
			}
			output := structpb.Struct{
				Fields: map[string]*structpb.Value{"scores": {Kind: &structpb.Value_ListValue{ListValue: &scores}}},
			}
			outputs = append(outputs, &output)
		case conversationalTask:
			inputStruct := ConversationalRequest{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}
			jsonBody, _ := json.Marshal(inputStruct)
			resp, err := client.MakeHFAPIRequest(jsonBody, model)
			if err != nil {
				return nil, err
			}
			output := structpb.Struct{}
			err = protojson.Unmarshal(resp, &output)
			if err != nil {
				return nil, err
			}
			outputs = append(outputs, &output)
		case imageClassificationTask:
			inputStruct := ImageRequest{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}
			b, err := base64.StdEncoding.DecodeString(inputStruct.Image)
			if err != nil {
				return nil, err
			}
			resp, err := client.MakeHFAPIRequest(b, model)
			if err != nil {
				return nil, err
			}
			outputArr := []ClassificationResponse{}
			err = json.Unmarshal(resp, &outputArr)
			if err != nil {
				return nil, err
			}
			classes := structpb.ListValue{}
			classes.Values = make([]*structpb.Value, len(outputArr))
			for i := range outputArr {
				classes.Values[i] = &structpb.Value{Kind: &structpb.Value_StructValue{
					StructValue: &structpb.Struct{Fields: map[string]*structpb.Value{
						"score": {Kind: &structpb.Value_NumberValue{NumberValue: outputArr[i].Score}},
						"label": {Kind: &structpb.Value_StringValue{StringValue: outputArr[i].Label}},
					}}},
				}
			}
			output := structpb.Struct{}
			output.Fields = map[string]*structpb.Value{
				"classes": {Kind: &structpb.Value_ListValue{ListValue: &classes}},
			}
			outputs = append(outputs, &output)
		case imageSegmentationTask:
			inputStruct := ImageRequest{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}
			b, err := base64.StdEncoding.DecodeString(inputStruct.Image)
			if err != nil {
				return nil, err
			}
			resp, err := client.MakeHFAPIRequest(b, model)
			if err != nil {
				return nil, err
			}
			outputArr := []ImageSegmentationResponse{}
			err = json.Unmarshal(resp, &outputArr)
			if err != nil {
				return nil, err
			}
			segments := structpb.ListValue{}
			segments.Values = make([]*structpb.Value, len(outputArr))
			for i := range outputArr {
				segments.Values[i] = &structpb.Value{Kind: &structpb.Value_StructValue{
					StructValue: &structpb.Struct{Fields: map[string]*structpb.Value{
						"score": {Kind: &structpb.Value_NumberValue{NumberValue: outputArr[i].Score}},
						"label": {Kind: &structpb.Value_StringValue{StringValue: outputArr[i].Label}},
						"mask":  {Kind: &structpb.Value_StringValue{StringValue: outputArr[i].Mask}},
					}}},
				}
			}
			output := structpb.Struct{}
			output.Fields = map[string]*structpb.Value{
				"segments": {Kind: &structpb.Value_ListValue{ListValue: &segments}},
			}
			outputs = append(outputs, &output)
		case objectDetectionTask:
			inputStruct := ImageRequest{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}
			b, err := base64.StdEncoding.DecodeString(inputStruct.Image)
			if err != nil {
				return nil, err
			}
			resp, err := client.MakeHFAPIRequest(b, model)
			if err != nil {
				return nil, err
			}
			outputArr := []ObjectDetectionResponse{}
			err = json.Unmarshal(resp, &outputArr)
			if err != nil {
				return nil, err
			}
			objects := structpb.ListValue{}
			objects.Values = make([]*structpb.Value, len(outputArr))
			for i := range outputArr {
				objects.Values[i] = &structpb.Value{Kind: &structpb.Value_StructValue{
					StructValue: &structpb.Struct{Fields: map[string]*structpb.Value{
						"score": {Kind: &structpb.Value_NumberValue{NumberValue: outputArr[i].Score}},
						"label": {Kind: &structpb.Value_StringValue{StringValue: outputArr[i].Label}},
						"box": {Kind: &structpb.Value_StructValue{StructValue: &structpb.Struct{
							Fields: map[string]*structpb.Value{
								"xmin": {Kind: &structpb.Value_NumberValue{NumberValue: float64(outputArr[i].Box.XMin)}},
								"ymin": {Kind: &structpb.Value_NumberValue{NumberValue: float64(outputArr[i].Box.YMin)}},
								"xmax": {Kind: &structpb.Value_NumberValue{NumberValue: float64(outputArr[i].Box.XMax)}},
								"ymax": {Kind: &structpb.Value_NumberValue{NumberValue: float64(outputArr[i].Box.YMax)}},
							},
						}}},
					}},
				}}
			}
			output := structpb.Struct{}
			output.Fields = map[string]*structpb.Value{
				"objects": {Kind: &structpb.Value_ListValue{ListValue: &objects}},
			}
			outputs = append(outputs, &output)
		case imageToTextTask:
			inputStruct := ImageRequest{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}
			b, err := base64.StdEncoding.DecodeString(inputStruct.Image)
			if err != nil {
				return nil, err
			}
			resp, err := client.MakeHFAPIRequest(b, model)
			if err != nil {
				return nil, err
			}
			outputArr := []ImageToTextResponse{}
			err = json.Unmarshal(resp, &outputArr)
			if err != nil {
				return nil, err
			}
			if len(outputArr) <= 0 {
				return nil, errors.New("invalid response")
			}
			output := structpb.Struct{
				Fields: map[string]*structpb.Value{
					"text": {Kind: &structpb.Value_StringValue{StringValue: outputArr[0].GeneratedText}},
				},
			}
			outputs = append(outputs, &output)
		case speechRecognitionTask:
			inputStruct := AudioRequest{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}
			b, err := base64.StdEncoding.DecodeString(inputStruct.Audio)
			if err != nil {
				return nil, err
			}
			resp, err := client.MakeHFAPIRequest(b, model)
			if err != nil {
				return nil, err
			}
			output := structpb.Struct{}
			err = protojson.Unmarshal(resp, &output)
			if err != nil {
				return nil, err
			}
			outputs = append(outputs, &output)
		case audioClassificationTask:
			inputStruct := AudioRequest{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}
			b, err := base64.StdEncoding.DecodeString(inputStruct.Audio)
			if err != nil {
				return nil, err
			}
			resp, err := client.MakeHFAPIRequest(b, model)
			if err != nil {
				return nil, err
			}
			outputArr := []ClassificationResponse{}
			err = json.Unmarshal(resp, &outputArr)
			if err != nil {
				return nil, err
			}
			classes := structpb.ListValue{}
			classes.Values = make([]*structpb.Value, len(outputArr))
			for i := range outputArr {
				classes.Values[i] = &structpb.Value{Kind: &structpb.Value_StructValue{
					StructValue: &structpb.Struct{Fields: map[string]*structpb.Value{
						"score": {Kind: &structpb.Value_NumberValue{NumberValue: outputArr[i].Score}},
						"label": {Kind: &structpb.Value_StringValue{StringValue: outputArr[i].Label}},
					}}},
				}
			}
			output := structpb.Struct{}
			output.Fields = map[string]*structpb.Value{
				"classes": {Kind: &structpb.Value_ListValue{ListValue: &classes}},
			}
			outputs = append(outputs, &output)
		default:
			return nil, fmt.Errorf("not supported task: %s", e.Task)
		}
	}

	return outputs, nil
}

func (c *Connector) Test(defUid uuid.UUID, config *structpb.Struct, logger *zap.Logger) (connectorPB.ConnectorResource_State, error) {
	client := NewClient(getAPIKey(config), getBaseURL(config), isCustomEndpoint(config))
	return client.GetConnectionState()
}
