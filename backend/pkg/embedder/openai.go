package embedder

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

type OpenAIEmbedder struct {
	apiKey  string
	model   string
	baseURL string
	dims    int
	client  *http.Client
}

func NewOpenAI(ctx context.Context, apiKey, model string) (*OpenAIEmbedder, error) {
	return NewOpenAICompatible(ctx, apiKey, model, openAIDefaultBaseURL, false, 30*time.Second)
}

func NewOpenAICompatible(ctx context.Context, apiKey, model, baseURL string, disableCompression bool, timeout time.Duration) (*OpenAIEmbedder, error) {
	transport := http.DefaultTransport
	if disableCompression {
		transport = &http.Transport{DisableCompression: true}
	}
	e := &OpenAIEmbedder{
		apiKey:  apiKey,
		model:   model,
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: timeout, Transport: transport},
	}

	return e, nil
}

func (e *OpenAIEmbedder) Model() string { return e.model }

func (e *OpenAIEmbedder) Dimensions() (int, error) {
	if e.dims == 0 {
		dims, err := probeDimensions(context.Background(), e)
		if err != nil {
			return 0, err
		}
		e.dims = dims
	}
	return e.dims, nil
}

func (e *OpenAIEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, Usage, error) {
	body, err := json.Marshal(map[string]any{
		"input": texts,
		"model": e.model,
	})
	if err != nil {
		return nil, Usage{}, fmt.Errorf("openai embedder: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.baseURL+"/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, Usage{}, fmt.Errorf("openai embedder: build request: %w", err)
	}
	if e.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+e.apiKey)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, Usage{}, fmt.Errorf("openai embedder: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errBody map[string]any
		json.NewDecoder(resp.Body).Decode(&errBody)
		return nil, Usage{}, fmt.Errorf("openai embedder: status %d: %v", resp.StatusCode, errBody)
	}

	var result struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
			Index     int       `json:"index"`
		} `json:"data"`
		Usage struct {
			TotalTokens int `json:"total_tokens"`
		} `json:"usage"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, Usage{}, fmt.Errorf("openai embedder: decode response: %w", err)
	}

	// sort by index to guarantee order matches input
	embeddings := make([][]float32, len(result.Data))
	for _, d := range result.Data {
		embeddings[d.Index] = d.Embedding
	}

	return embeddings, Usage{TotalTokens: result.Usage.TotalTokens}, nil
}
