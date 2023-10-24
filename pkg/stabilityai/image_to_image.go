package stabilityai

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"

	"github.com/instill-ai/connector/pkg/util"
)

type ImageToImageInput struct {
	Task               string     `json:"task"`
	Engine             string     `json:"engine"`
	Prompts            []string   `json:"prompts"`
	InitImage          string     `json:"init_image"`
	Weights            *[]float64 `json:"weights,omitempty"`
	InitImageMode      *string    `json:"init_image_mode,omitempty"`
	ImageStrength      *float64   `json:"image_strength,omitempty"`
	StepScheduleStart  *float64   `json:"step_schedule_start,omitempty"`
	StepScheduleEnd    *float64   `json:"step_schedule_end,omitempty"`
	CfgScale           *float64   `json:"cfg_scale,omitempty"`
	ClipGuidancePreset *string    `json:"clip_guidance_preset,omitempty"`
	Sampler            *string    `json:"sampler,omitempty"`
	Samples            *uint32    `json:"samples,omitempty"`
	Seed               *uint32    `json:"seed,omitempty"`
	Steps              *uint32    `json:"steps,omitempty"`
	StylePreset        *string    `json:"style_preset,omitempty"`
}

type ImageToImageOutput struct {
	Images []string `json:"images"`
	Seeds  []uint32 `json:"seeds"`
}

// ImageToImageReq represents the request body for image-to-image API
type ImageToImageReq struct {
	TextPrompts        []TextPrompt `json:"text_prompts" om:"texts[:]"`
	InitImage          string       `json:"init_image" om:"images[0]"`
	CFGScale           *float64     `json:"cfg_scale,omitempty" om:"metadata.cfg_scale"`
	ClipGuidancePreset *string      `json:"clip_guidance_preset,omitempty" om:"metadata.clip_guidance_preset"`
	Sampler            *string      `json:"sampler,omitempty" om:"metadata.sampler"`
	Samples            *uint32      `json:"samples,omitempty" om:"metadata.samples"`
	Seed               *uint32      `json:"seed,omitempty" om:"metadata.seed"`
	Steps              *uint32      `json:"steps,omitempty" om:"metadata.steps"`
	StylePreset        *string      `json:"style_preset,omitempty" om:"metadata.style_preset"`
	InitImageMode      *string      `json:"init_image_mode,omitempty" om:"metadata.init_image_mode"`
	ImageStrength      *float64     `json:"image_strength,omitempty" om:"metadata.image_strength"`
	StepScheduleStart  *float64     `json:"step_schedule_start,omitempty" om:"metadata.step_schedule_start"`
	StepScheduleEnd    *float64     `json:"step_schedule_end,omitempty" om:"metadata.step_schedule_end"`
}

// GenerateImageFromImage makes a call to the image-to-image API from Stability AI.
// https://platform.stability.ai/rest-api#tag/v1generation/operation/imageToImage
func (c *Client) GenerateImageFromImage(req ImageToImageReq, engine string) (results []Image, err error) {
	var resp ImageTaskRes
	if engine == "" {
		return nil, fmt.Errorf("no engine selected")
	}
	imageToImageURL := host + "/v1/generation/" + engine + "/image-to-image"
	formData, contentType, err := getBytes(req)
	if err != nil {
		return nil, err
	}
	err = c.sendReq(imageToImageURL, http.MethodPost, contentType, formData, &resp)
	for _, i := range resp.Images {
		if i.FinishReason == successFinishReason {
			results = append(results, i)
		}
	}
	return
}

func getBytes(req ImageToImageReq) (*bytes.Reader, string, error) {
	data := &bytes.Buffer{}
	initImage, err := DecodeBase64(req.InitImage)
	if err != nil {
		return nil, "", err
	}
	writer := multipart.NewWriter(data)
	err = util.WriteFile(writer, "init_image", initImage)
	if err != nil {
		return nil, "", err
	}
	if req.CFGScale != nil {
		util.WriteField(writer, "cfg_scale", fmt.Sprintf("%f", *req.CFGScale))
	}
	if req.ClipGuidancePreset != nil {
		util.WriteField(writer, "clip_guidance_preset", *req.ClipGuidancePreset)
	}
	if req.Sampler != nil {
		util.WriteField(writer, "sampler", *req.Sampler)
	}
	if req.Seed != nil {
		util.WriteField(writer, "seed", fmt.Sprintf("%d", *req.Seed))
	}
	if req.StylePreset != nil {
		util.WriteField(writer, "style_preset", *req.StylePreset)
	}
	if req.InitImageMode != nil {
		util.WriteField(writer, "init_image_mode", *req.InitImageMode)
	}
	if req.ImageStrength != nil {
		util.WriteField(writer, "image_strength", fmt.Sprintf("%f", *req.ImageStrength))
	}
	if req.Samples != nil {
		util.WriteField(writer, "samples", fmt.Sprintf("%d", *req.Samples))
	}
	if req.Steps != nil {
		util.WriteField(writer, "steps", fmt.Sprintf("%d", *req.Steps))
	}

	i := 0
	for _, t := range req.TextPrompts {
		if t.Text == "" {
			continue
		}
		util.WriteField(writer, fmt.Sprintf("text_prompts[%d][text]", i), t.Text)
		if t.Weight != nil {
			util.WriteField(writer, fmt.Sprintf("text_prompts[%d][weight]", i), fmt.Sprintf("%f", *t.Weight))
		}
		i++
	}
	writer.Close()
	return bytes.NewReader(data.Bytes()), writer.FormDataContentType(), nil
}
