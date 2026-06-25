package embedder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const hfInferenceBaseURL = "https://api-inference.huggingface.co"

type HuggingFaceEmbedder struct {
	apiKey string
	model  string
	dims   int
	client *http.Client
}

func NewHuggingFace(ctx context.Context, apiKey, model string) (*HuggingFaceEmbedder, error) {
	e := &HuggingFaceEmbedder{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{Timeout: 2 * time.Minute},
	}

	return e, nil
}

func (e *HuggingFaceEmbedder) Model() string { return e.model }

func (e *HuggingFaceEmbedder) Dimensions() int {
	if e.dims == 0 {
		if dims, err := probeDimensions(context.Background(), e); err == nil {
			e.dims = dims
		}
	}
	return e.dims
}

func (e *HuggingFaceEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	body, err := json.Marshal(map[string]any{
		"inputs":  texts,
		"options": map[string]bool{"wait_for_model": true},
	})
	if err != nil {
		return nil, fmt.Errorf("huggingface embedder: marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/pipeline/feature-extraction/%s", hfInferenceBaseURL, e.model)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("huggingface embedder: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+e.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("huggingface embedder: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errBody map[string]any
		json.NewDecoder(resp.Body).Decode(&errBody)
		return nil, fmt.Errorf("huggingface embedder: status %d: %v", resp.StatusCode, errBody)
	}

	var raw json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("huggingface embedder: decode response: %w", err)
	}

	return parseHFEmbeddings(raw, len(texts))
}

func parseHFEmbeddings(raw json.RawMessage, n int) ([][]float32, error) {
	var embeddings [][]float32
	if err := json.Unmarshal(raw, &embeddings); err != nil {
		return nil, fmt.Errorf("huggingface embedder: unexpected response shape — use a sentence-transformer model (e.g. BAAI/bge-small-en-v1.5): %w", err)
	}
	if len(embeddings) != n {
		return nil, fmt.Errorf("huggingface embedder: expected %d embeddings, got %d", n, len(embeddings))
	}
	return embeddings, nil
}
