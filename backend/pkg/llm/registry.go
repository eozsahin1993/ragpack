package llm

import (
	"fmt"
	"log"

	"ragpack/pkg/config"
)

type Registry struct {
	providers    map[string]LLM
	defaultModel string
}

func NewRegistry() *Registry {
	return &Registry{providers: make(map[string]LLM)}
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
