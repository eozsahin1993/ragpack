package mocks

import (
	"context"

	"ragpack/pkg/llm"
)

// LLM echoes a deterministic string — enough to exercise RAG's response
// shape (answer + chunks) without a real completion API.
type LLM struct{}

func (LLM) Complete(_ context.Context, prompt string) (string, llm.Usage, error) {
	return "fake answer for: " + prompt[:min(40, len(prompt))], llm.Usage{}, nil
}
func (LLM) Model() string { return "mock-llm" }
