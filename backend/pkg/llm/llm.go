package llm

import "context"

// Usage reports token counts for a single Complete call. Providers that
// don't return usage (e.g. some Ollama models) leave both fields zero.
type Usage struct {
	InputTokens  int
	OutputTokens int
}

type LLM interface {
	Complete(ctx context.Context, prompt string) (string, Usage, error)
	Model() string
}
