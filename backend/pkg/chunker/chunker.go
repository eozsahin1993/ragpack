package chunker

import (
	"context"
	"fmt"
	"io"
	"strings"
)

type Chunk struct {
	Text  string
	Index int
}

type Chunker interface {
	Chunk(ctx context.Context, r io.ReadCloser) ([]Chunk, error)
}

type Config struct {
	ChunkSize int // characters per chunk
	Overlap   int // overlap between consecutive chunks
}

func DefaultConfig() Config {
	return Config{
		ChunkSize: 2000,
		Overlap:   200,
	}
}

// New returns the appropriate Chunker for the given MIME type.
func New(mimeType string, cfg Config) (Chunker, error) {
	switch {
	case strings.HasPrefix(mimeType, "text/"):
		return &TextChunker{cfg: cfg}, nil
	case mimeType == "application/pdf":
		return nil, fmt.Errorf("chunker: PDF not yet supported")
	case strings.HasPrefix(mimeType, "audio/"):
		return nil, fmt.Errorf("chunker: audio not yet supported")
	case strings.HasPrefix(mimeType, "image/"):
		return nil, fmt.Errorf("chunker: image not yet supported")
	case strings.HasPrefix(mimeType, "video/"):
		return nil, fmt.Errorf("chunker: video not yet supported")
	default:
		return nil, fmt.Errorf("chunker: unsupported mime type %q", mimeType)
	}
}
