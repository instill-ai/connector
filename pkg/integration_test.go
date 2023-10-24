//go:build integration
// +build integration

package pkg

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"
	"github.com/instill-ai/connector/pkg/huggingface"
	"github.com/instill-ai/connector/pkg/openai"
	"github.com/instill-ai/connector/pkg/stabilityai"
)

var (
	stabilityAIKey = "<valid api key>"
	openAIKey      = "<valid api key>"
	huggingFaceKey = "<valid api key>"
	hfCon          base.IExecution
	hfConfig       *structpb.Struct
	jsonKey        []byte
	gcsCon         base.IExecution
)

func init() {

	b, _ := ioutil.ReadFile("test_artifacts/open_ai.txt")
	openAIKey = string(b)
	b, _ = ioutil.ReadFile("test_artifacts/stability_ai.txt")
	stabilityAIKey = string(b)
	b, _ = ioutil.ReadFile("test_artifacts/hugging_face.txt")
	huggingFaceKey = string(b)
	hfConfig = &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"api_key":            {Kind: &structpb.Value_StringValue{StringValue: huggingFaceKey}},
			"base_url":           {Kind: &structpb.Value_StringValue{StringValue: "https://api-inference.huggingface.co"}},
			"is_custom_endpoint": {Kind: &structpb.Value_BoolValue{BoolValue: false}},
		}}
	c := Init(nil)
	hfCon, _ = c.CreateExecution(c.ListDefinitionUids()[3], hfConfig, nil)

	jsonKey, _ = ioutil.ReadFile("test_artifacts/gcp_key.json")
	gcsConfig := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"json_key":    {Kind: &structpb.Value_StringValue{StringValue: string(jsonKey)}},
			"bucket_name": {Kind: &structpb.Value_StringValue{StringValue: "gcs-connector-sjne"}},
		}}
	logger, _ = zap.NewDevelopment()
	c := Init(logger, ConnectorOptions{})
	uuid := c.ListDefinitionUids()
	gcsCon, _ = c.CreateExecution(uuid[len(uuid)-1], gcsConfig, nil)

}

func TestStabilityAITextToImage(t *testing.T) {
	config := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"api_key": {Kind: &structpb.Value_StringValue{StringValue: stabilityAIKey}},
		},
	}
	inputStruct := stabilityai.TextToImageInput{
		Task:    "TASK_TEXT_TO_IMAGE",
		Prompts: []string{"black", "dog"},
		Engine:  "stable-diffusion-v1-5",
	}
	in, err := base.ConvertToStructpb(inputStruct)
	fmt.Printf("err:%s", err)
	c := Init(nil)
	con, err := c.CreateExecution(c.ListDefinitionUids()[0], config, nil)
	fmt.Printf("err:%s", err)
	op, err := con.Execute([]*structpb.Struct{in})
	fmt.Printf("\n op :%v, err:%s", op, err)
}

func Test_ListEngines(t *testing.T) {
	client := stabilityai.NewClient(stabilityAIKey)
	engines, err := client.ListEngines()
	fmt.Printf("engines: %v, err: %v", engines, err)
}

func TestStabilityAIImageToImage(t *testing.T) {
	config := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"api_key": {Kind: &structpb.Value_StringValue{StringValue: stabilityAIKey}},
		},
	}
	b, _ := ioutil.ReadFile("test_artifacts/image.jpg")
	inputStruct := stabilityai.ImageToImageInput{
		Task:      "TASK_IMAGE_TO_IMAGE",
		Engine:    "stable-diffusion-v1",
		Prompts:   []string{"invert colors"},
		InitImage: base64.StdEncoding.EncodeToString(b),
	}
	in, err := base.ConvertToStructpb(inputStruct)
	fmt.Printf("err:%s", err)
	c := Init(nil)
	con, err := c.CreateExecution(c.ListDefinitionUids()[0], config, nil)
	fmt.Printf("\n err: %s", err)
	op, err := con.Execute([]*structpb.Struct{in})
	fmt.Printf("\n op: %v, err: %s", op, err)
	l := op[0].Fields["images"].GetListValue()
	b, _ = stabilityai.DecodeBase64(l.Values[0].GetStringValue())
	err = ioutil.WriteFile("test_artifacts/image_op.png", b, 0644)
}

