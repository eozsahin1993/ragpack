package embedder

import (
	"context"
	"fmt"
	"log"

	"ragpack/pkg/config"
)

type Registry struct {
	embedders    map[string]Embedder
	defaultModel string
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
			registry.Register(emb)
			log.Printf("embedder registered: openai/%s", emb.Model())
		}
	}

	if cfg.Ollama.BaseURL != "" && cfg.Ollama.Model != "" {
		emb, err := NewOllama(ctx, cfg.Ollama.BaseURL, cfg.Ollama.Model)
		if err != nil {
			log.Printf("warning: failed to init Ollama embedder (%s): %v", cfg.Ollama.Model, err)
		} else {
			registry.Register(emb)
			log.Printf("embedder registered: ollama/%s", emb.Model())
		}
	}

	if cfg.TEI.BaseURL != "" && cfg.TEI.Model != "" {
		emb, err := NewOpenAICompatible(ctx, "", cfg.TEI.Model, cfg.TEI.BaseURL+"/v1")
		if err != nil {
			log.Printf("warning: failed to init TEI embedder (%s): %v", cfg.TEI.Model, err)
		} else {
			registry.Register(emb)
			log.Printf("embedder registered: tei/%s", emb.Model())
		}
	}

	if cfg.HuggingFace.APIKey != "" && cfg.HuggingFace.Model != "" {
		emb, err := NewHuggingFace(ctx, cfg.HuggingFace.APIKey, cfg.HuggingFace.Model)
		if err != nil {
			log.Printf("warning: failed to init HuggingFace embedder (%s): %v", cfg.HuggingFace.Model, err)
		} else {
			registry.Register(emb)
			log.Printf("embedder registered: huggingface/%s", emb.Model())
		}
	}

	switch cfg.EmbedProvider {
	case "openai":
		registry.defaultModel = cfg.OpenAI.Model
	case "ollama":
		registry.defaultModel = cfg.Ollama.Model
	case "tei":
		registry.defaultModel = cfg.TEI.Model
	case "huggingface":
		registry.defaultModel = cfg.HuggingFace.Model
	case "":
		// no default set — Default() will auto-detect if only one embedder is registered
	default:
		log.Printf("warning: unknown DEFAULT_EMBED_PROVIDER %q — valid values: openai, ollama, tei, huggingface", cfg.EmbedProvider)
	}

	return registry
}

func (r *Registry) Register(emb Embedder) {
	r.embedders[emb.Model()] = emb
}

func (r *Registry) Get(model string) (Embedder, error) {
	emb, ok := r.embedders[model]
	if !ok {
		return nil, fmt.Errorf("embed model %q not found in registry — ensure it is configured in .env", model)
	}
	return emb, nil
}

func (r *Registry) Default() (string, Embedder, error) {
	if r.defaultModel == "" && len(r.embedders) == 1 {
		for model, emb := range r.embedders {
			return model, emb, nil
		}
	}
	if r.defaultModel == "" {
		return "", nil, fmt.Errorf("no default embed model — set DEFAULT_EMBED_PROVIDER in .env")
	}
	emb, err := r.Get(r.defaultModel)
	if err != nil {
		return "", nil, fmt.Errorf("default embed model %q is not registered — check .env configuration", r.defaultModel)
	}
	return r.defaultModel, emb, nil
}
