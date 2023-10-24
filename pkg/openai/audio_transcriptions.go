package openai

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"

	"github.com/instill-ai/connector/pkg/util"
)

const (
	transcriptionsURL = host + "/v1/audio/transcriptions"
)

type AudioTranscriptionInput struct {
	Audio       string   `json:"audio"`
	Model       string   `json:"model"`
	Prompt      *string  `json:"prompt,omitempty"`
	Temperature *float64 `json:"temperature,omitempty"`
	Language    *string  `json:"language,omitempty"`
}

type AudioTranscriptionReq struct {
	File        []byte   `json:"file"`
	Model       string   `json:"model"`
	Prompt      *string  `json:"prompt,omitempty"`
	Language    *string  `json:"language,omitempty"`
	Temperature *float64 `json:"temperature,omitempty"`
}

type AudioTranscriptionResp struct {
	Text string `json:"text"`
}

// GenerateAudioTranscriptions makes a call to the audio transcriptions API from OpenAI.
// https://platform.openai.com/docs/api-reference/audio/create-transcription
func (c *Client) GenerateAudioTranscriptions(req AudioTranscriptionReq) (result AudioTranscriptionResp, err error) {
	formData, contentType, err := getBytes(req)
	if err != nil {
		return result, err
	}
	err = c.sendReq(transcriptionsURL, http.MethodPost, contentType, formData, &result)
	return result, err
}

func getBytes(req AudioTranscriptionReq) (*bytes.Reader, string, error) {
	data := &bytes.Buffer{}
	writer := multipart.NewWriter(data)
	err := util.WriteFile(writer, "file", req.File)
	if err != nil {
		return nil, "", err
	}
	util.WriteField(writer, "model", req.Model)
	if req.Prompt != nil {
		util.WriteField(writer, "prompt", *req.Prompt)
	}
	if req.Language != nil {
		util.WriteField(writer, "language", *req.Language)
	}
	if req.Temperature != nil {
		util.WriteField(writer, "temperature", fmt.Sprintf("%f", *req.Temperature))
	}
	writer.Close()
	return bytes.NewReader(data.Bytes()), writer.FormDataContentType(), nil
}