func TestOpenAITextGeneration(t *testing.T) {
	config := &structpb.Struct{Fields: map[string]*structpb.Value{"api_key": {Kind: &structpb.Value_StringValue{StringValue: openAIKey}}}}
	inputStruct := openai.TextCompletionInput{Prompt: "how are you doing?", Model: "gpt-3.5-turbo"}
	in, err := base.ConvertToStructpb(inputStruct)
	in.Fields["task"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "TASK_TEXT_GENERATION"}}
	fmt.Printf("err:%s", err)
	c := Init(nil)
	con, err := c.CreateExecution(c.ListDefinitionUids()[2], config, nil)
	fmt.Printf("err:%s", err)
	op, err := con.Execute([]*structpb.Struct{in})
	fmt.Printf("\n op :%v, err:%s", op, err)
}

func Test_ListModels(t *testing.T) {
	c := openai.Client{
		APIKey:     openAIKey,
		HTTPClient: &http.Client{},
	}
	res, err := c.ListModels()
	fmt.Printf("res: %v, err: %v", res, err)
}

func TestOpenAIAudioTranscription(t *testing.T) {
	config := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"api_key": {Kind: &structpb.Value_StringValue{StringValue: openAIKey}},
		},
	}
	b, _ := ioutil.ReadFile("test_artifacts/recording.m4a")
	inputStruct := openai.AudioTranscriptionInput{
		Audio: base64.StdEncoding.EncodeToString(b),
		Model: "whisper-1",
	}
	in, err := base.ConvertToStructpb(inputStruct)
	in.Fields["task"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "TASK_SPEECH_RECOGNITION"}}
	fmt.Printf("err:%s", err)
	c := Init(nil)
	con, err := c.CreateExecution(c.ListDefinitionUids()[2], config, nil)
	fmt.Printf("err:%s", err)
	op, err := con.Execute([]*structpb.Struct{in})
	fmt.Printf("\n op :%v, err:%s", op, err)
}

func TestGetConnectionState(t *testing.T) {
	c := Init(nil)
	state, err := c.Test(c.ListDefinitionUids()[3], hfConfig, nil)
	fmt.Printf("\n state: %v, err: %v", state, err)
}

func TestHuggingFaceTextToImage(t *testing.T) {
	req := huggingface.TextToImageRequest{Inputs: "a black dog"}
	var in structpb.Struct
	b, _ := json.Marshal(req)
	protojson.Unmarshal(b, &in)
	in.Fields["task"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "TEXT_TO_IMAGE"}}
	in.Fields["model"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "runwayml/stable-diffusion-v1-5"}}
	op, err := hfCon.Execute([]*structpb.Struct{&in})
	fmt.Printf("\n op :%v, err:%s", op, err)
}

func TestHuggingFaceFillMask(t *testing.T) {
	req := huggingface.FillMaskRequest{Inputs: "The answer to the universe is [MASK]."}
	var in structpb.Struct
	b, _ := json.Marshal(req)
	protojson.Unmarshal(b, &in)
	in.Fields["task"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "FILL_MASK"}}
	in.Fields["model"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "bert-base-uncased"}}
	op, err := hfCon.Execute([]*structpb.Struct{&in})
	fmt.Printf("\n op :%v, err:%s", op, err)
}

func TestHuggingFaceSummarization(t *testing.T) {
	req := huggingface.SummarizationRequest{Inputs: "The tower is 324 metres (1,063 ft) tall, about the same height as an 81-storey building, and the tallest structure in Paris. Its base is square, measuring 125 metres (410 ft) on each side. During its construction, the Eiffel Tower surpassed the Washington Monument to become the tallest man-made structure in the world, a title it held for 41 years until the Chrysler Building in New York City was finished in 1930. It was the first structure to reach a height of 300 metres. Due to the addition of a broadcasting aerial at the top of the tower in 1957, it is now taller than the Chrysler Building by 5.2 metres (17 ft). Excluding transmitters, the Eiffel Tower is the second tallest free-standing structure in France after the Millau Viaduct."}
	var in structpb.Struct
	b, _ := json.Marshal(req)
	protojson.Unmarshal(b, &in)
	in.Fields["task"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "SUMMARIZATION"}}
	in.Fields["model"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "facebook/bart-large-cnn"}}
	op, err := hfCon.Execute([]*structpb.Struct{&in})
	fmt.Printf("\n op :%v, err:%s", op, err)
}

func TestHuggingFaceTextClassification(t *testing.T) {
	req := huggingface.TextClassificationRequest{Inputs: "I like you. I love you"}
	var in structpb.Struct
	b, _ := json.Marshal(req)
	protojson.Unmarshal(b, &in)
	in.Fields["task"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "TEXT_CLASSIFICATION"}}
	in.Fields["model"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "distilbert-base-uncased-finetuned-sst-2-english"}}
	op, err := hfCon.Execute([]*structpb.Struct{&in})
	fmt.Printf("\n op :%v, err:%s", op, err)
}

