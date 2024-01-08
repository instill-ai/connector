package instill

import (
	"context"
	"fmt"

	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"
	modelPB "github.com/instill-ai/protogen-go/model/model/v1alpha"
)

func (c *Execution) executeImageToImage(grpcClient modelPB.ModelPublicServiceClient, modelName string, inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	if len(inputs) <= 0 {
		return nil, fmt.Errorf("invalid input: %v for model: %s", inputs, modelName)
	}

	if grpcClient == nil {
		return nil, fmt.Errorf("uninitialized client")
	}

	outputs := []*structpb.Struct{}

	for _, input := range inputs {

		prompt := input.GetFields()["prompt"].GetStringValue()
		imageToImageInput := &modelPB.ImageToImageInput{
			Prompt: &prompt,
		}
		if _, ok := input.GetFields()["steps"]; ok {
			v := int32(input.GetFields()["steps"].GetNumberValue())
			imageToImageInput.Steps = &v
		}
		if _, ok := input.GetFields()["image_base64"]; ok {
			imageToImageInput.Type = &modelPB.ImageToImageInput_PromptImageBase64{
				PromptImageBase64: base.TrimBase64Mime(input.GetFields()["image_base64"].GetStringValue()),
			}
		}
		if _, ok := input.GetFields()["temperature"]; ok {
			v := int32(input.GetFields()["temperature"].GetNumberValue())
			imageToImageInput.Seed = &v
		}
		if _, ok := input.GetFields()["cfg_scale"]; ok {
			v := float32(input.GetFields()["cfg_scale"].GetNumberValue())
			imageToImageInput.CfgScale = &v
		}

		if _, ok := input.GetFields()["seed"]; ok {
			v := int32(input.GetFields()["seed"].GetNumberValue())
			imageToImageInput.Seed = &v
		}
		extraParams := []*modelPB.ExtraParamObject{}
		if _, ok := input.GetFields()["extra_params"]; ok {
			for _, item := range input.GetFields()["extra_params"].GetListValue().AsSlice() {
				extraParams = append(extraParams, &modelPB.ExtraParamObject{
					ParamName:  item.(map[string]interface{})["param_name"].(string),
					ParamValue: item.(map[string]interface{})["param_value"].(string),
				})
			}
			imageToImageInput.ExtraParams = extraParams
		}

		taskInput := &modelPB.TaskInput_ImageToImage{
			ImageToImage: imageToImageInput,
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

		imageToImageOutput := taskOutputs[0].GetImageToImage()
		if imageToImageOutput == nil {
			return nil, fmt.Errorf("invalid output: %v for model: %s", imageToImageOutput, modelName)
		}
		for imageIdx := range imageToImageOutput.Images {
			imageToImageOutput.Images[imageIdx] = fmt.Sprintf("data:image/jpeg;base64,%s", imageToImageOutput.Images[imageIdx])
		}

		outputJson, err := protojson.MarshalOptions{
			UseProtoNames:   true,
			EmitUnpopulated: true,
		}.Marshal(imageToImageOutput)
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
