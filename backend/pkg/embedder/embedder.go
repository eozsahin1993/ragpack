package embedder

import (
	"context"
	"fmt"
)

type Embedder interface {
	Embed(ctx context.Context, texts []string) ([][]float32, error)
	Dimensions() int
	Model() string
}

func probeDimensions(ctx context.Context, e Embedder) (int, error) {
	vecs, err := e.Embed(ctx, []string{"probe"})
	if err != nil {
		return 0, err
	}
	if len(vecs) == 0 || len(vecs[0]) == 0 {
		return 0, fmt.Errorf("empty embedding response")
	}
	return len(vecs[0]), nil
}
