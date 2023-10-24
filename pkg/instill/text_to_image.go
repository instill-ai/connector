package instill

import (
	"context"
	"fmt"

	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	modelPB "github.com/instill-ai/protogen-go/model/model/v1alpha"
)

func (c *Execution) executeTextToImage(grpcClient modelPB.ModelPublicServiceClient, modelName string, inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	if len(inputs) <= 0 {
		return nil, fmt.Errorf("invalid input: %v for model: %s", inputs, modelName)
	}

	outputs := []*structpb.Struct{}
	for _, input := range inputs {
		inputJson, err := protojson.Marshal(input)
		if err != nil {
			return nil, err
		}
		textToImageInput := &modelPB.TextToImageInput{}
		err = protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(inputJson, textToImageInput)
		if err != nil {
			return nil, err
		}

		taskInput := &modelPB.TaskInput_TextToImage{
			TextToImage: textToImageInput,
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

		textToImgOutput := taskOutputs[0].GetTextToImage()
		if textToImgOutput == nil || len(textToImgOutput.Images) <= 0 {
			return nil, fmt.Errorf("invalid output: %v for model: %s", textToImgOutput, modelName)
		}

		outputJson, err := protojson.MarshalOptions{
			UseProtoNames:   true,
			EmitUnpopulated: true,
		}.Marshal(textToImgOutput)
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
