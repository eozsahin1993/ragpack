package embedder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const openAIEmbeddingsURL = "https://api.openai.com/v1/embeddings"

type OpenAIEmbedder struct {
	apiKey string
	model  string
	dims   int
	client *http.Client
}

func NewOpenAI(ctx context.Context, apiKey, model string) (*OpenAIEmbedder, error) {
	e := &OpenAIEmbedder{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{Timeout: 30 * time.Second},
	}

	vecs, err := e.Embed(ctx, []string{"probe"})
	if err != nil {
		return nil, fmt.Errorf("openai embedder: probe call failed: %w", err)
	}
	e.dims = len(vecs[0])

	return e, nil
}

func (e *OpenAIEmbedder) Dimensions() int { return e.dims }

func (e *OpenAIEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	body, err := json.Marshal(map[string]any{
		"input": texts,
		"model": e.model,
	})
	if err != nil {
		return nil, fmt.Errorf("openai embedder: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, openAIEmbeddingsURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("openai embedder: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+e.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openai embedder: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errBody map[string]any
		json.NewDecoder(resp.Body).Decode(&errBody)
		return nil, fmt.Errorf("openai embedder: status %d: %v", resp.StatusCode, errBody)
	}

	var result struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
			Index     int       `json:"index"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("openai embedder: decode response: %w", err)
	}

	// sort by index to guarantee order matches input
	embeddings := make([][]float32, len(result.Data))
	for _, d := range result.Data {
		embeddings[d.Index] = d.Embedding
	}

	return embeddings, nil
}
