package instill

import (
	"context"
	"fmt"

	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	modelPB "github.com/instill-ai/protogen-go/model/model/v1alpha"
)

type LLMInput struct {

	// The prompt text
	Prompt string
	// The prompt images
	PromptImages []*modelPB.PromptImage
	// The chat history
	ChatHistory []*modelPB.Message
	// The system message
	SystemMessage *string
	// The maximum number of tokens for model to generate
	MaxNewTokens *int32
	// The temperature for sampling
	Temperature *float32
	// Top k for sampling
	TopK *int32
	// The seed
	Seed *int32
	// The extra parameters
	ExtraParams *structpb.Struct
}

func ExtractChatHistory(input *structpb.Struct) []*modelPB.Message {

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
	return history
}

func (c *Execution) SendLLMRequest(grpcClient modelPB.ModelPublicServiceClient, req *modelPB.TriggerUserModelRequest, modelName string) (*structpb.Struct, error) {

	md := metadata.Pairs("Authorization", fmt.Sprintf("Bearer %s", getAPIKey(c.Config)), "Instill-User-Uid", getInstillUserUid(c.Config))
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	res, err := grpcClient.TriggerUserModel(ctx, req)
	if err != nil || res == nil {
		return nil, err
	}
	taskOutputs := res.GetTaskOutputs()
	if len(taskOutputs) <= 0 {
		return nil, fmt.Errorf("invalid output: %v for model: %s", taskOutputs, modelName)
	}

	textGenChatOutput := taskOutputs[0].GetTextGenerationChat()
	if textGenChatOutput == nil {
		return nil, fmt.Errorf("invalid output: %v for model: %s", textGenChatOutput, modelName)
	}
	outputJson, err := protojson.MarshalOptions{
		UseProtoNames:   true,
		EmitUnpopulated: true,
	}.Marshal(textGenChatOutput)
	if err != nil {
		return nil, err
	}
	output := &structpb.Struct{}
	err = protojson.Unmarshal(outputJson, output)
	if err != nil {
		return nil, err
	}
	return output, nil
}
