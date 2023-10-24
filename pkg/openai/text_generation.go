package openai

import (
	"bytes"
	"encoding/json"
	"net/http"
)

const (
	completionsURL = host + "/v1/chat/completions"
)

type TextCompletionInput struct {
	Prompt        string   `json:"prompt"`
	Model         string   `json:"model"`
	SystemMessage *string  `json:"system_message,omitempty"`
	Temperature   *float32 `json:"temperature,omitempty"`
	N             *int     `json:"n,omitempty"`
	MaxTokens     *int     `json:"max_tokens,omitempty"`
}

type TextCompletionOutput struct {
	Texts []string `json:"texts"`
}

type TextCompletionReq struct {
	Model            string    `json:"model"`
	Messages         []Message `json:"messages"`
	Temperature      *float32  `json:"temperature,omitempty"`
	TopP             *float32  `json:"top_p,omitempty"`
	N                *int      `json:"n,omitempty"`
	Stream           *bool     `json:"stream,omitempty"`
	Stop             *string   `json:"stop,omitempty"`
	MaxTokens        *int      `json:"max_tokens,omitempty"`
	PresencePenalty  *float32  `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float32  `json:"frequency_penalty,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type TextCompletionResp struct {
	ID      string    `json:"id"`
	Object  string    `json:"object"`
	Created int       `json:"created"`
	Choices []Choices `json:"choices"`
	Usage   Usage     `json:"usage"`
}

type Choices struct {
	Index        int     `json:"index"`
	FinishReason string  `json:"finish_reason"`
	Message      Message `json:"message"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// GenerateTextCompletion makes a call to the completions API from OpenAI.
// https://platform.openai.com/docs/api-reference/completions
func (c *Client) GenerateTextCompletion(req TextCompletionReq) (result TextCompletionResp, err error) {
	data, _ := json.Marshal(req)
	err = c.sendReq(completionsURL, http.MethodPost, jsonMimeType, bytes.NewBuffer(data), &result)
	return result, err
}
