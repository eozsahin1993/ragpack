package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const openAIDefaultBaseURL = "https://api.openai.com/v1"

type OpenAILLM struct {
	apiKey    string
	model     string
	baseURL   string
	maxTokens int
	client    *http.Client
}

func NewOpenAI(apiKey, model string) *OpenAILLM {
	return NewOpenAICompatible(apiKey, model, openAIDefaultBaseURL, 30*time.Second)
}

func NewOpenAICompatible(apiKey, model, baseURL string, timeout time.Duration) *OpenAILLM {
	return &OpenAILLM{
		apiKey:    apiKey,
		model:     model,
		baseURL:   strings.TrimRight(baseURL, "/"),
		maxTokens: 1024,
		client:    &http.Client{Timeout: timeout},
	}
}

func (l *OpenAILLM) Model() string { return l.model }

func (l *OpenAILLM) Complete(ctx context.Context, prompt string) (string, Usage, error) {
	body, err := json.Marshal(map[string]any{
		"model":      l.model,
		"max_tokens": l.maxTokens,
		"messages":   []map[string]string{{"role": "user", "content": prompt}},
	})
	if err != nil {
		return "", Usage{}, fmt.Errorf("openai llm: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, l.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", Usage{}, fmt.Errorf("openai llm: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if l.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+l.apiKey)
	}

	resp, err := l.client.Do(req)
	if err != nil {
		return "", Usage{}, fmt.Errorf("openai llm: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errBody map[string]any
		json.NewDecoder(resp.Body).Decode(&errBody)
		return "", Usage{}, fmt.Errorf("openai llm: status %d: %v", resp.StatusCode, errBody)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", Usage{}, fmt.Errorf("openai llm: decode response: %w", err)
	}
	if len(result.Choices) == 0 {
		return "", Usage{}, fmt.Errorf("openai llm: empty response")
	}

	usage := Usage{InputTokens: result.Usage.PromptTokens, OutputTokens: result.Usage.CompletionTokens}
	return result.Choices[0].Message.Content, usage, nil
}
