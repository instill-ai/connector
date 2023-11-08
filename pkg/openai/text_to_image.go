package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	generationURL = host + "/v1/images/generations"
)

type ImagesGenerationInput struct {
	Prompt  string  `json:"prompt"`
	Model   string  `json:"model"`
	N       *int    `json:"n,omitempty"`
	Quality *string `json:"quality,omitempty"`
	Size    *string `json:"size,omitempty"`
	Style   *string `json:"style,omitempty"`
}

type ImageGenerationsOutputResult struct {
	Image         string `json:"image"`
	RevisedPrompt string `json:"revised_prompt"`
}
type ImageGenerationsOutput struct {
	Results []ImageGenerationsOutputResult `json:"results"`
}

type ImageGenerationsReq struct {
	Prompt         string  `json:"prompt"`
	Model          string  `json:"model"`
	N              *int    `json:"n,omitempty"`
	Quality        *string `json:"quality,omitempty"`
	Size           *string `json:"size,omitempty"`
	Style          *string `json:"style,omitempty"`
	ResponseFormat string  `json:"response_format"`
}

type ImageGenerationsRespData struct {
	Image         string `json:"b64_json"`
	RevisedPrompt string `json:"revised_prompt"`
}
type ImageGenerationsResp struct {
	Data []ImageGenerationsRespData `json:"data"`
}

func (c *Client) GenerateImagesGenerations(req ImageGenerationsReq) (result ImageGenerationsResp, err error) {
	data, _ := json.Marshal(req)
	fmt.Println("datra", string(data))
	err = c.sendReq(generationURL, http.MethodPost, jsonMimeType, bytes.NewBuffer(data), &result)
	return result, err
}
