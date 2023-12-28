package stabilityai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/instill-ai/connector/pkg/util"
)

const (
	successFinishReason = "SUCCESS"
)

type TextToImageInput struct {
	Task               string     `json:"task"`
	Prompts            []string   `json:"prompts"`
	Engine             string     `json:"engine"`
	Weights            *[]float64 `json:"weights,omitempty"`
	Height             *uint32    `json:"height,omitempty"`
	Width              *uint32    `json:"width,omitempty"`
	CfgScale           *float64   `json:"cfg_scale,omitempty"`
	ClipGuidancePreset *string    `json:"clip_guidance_preset,omitempty"`
	Sampler            *string    `json:"sampler,omitempty"`
	Samples            *uint32    `json:"samples,omitempty"`
	Seed               *uint32    `json:"seed,omitempty"`
	Steps              *uint32    `json:"steps,omitempty"`
	StylePreset        *string    `json:"style_preset,omitempty"`
}

type TextToImageOutput struct {
	Images []string `json:"images"`
	Seeds  []uint32 `json:"seeds"`
}

// TextToImageReq represents the request body for text-to-image API
type TextToImageReq struct {
	TextPrompts        []TextPrompt `json:"text_prompts" om:"texts[:]"`
	CFGScale           *float64     `json:"cfg_scale,omitempty" om:"metadata.cfg_scale"`
	ClipGuidancePreset *string      `json:"clip_guidance_preset,omitempty" om:"metadata.clip_guidance_preset"`
	Sampler            *string      `json:"sampler,omitempty" om:"metadata.sampler"`
	Samples            *uint32      `json:"samples,omitempty" om:"metadata.samples"`
	Seed               *uint32      `json:"seed,omitempty" om:"metadata.seed"`
	Steps              *uint32      `json:"steps,omitempty" om:"metadata.steps"`
	StylePreset        *string      `json:"style_preset,omitempty" om:"metadata.style_preset"`
	Height             *uint32      `json:"height,omitempty" om:"metadata.height"`
	Width              *uint32      `json:"width,omitempty" om:"metadata.width"`
}

// TextPrompt holds a prompt's text and its weight.
type TextPrompt struct {
	Text   string   `json:"text" om:"."`
	Weight *float64 `json:"weight"`
}

// Image represents a single image
type Image struct {
	Base64       string `json:"base64"`
	Seed         uint32 `json:"seed"`
	FinishReason string `json:"finishReason"`
}

// ImageTaskRes represents the response body for text-to-image API
type ImageTaskRes struct {
	Images []Image `json:"artifacts"`
}

// GenerateImageFromText makes a call to the text-to-image API from Stability AI.
// https://platform.stability.ai/rest-api#tag/v1generation/operation/textToImage
func (c *Client) GenerateImageFromText(params TextToImageReq, engine string) (results []Image, err error) {
	var resp ImageTaskRes
	if engine == "" {
		return nil, fmt.Errorf("no engine selected")
	}
	textToImageURL := host + "/v1/generation/" + engine + "/text-to-image"
	data, _ := json.Marshal(params)
	err = c.sendReq(textToImageURL, http.MethodPost, util.MIMETypeJSON, bytes.NewBuffer(data), &resp)
	for _, i := range resp.Images {
		if i.FinishReason == successFinishReason {
			results = append(results, i)
		}
	}
	return
}
