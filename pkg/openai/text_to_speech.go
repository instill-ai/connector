package openai

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/instill-ai/connector/pkg/util"
)

const (
	createSpeechURL = host + "/v1/audio/speech"
)

type TextToSpeechInput struct {
	Text           string   `json:"text"`
	Model          string   `json:"model"`
	Voice          string   `json:"voice"`
	ResponseFormat *string  `json:"response_format,omitempty"`
	Speed          *float64 `json:"speed,omitempty"`
}

type TextToSpeechOutput struct {
	Audio string `json:"audio"`
}

type TextToSpeechReq struct {
	Input          string   `json:"input"`
	Model          string   `json:"model"`
	Voice          string   `json:"voice"`
	ResponseFormat *string  `json:"response_format,omitempty"`
	Speed          *float64 `json:"speed,omitempty"`
}

func (c *Client) CreateSpeech(req TextToSpeechReq) (TextToSpeechOutput, error) {
	data, _ := json.Marshal(req)
	body, err := c.sendReq(createSpeechURL, http.MethodPost, util.MIMETypeJSON, bytes.NewBuffer(data))
	if err != nil {
		return TextToSpeechOutput{}, err
	}

	result := TextToSpeechOutput{
		Audio: base64.StdEncoding.EncodeToString(body),
	}

	return result, nil
}
