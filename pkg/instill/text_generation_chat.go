package instill

import (
	"fmt"

	"google.golang.org/protobuf/types/known/structpb"

	modelPB "github.com/instill-ai/protogen-go/model/model/v1alpha"
)

func (c *Execution) executeTextGenerationChat(grpcClient modelPB.ModelPublicServiceClient, modelName string, inputs []*structpb.Struct) ([]*structpb.Struct, error) {

	if len(inputs) <= 0 {
		return nil, fmt.Errorf("invalid input: %v for model: %s", inputs, modelName)
	}

	if grpcClient == nil {
		return nil, fmt.Errorf("uninitialized client")
	}

	outputs := []*structpb.Struct{}

	for _, input := range inputs {
		llmInput := ConvertLLMInput(input)
		taskInput := &modelPB.TaskInput_TextGenerationChat{
			TextGenerationChat: &modelPB.TextGenerationChatInput{
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
