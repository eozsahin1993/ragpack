package llm

import (
	"fmt"
	"log"

	"ragpack/pkg/config"
)

type Registry struct {
	providers    map[string]LLM
	localModels  map[string]bool
	defaultModel string
}

func NewRegistry() *Registry {
	return &Registry{providers: make(map[string]LLM), localModels: make(map[string]bool)}
}

func NewRegistryFromConfig(cfg config.Config) *Registry {
	registry := NewRegistry()

	if cfg.OpenAI.APIKey != "" && cfg.OpenAI.LLMModel != "" {
		llm := NewOpenAI(cfg.OpenAI.APIKey, cfg.OpenAI.LLMModel)
		registry.Register(llm)
		log.Printf("llm registered: openai/%s", llm.Model())
	}

	if cfg.Ollama.BaseURL != "" && cfg.Ollama.LLMModel != "" {
		llm := NewOllama(cfg.Ollama.BaseURL, cfg.Ollama.LLMModel)
		registry.Register(llm)
		registry.localModels[llm.Model()] = true
		log.Printf("llm registered: ollama/%s", llm.Model())
	}

	if cfg.Anthropic.APIKey != "" && cfg.Anthropic.Model != "" {
		llm := NewAnthropic(cfg.Anthropic.APIKey, cfg.Anthropic.Model)
		registry.Register(llm)
		log.Printf("llm registered: anthropic/%s", llm.Model())
	}

	switch cfg.LLMProvider {
	case "openai":
		registry.defaultModel = cfg.OpenAI.LLMModel
	case "ollama":
		registry.defaultModel = cfg.Ollama.LLMModel
	case "anthropic":
		registry.defaultModel = cfg.Anthropic.Model
	case "":
		// no default — Default() will auto-detect if only one provider is registered
	default:
		log.Printf("warning: unknown DEFAULT_LLM_PROVIDER %q — valid values: openai, ollama, anthropic", cfg.LLMProvider)
	}

	return registry
}

func (r *Registry) Models() []string {
	models := make([]string, 0, len(r.providers))
	for model := range r.providers {
		models = append(models, model)
	}
	return models
}

func (r *Registry) Register(l LLM) {
	r.providers[l.Model()] = l
}

func (r *Registry) Get(model string) (LLM, error) {
	l, ok := r.providers[model]
	if !ok {
		return nil, fmt.Errorf("llm model %q not found in registry — ensure it is configured in .env", model)
	}
	return l, nil
}

// IsLocal reports whether model runs on a local/self-hosted provider (Ollama),
// which is a confirmed $0 cost — distinct from a hosted model with no pricing
// table entry, which is unpriced/unknown. Ollama model names are user-chosen
// and can't be recognized by a pricing lookup, so locality is tracked here
// at registration time instead.
func (r *Registry) IsLocal(model string) bool {
	return r.localModels[model]
}

func (r *Registry) Default() (string, LLM, error) {
	if r.defaultModel == "" && len(r.providers) == 1 {
		for model, l := range r.providers {
			return model, l, nil
		}
	}
	if r.defaultModel == "" {
		return "", nil, fmt.Errorf("no default LLM — set LLM_PROVIDER in .env")
	}
	l, err := r.Get(r.defaultModel)
	if err != nil {
		return "", nil, fmt.Errorf("default llm model %q is not registered — check .env configuration", r.defaultModel)
	}
	return r.defaultModel, l, nil
}