func TestHuggingFaceTextGeneration(t *testing.T) {
	req := huggingface.TextClassificationRequest{Inputs: "The answer to the universe is"}
	var in structpb.Struct
	b, _ := json.Marshal(req)
	protojson.Unmarshal(b, &in)
	in.Fields["task"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "TEXT_GENERATION"}}
	in.Fields["model"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "gpt2"}}
	op, err := hfCon.Execute([]*structpb.Struct{&in})
	fmt.Printf("\n op :%v, err:%s", op, err)
}

func TestHuggingFaceTokenClassification(t *testing.T) {
	req := huggingface.TextClassificationRequest{Inputs: "My name is Sarah Jessica Parker but you can call me Jessica"}
	var in structpb.Struct
	b, _ := json.Marshal(req)
	protojson.Unmarshal(b, &in)
	in.Fields["task"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "TOKEN_CLASSIFICATION"}}
	in.Fields["model"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "dbmdz/bert-large-cased-finetuned-conll03-english"}}
	op, err := hfCon.Execute([]*structpb.Struct{&in})
	fmt.Printf("\n op :%v, err:%s", op, err)
}

func TestHuggingFaceTranslation(t *testing.T) {
	req := huggingface.TranslationRequest{Inputs: "Меня зовут Вольфганг и я живу в Берлине"}
	var in structpb.Struct
	b, _ := json.Marshal(req)
	protojson.Unmarshal(b, &in)
	in.Fields["task"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "TRANSLATION"}}
	in.Fields["model"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "Helsinki-NLP/opus-mt-ru-en"}}
	op, err := hfCon.Execute([]*structpb.Struct{&in})
	fmt.Printf("\n op :%v, err:%s", op, err)
}

func TestHuggingFaceZeroShotClassification(t *testing.T) {
	req := huggingface.ZeroShotRequest{
		Inputs:     "Hi, I recently bought a device from your company but it is not working as advertised and I would like to get reimbursed!",
		Parameters: huggingface.ZeroShotParameters{CandidateLabels: []string{"refund", "legal", "faq"}},
	}
	var in structpb.Struct
	b, _ := json.Marshal(req)
	protojson.Unmarshal(b, &in)
	in.Fields["task"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "ZERO_SHOT_CLASSIFICATION"}}
	in.Fields["model"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "facebook/bart-large-mnli"}}
	op, err := hfCon.Execute([]*structpb.Struct{&in})
	fmt.Printf("\n op :%v, err:%s", op, err)
}

func TestHuggingFaceFeatureExtraction(t *testing.T) {
	req := huggingface.FeatureExtractionRequest{Inputs: "I love programming"}
	var in structpb.Struct
	b, _ := json.Marshal(req)
	protojson.Unmarshal(b, &in)
	in.Fields["task"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "FEATURE_EXTRACTION"}}
	in.Fields["model"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "facebook/bart-large"}}
	op, err := hfCon.Execute([]*structpb.Struct{&in})
	fmt.Printf("\n op :%v, err:%s", op, err)
}

func TestHuggingFaceQuestionAnswering(t *testing.T) {
	req := huggingface.QuestionAnsweringRequest{
		Inputs: huggingface.QuestionAnsweringInputs{
			Question: "What is my name?",
			Context:  "My name is Clara and I live in Berkeley.",
		}}
	var in structpb.Struct
	b, _ := json.Marshal(req)
	protojson.Unmarshal(b, &in)
	in.Fields["task"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "QUESTION_ANSWERING"}}
	in.Fields["model"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "deepset/roberta-base-squad2"}}
	op, err := hfCon.Execute([]*structpb.Struct{&in})
	fmt.Printf("\n op :%v, err:%s", op, err)
}

func TestHuggingFaceTableQuestionAnswering(t *testing.T) {
	req := huggingface.TableQuestionAnsweringRequest{
		Inputs: huggingface.TableQuestionAnsweringInputs{
			Query: "How many stars does the transformers repository have?",
			Table: map[string][]string{
				"Repository":           {"Transformers", "Datasets", "Tokenizers"},
				"Stars":                {"36542", "4512", "3934"},
				"Contributors":         {"651", "77", "34"},
				"Programming language": {"Python", "Python", "Rust, Python and NodeJS"},
			},
		}}
	var in structpb.Struct
	b, _ := json.Marshal(req)
	protojson.Unmarshal(b, &in)
	in.Fields["task"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "TABLE_QUESTION_ANSWERING"}}
	in.Fields["model"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "google/tapas-base-finetuned-wtq"}}
	op, err := hfCon.Execute([]*structpb.Struct{&in})
	fmt.Printf("\n op :%v, err:%s", op, err)
}

