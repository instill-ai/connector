package instill

import (
	"context"
	"fmt"

	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
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

		conversation := []*modelPB.ConversationObject{}
		for _, item := range input.GetFields()["conversation"].GetListValue().AsSlice() {
			conversation = append(conversation, &modelPB.ConversationObject{
				Role:    item.(map[string]interface{})["role"].(string),
				Content: item.(map[string]interface{})["content"].(string),
			})
		}
		textGenerationChatInput := &modelPB.TextGenerationChatInput{
			Conversation: conversation,
		}
		if _, ok := input.GetFields()["max_new_tokens"]; ok {
			v := int32(input.GetFields()["max_new_tokens"].GetNumberValue())
			textGenerationChatInput.MaxNewTokens = &v
		}
		if _, ok := input.GetFields()["temperature"]; ok {
			v := float32(input.GetFields()["temperature"].GetNumberValue())
			textGenerationChatInput.Temperature = &v
		}
		if _, ok := input.GetFields()["top_k"]; ok {
			v := int32(input.GetFields()["top_k"].GetNumberValue())
			textGenerationChatInput.TopK = &v
		}
		if _, ok := input.GetFields()["seed"]; ok {
			v := int32(input.GetFields()["seed"].GetNumberValue())
			textGenerationChatInput.Seed = &v
		}

		extraParams := []*modelPB.ExtraParamObject{}
		if _, ok := input.GetFields()["extra_params"]; ok {
			for _, item := range input.GetFields()["extra_params"].GetListValue().AsSlice() {
				extraParams = append(extraParams, &modelPB.ExtraParamObject{
					ParamName:  item.(map[string]interface{})["param_name"].(string),
					ParamValue: item.(map[string]interface{})["param_value"].(string),
				})
			}
			textGenerationChatInput.ExtraParams = extraParams
		}

		taskInput := &modelPB.TaskInput_TextGenerationChat{
			TextGenerationChat: textGenerationChatInput,
		}

		// only support batch 1
		req := modelPB.TriggerUserModelRequest{
			Name:       modelName,
			TaskInputs: []*modelPB.TaskInput{{Input: taskInput}},
		}
		md := metadata.Pairs("Authorization", fmt.Sprintf("Bearer %s", getAPIKey(c.Config)), "Instill-User-Uid", getInstillUserUid(c.Config))
		ctx := metadata.NewOutgoingContext(context.Background(), md)
		res, err := grpcClient.TriggerUserModel(ctx, &req)
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
		outputs = append(outputs, output)

	}
	return outputs, nil
}
