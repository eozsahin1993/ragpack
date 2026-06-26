package llm

import "context"

type LLM interface {
	Complete(ctx context.Context, prompt string) (string, error)
	Model() string
}