func TestHuggingFaceSentenceSimilarity(t *testing.T) {
	req := huggingface.SentenceSimilarityRequest{
		Inputs: huggingface.SentenceSimilarityInputs{
			SourceSentence: "That is a happy person",
			Sentences:      []string{"That is a happy dog", "That is a very happy person", "Today is a sunny day"},
		},
	}
	var in structpb.Struct
	b, _ := json.Marshal(req)
	protojson.Unmarshal(b, &in)
	in.Fields["task"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "SENTENCE_SIMILARITY"}}
	in.Fields["model"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "sentence-transformers/all-MiniLM-L6-v2"}}
	op, err := hfCon.Execute([]*structpb.Struct{&in})
	fmt.Printf("\n op :%v, err:%s", op, err)
}

func TestHuggingFaceConversational(t *testing.T) {
	req := huggingface.ConversationalRequest{
		Inputs: huggingface.ConverstationalInputs{
			Text:               "Can you explain why ?",
			GeneratedResponses: []string{"It is Die Hard for sure."},
			PastUserInputs:     []string{"Which movie is the best ?"},
		},
	}
	var in structpb.Struct
	b, _ := json.Marshal(req)
	protojson.Unmarshal(b, &in)
	in.Fields["task"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "CONVERSATIONAL"}}
	in.Fields["model"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "microsoft/DialoGPT-large"}}
	op, err := hfCon.Execute([]*structpb.Struct{&in})
	fmt.Printf("\n op :%v, err:%s", op, err)
}

func TestHuggingFaceImageClassification(t *testing.T) {
	b, _ := ioutil.ReadFile("test_artifacts/image.jpg")
	req := huggingface.ImageRequest{Image: base64.StdEncoding.EncodeToString(b)}
	var in structpb.Struct
	b, _ = json.Marshal(req)
	protojson.Unmarshal(b, &in)
	in.Fields["task"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "IMAGE_CLASSIFICATION"}}
	in.Fields["model"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "google/vit-base-patch16-224"}}
	op, err := hfCon.Execute([]*structpb.Struct{&in})
	fmt.Printf("\n op :%v, err:%s", op, err)
}

func TestHuggingFaceImageSegmentation(t *testing.T) {
	b, _ := ioutil.ReadFile("test_artifacts/image.jpg")
	req := huggingface.ImageRequest{Image: base64.StdEncoding.EncodeToString(b)}
	var in structpb.Struct
	b, _ = json.Marshal(req)
	protojson.Unmarshal(b, &in)
	in.Fields["task"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "IMAGE_SEGMENTATION"}}
	in.Fields["model"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "facebook/detr-resnet-50-panoptic"}}
	op, err := hfCon.Execute([]*structpb.Struct{&in})
	fmt.Printf("\n op :%v, err:%s", op, err)
}

func TestHuggingFaceObjectDetection(t *testing.T) {
	b, _ := ioutil.ReadFile("test_artifacts/image.jpg")
	req := huggingface.ImageRequest{Image: base64.StdEncoding.EncodeToString(b)}
	var in structpb.Struct
	b, _ = json.Marshal(req)
	protojson.Unmarshal(b, &in)
	in.Fields["task"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "OBJECT_DETECTION"}}
	in.Fields["model"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "facebook/detr-resnet-50"}}
	op, err := hfCon.Execute([]*structpb.Struct{&in})
	fmt.Printf("\n op :%v, err:%s", op, err)
}

func TestHuggingFaceImageToText(t *testing.T) {
	b, _ := ioutil.ReadFile("test_artifacts/image.jpg")
	req := huggingface.ImageRequest{Image: base64.StdEncoding.EncodeToString(b)}
	var in structpb.Struct
	b, _ = json.Marshal(req)
	protojson.Unmarshal(b, &in)
	in.Fields["task"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "IMAGE_TO_TEXT"}}
	in.Fields["model"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "nlpconnect/vit-gpt2-image-captioning"}}
	op, err := hfCon.Execute([]*structpb.Struct{&in})
	fmt.Printf("\n op :%v, err:%s", op, err)
}

