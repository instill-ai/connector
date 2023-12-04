package instill

import (
	"context"
	"fmt"

	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	modelPB "github.com/instill-ai/protogen-go/model/model/v1alpha"
)

func (c *Execution) executeTextGeneration(grpcClient modelPB.ModelPublicServiceClient, modelName string, inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	if len(inputs) <= 0 {
		return nil, fmt.Errorf("invalid input: %v for model: %s", inputs, modelName)
	}

	outputs := []*structpb.Struct{}

	for _, input := range inputs {

		textGenerationInput := &modelPB.TextGenerationInput{
			Prompt: input.GetFields()["prompt"].GetStringValue(),
		}
		if _, ok := input.GetFields()["max_new_tokens"]; ok {
			v := int32(input.GetFields()["max_new_tokens"].GetNumberValue())
			textGenerationInput.MaxNewTokens = &v
		}
		if _, ok := input.GetFields()["temperature"]; ok {
			v := float32(input.GetFields()["temperature"].GetNumberValue())
			textGenerationInput.Temperature = &v
		}
		if _, ok := input.GetFields()["top_k"]; ok {
			v := int32(input.GetFields()["top_k"].GetNumberValue())
			textGenerationInput.TopK = &v
		}
		if _, ok := input.GetFields()["seed"]; ok {
			v := int32(input.GetFields()["seed"].GetNumberValue())
			textGenerationInput.Seed = &v
		}

		taskInput := &modelPB.TaskInput_TextGeneration{
			TextGeneration: textGenerationInput,
		}

		// only support batch 1
		req := modelPB.TriggerUserModelRequest{
			Name:       modelName,
			TaskInputs: []*modelPB.TaskInput{{Input: taskInput}},
		}
		if c.client == nil || grpcClient == nil {
			return nil, fmt.Errorf("client not setup: %v", c.client)
		}
		md := metadata.Pairs("Authorization", fmt.Sprintf("Bearer %s", getAPIKey(c.Config)))
		ctx := metadata.NewOutgoingContext(context.Background(), md)
		res, err := grpcClient.TriggerUserModel(ctx, &req)
		if err != nil || res == nil {
			return nil, err
		}
		taskOutputs := res.GetTaskOutputs()
		if len(taskOutputs) <= 0 {
			return nil, fmt.Errorf("invalid output: %v for model: %s", taskOutputs, modelName)
		}

		textGenOutput := taskOutputs[0].GetTextGeneration()
		if textGenOutput == nil {
			return nil, fmt.Errorf("invalid output: %v for model: %s", textGenOutput, modelName)
		}
		outputJson, err := protojson.MarshalOptions{
			UseProtoNames:   true,
			EmitUnpopulated: true,
		}.Marshal(textGenOutput)
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
