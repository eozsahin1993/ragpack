package embedder

import (
	"context"
	"fmt"
)

// Usage reports the token count for a single Embed call. Providers that
// don't return usage leave it zero.
type Usage struct {
	TotalTokens int
}

type Embedder interface {
	Embed(ctx context.Context, texts []string) ([][]float32, Usage, error)
	Dimensions() (int, error)
	Model() string
}

func probeDimensions(ctx context.Context, e Embedder) (int, error) {
	vecs, _, err := e.Embed(ctx, []string{"probe"})
	if err != nil {
		return 0, err
	}
	if len(vecs) == 0 || len(vecs[0]) == 0 {
		return 0, fmt.Errorf("empty embedding response")
	}
	return len(vecs[0]), nil
}