func TestHuggingFaceSpeechRecognition(t *testing.T) {
	b, _ := ioutil.ReadFile("test_artifacts/recording.m4a")
	req := huggingface.AudioRequest{Audio: base64.StdEncoding.EncodeToString(b)}
	var in structpb.Struct
	b, _ = json.Marshal(req)
	protojson.Unmarshal(b, &in)
	in.Fields["task"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "SPEECH_RECOGNITION"}}
	in.Fields["model"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "facebook/wav2vec2-base-960h"}}
	op, err := hfCon.Execute([]*structpb.Struct{&in})
	fmt.Printf("\n op :%v, err:%s", op, err)
}

func TestHuggingFaceAudioClassification(t *testing.T) {
	b, _ := ioutil.ReadFile("test_artifacts/recording.m4a")
	req := huggingface.AudioRequest{Audio: base64.StdEncoding.EncodeToString(b)}
	var in structpb.Struct
	b, _ = json.Marshal(req)
	protojson.Unmarshal(b, &in)
	in.Fields["task"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "AUDIO_CLASSIFICATION"}}
	in.Fields["model"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "superb/hubert-large-superb-er"}}
	op, err := hfCon.Execute([]*structpb.Struct{&in})
	fmt.Printf("\n op :%v, err:%s", op, err)
}

func TestHuggingFaceCustomEndpointImageClassification(t *testing.T) {
	hfCustomConfig := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"api_key":            {Kind: &structpb.Value_StringValue{StringValue: huggingFaceKey}},
			"base_url":           {Kind: &structpb.Value_StringValue{StringValue: "valid URL here"}},
			"is_custom_endpoint": {Kind: &structpb.Value_BoolValue{BoolValue: true}},
		}}
	c := Init(nil)
	hfCustomCon, _ := c.CreateExecution(c.ListDefinitionUids()[3], hfCustomConfig, nil)

	b, _ := ioutil.ReadFile("test_artifacts/image.jpg")
	req := huggingface.ImageRequest{Image: base64.StdEncoding.EncodeToString(b)}
	var in structpb.Struct
	b, _ = json.Marshal(req)
	protojson.Unmarshal(b, &in)
	in.Fields["task"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "IMAGE_CLASSIFICATION"}}
	in.Fields["model"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: ""}}
	op, err := hfCustomCon.Execute([]*structpb.Struct{&in})
	fmt.Printf("\n op :%v, err:%s", op, err)
}

func TestGCSUpload(*testing.T) {
	objectName := "test_dog_img.jpg"
	fileContents, _ := ioutil.ReadFile("test_artifacts/dog.jpg")
	base64Str := base64.StdEncoding.EncodeToString(fileContents)
	input := []*structpb.Struct{{
		Fields: map[string]*structpb.Value{
			"task":        {Kind: &structpb.Value_StringValue{StringValue: "TASK_UPLOAD"}},
			"object_name": {Kind: &structpb.Value_StringValue{StringValue: objectName}},
			"data":        {Kind: &structpb.Value_StringValue{StringValue: base64Str}},
		}}}
	op, err := gcsCon.Execute(input)
	fmt.Printf("op: %v, err: %v", op, err)
}

func TestBigQueryInsert(*testing.T) {
	bigQueryConfig := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"json_key":   {Kind: &structpb.Value_StringValue{StringValue: string(jsonKey)}},
			"project_id": {Kind: &structpb.Value_StringValue{StringValue: "prj-c-connector-879a"}},
			"dataset_id": {Kind: &structpb.Value_StringValue{StringValue: "test_data_set"}},
			"table_name": {Kind: &structpb.Value_StringValue{StringValue: "test_table"}},
		}}

	logger, _ = zap.NewDevelopment()
	c := Init(logger, ConnectorOptions{})
	uuids := c.ListDefinitionUids()
	uuid := uuids[len(uuids)-2]
	bigQueryCon, _ := c.CreateExecution(uuid, bigQueryConfig, nil)
	state, err := c.Test(uuid, bigQueryConfig, nil)
	fmt.Printf("state: %v, err: %v", state, err)
	input := []*structpb.Struct{{
		Fields: map[string]*structpb.Value{
			"task": {Kind: &structpb.Value_StringValue{StringValue: "TASK_INSERT"}},
			"input": {Kind: &structpb.Value_StructValue{
				StructValue: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"id":   {Kind: &structpb.Value_NumberValue{NumberValue: 5}},
						"name": {Kind: &structpb.Value_StringValue{StringValue: "Tobias"}},
					},
				},
			}},
		}}}
	op, err := bigQueryCon.Execute(input)
	fmt.Printf("op: %v, err: %v", op, err)
}
