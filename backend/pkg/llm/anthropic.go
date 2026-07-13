package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const anthropicMessagesURL = "https://api.anthropic.com/v1/messages"
const anthropicVersion = "2023-06-01"

type AnthropicLLM struct {
	apiKey    string
	model     string
	maxTokens int
	client    *http.Client
}

func NewAnthropic(apiKey, model string) *AnthropicLLM {
	return &AnthropicLLM{
		apiKey:    apiKey,
		model:     model,
		maxTokens: 1024,
		client:    &http.Client{Timeout: 60 * time.Second},
	}
}

func (l *AnthropicLLM) Model() string { return l.model }

func (l *AnthropicLLM) Complete(ctx context.Context, prompt string) (string, Usage, error) {
	body, err := json.Marshal(map[string]any{
		"model":      l.model,
		"max_tokens": l.maxTokens,
		"messages":   []map[string]string{{"role": "user", "content": prompt}},
	})
	if err != nil {
		return "", Usage{}, fmt.Errorf("anthropic llm: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, anthropicMessagesURL, bytes.NewReader(body))
	if err != nil {
		return "", Usage{}, fmt.Errorf("anthropic llm: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", l.apiKey)
	req.Header.Set("anthropic-version", anthropicVersion)

	resp, err := l.client.Do(req)
	if err != nil {
		return "", Usage{}, fmt.Errorf("anthropic llm: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errBody map[string]any
		json.NewDecoder(resp.Body).Decode(&errBody)
		return "", Usage{}, fmt.Errorf("anthropic llm: status %d: %v", resp.StatusCode, errBody)
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", Usage{}, fmt.Errorf("anthropic llm: decode response: %w", err)
	}
	if len(result.Content) == 0 {
		return "", Usage{}, fmt.Errorf("anthropic llm: empty response")
	}

	usage := Usage{InputTokens: result.Usage.InputTokens, OutputTokens: result.Usage.OutputTokens}
	return result.Content[0].Text, usage, nil
}
