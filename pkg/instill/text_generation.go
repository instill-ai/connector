package instill

import (
	"fmt"

	"google.golang.org/protobuf/types/known/structpb"

	modelPB "github.com/instill-ai/protogen-go/model/model/v1alpha"
)

func ConvertLLMInput(input *structpb.Struct) *LLMInput {
	llmInput := &LLMInput{
		Prompt: input.GetFields()["prompt"].GetStringValue(),
	}

	if _, ok := input.GetFields()["system_message"]; ok {
		v := input.GetFields()["system_message"].GetStringValue()
		llmInput.SystemMessage = &v
	}

	if _, ok := input.GetFields()["prompt_images"]; ok {
		promptImages := []*modelPB.PromptImage{}
		for _, item := range input.GetFields()["prompt_images"].GetListValue().GetValues() {
			image := &modelPB.PromptImage{}
			image.Type = &modelPB.PromptImage_PromptImageBase64{
				PromptImageBase64: item.GetStringValue(),
			}
			promptImages = append(promptImages, image)
		}
		llmInput.PromptImages = promptImages
	}

	if _, ok := input.GetFields()["chat_history"]; ok {
		history := []*modelPB.Message{}
		for _, item := range input.GetFields()["chat_history"].GetListValue().GetValues() {
			contents := []*modelPB.MessageContent{}
			for _, contentItem := range item.GetStructValue().Fields["content"].GetListValue().GetValues() {
				t := contentItem.GetStructValue().Fields["type"].GetStringValue()
				content := &modelPB.MessageContent{
					Type: t,
				}
				if t == "text" {
					content.Content = &modelPB.MessageContent_Text{
						Text: contentItem.GetStructValue().Fields["text"].GetStringValue(),
					}
				} else {
					image := &modelPB.PromptImage{}
					image.Type = &modelPB.PromptImage_PromptImageBase64{
						PromptImageBase64: contentItem.GetStructValue().Fields["image_url"].GetStructValue().Fields["url"].GetStringValue(),
					}
					content.Content = &modelPB.MessageContent_ImageUrl{
						ImageUrl: &modelPB.ImageContent{
							ImageUrl: image,
						},
					}
				}
				contents = append(contents, content)
			}
			history = append(history, &modelPB.Message{
				Role:    item.GetStructValue().Fields["role"].GetStringValue(),
				Content: contents,
			})

		}
		llmInput.ChatHistory = history
	}

	if _, ok := input.GetFields()["max_new_tokens"]; ok {
		v := int32(input.GetFields()["max_new_tokens"].GetNumberValue())
		llmInput.MaxNewTokens = &v
	}
	if _, ok := input.GetFields()["temperature"]; ok {
		v := float32(input.GetFields()["temperature"].GetNumberValue())
		llmInput.Temperature = &v
	}
	if _, ok := input.GetFields()["top_k"]; ok {
		v := int32(input.GetFields()["top_k"].GetNumberValue())
		llmInput.TopK = &v
	}
	if _, ok := input.GetFields()["seed"]; ok {
		v := int32(input.GetFields()["seed"].GetNumberValue())
		llmInput.Seed = &v
	}
	if _, ok := input.GetFields()["extra_params"]; ok {
		v := input.GetFields()["extra_params"].GetStructValue()
		llmInput.ExtraParams = v
	}
	return llmInput

}

func (c *Execution) executeTextGeneration(grpcClient modelPB.ModelPublicServiceClient, modelName string, inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	if len(inputs) <= 0 {
		return nil, fmt.Errorf("invalid input: %v for model: %s", inputs, modelName)
	}

	if grpcClient == nil {
		return nil, fmt.Errorf("uninitialized client")
	}

	outputs := []*structpb.Struct{}

	for _, input := range inputs {

		llmInput := ConvertLLMInput(input)
		taskInput := &modelPB.TaskInput_TextGeneration{
			TextGeneration: &modelPB.TextGenerationInput{
				Prompt:        llmInput.Prompt,
				PromptImages:  llmInput.PromptImages,
				ChatHistory:   llmInput.ChatHistory,
				SystemMessage: llmInput.SystemMessage,
				MaxNewTokens:  llmInput.MaxNewTokens,
				Temperature:   llmInput.Temperature,
				TopK:          llmInput.TopK,
				Seed:          llmInput.Seed,
				ExtraParams:   llmInput.ExtraParams,
			},
		}

		// only support batch 1
		output, err := c.SendLLMRequest(grpcClient, &modelPB.TriggerUserModelRequest{
			Name:       modelName,
			TaskInputs: []*modelPB.TaskInput{{Input: taskInput}},
		}, modelName)
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, output)

	}
	return outputs, nil
}
