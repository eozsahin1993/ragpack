// Package mocks provides deterministic, no-network test doubles for the
// integration test suite (backend/test/integration) — real embedding/LLM
// providers need API keys or a running service, which the suite avoids.
package mocks

import (
	"context"
	"hash/fnv"
	"strings"

	"ragpack/pkg/embedder"
)

// Embedder hashes words into a fixed-size bag-of-words vector, so shared
// vocabulary between texts produces genuine (if crude) cosine similarity —
// enough to exercise ranking without a real embedding API.
type Embedder struct {
	Dim       int
	ModelName string
}

func (e *Embedder) Embed(_ context.Context, texts []string) ([][]float32, embedder.Usage, error) {
	vecs := make([][]float32, len(texts))
	for i, t := range texts {
		v := make([]float32, e.Dim)
		for _, word := range strings.Fields(strings.ToLower(t)) {
			h := fnv.New32a()
			h.Write([]byte(word))
			v[int(h.Sum32())%e.Dim] += 1
		}
		vecs[i] = v
	}
	return vecs, embedder.Usage{}, nil
}

func (e *Embedder) Dimensions() (int, error) { return e.Dim, nil }
func (e *Embedder) Model() string            { return e.ModelName }
