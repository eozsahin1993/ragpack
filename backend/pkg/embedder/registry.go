package embedder

import (
	"context"
	"fmt"
	"log"

	"ragpack/pkg/config"
)

type Registry struct {
	embedders map[string]Embedder
}

func NewRegistry() *Registry {
	return &Registry{embedders: make(map[string]Embedder)}
}

func NewRegistryFromConfig(ctx context.Context, cfg config.Config) *Registry {
	registry := NewRegistry()

	if cfg.OpenAI.APIKey != "" && cfg.OpenAI.Model != "" {
		emb, err := NewOpenAI(ctx, cfg.OpenAI.APIKey, cfg.OpenAI.Model)
		if err != nil {
			log.Printf("warning: failed to init OpenAI embedder (%s): %v", cfg.OpenAI.Model, err)
		} else {
			registry.Register(cfg.OpenAI.Model, emb)
			log.Printf("embedder registered: openai/%s", cfg.OpenAI.Model)
		}
	}

	if cfg.Ollama.BaseURL != "" && cfg.Ollama.Model != "" {
		emb, err := NewOllama(ctx, cfg.Ollama.BaseURL, cfg.Ollama.Model)
		if err != nil {
			log.Printf("warning: failed to init Ollama embedder (%s): %v", cfg.Ollama.Model, err)
		} else {
			registry.Register(cfg.Ollama.Model, emb)
			log.Printf("embedder registered: ollama/%s", cfg.Ollama.Model)
		}
	}

	return registry
}

func (r *Registry) Register(model string, emb Embedder) {
	r.embedders[model] = emb
}

func (r *Registry) Get(model string) (Embedder, error) {
	emb, ok := r.embedders[model]
	if !ok {
		return nil, fmt.Errorf("no embedder registered for model %q", model)
	}
	return emb, nil
}

// Default returns the single registered model name and embedder.
// Returns an error if zero or more than one model is registered.
func (r *Registry) Default() (string, Embedder, error) {
	if len(r.embedders) == 0 {
		return "", nil, fmt.Errorf("no embedder configured")
	}
	if len(r.embedders) > 1 {
		return "", nil, fmt.Errorf("multiple embedders registered; specify embed_model explicitly")
	}
	for model, emb := range r.embedders {
		return model, emb, nil
	}
	return "", nil, fmt.Errorf("no embedder configured")
}
