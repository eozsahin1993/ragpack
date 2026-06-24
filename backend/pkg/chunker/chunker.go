package chunker

import (
	"fmt"
	"iter"
	"strings"

	"ragpack/pkg/parser"
)

// Chunk is a fixed-size piece of text ready for embedding.
type Chunk struct {
	Text  string
	Index int
}

// Config controls chunking behaviour.
type Config struct {
	ChunkSize int // max characters per chunk
	Overlap   int // characters carried into the next chunk
}

func DefaultConfig() Config {
	return Config{ChunkSize: 2000, Overlap: 200}
}

// Chunker groups parser units into embeddable text chunks.
// It receives a lazy stream of units and returns a lazy stream of chunks —
// nothing is materialised until the caller iterates.
type Chunker interface {
	Chunk(units iter.Seq2[parser.Unit, error]) iter.Seq2[Chunk, error]
}

// New returns the appropriate Chunker for the given MIME type.
func New(mimeType string, cfg Config) (Chunker, error) {
	switch {
	case mimeType == "text/markdown" || mimeType == "text/html":
		return &SectionChunker{cfg: cfg}, nil
	case strings.HasPrefix(mimeType, "text/"):
		return &ParagraphChunker{cfg: cfg}, nil
	case mimeType == "application/pdf":
		return &SlidingWindowChunker{cfg: cfg}, nil
	case mimeType == "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return &ParagraphChunker{cfg: cfg}, nil
	case mimeType == "application/vnd.openxmlformats-officedocument.presentationml.presentation":
		return &UnitChunker{cfg: cfg}, nil
	case mimeType == "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		return &RowGroupChunker{cfg: cfg}, nil
	default:
		return nil, fmt.Errorf("chunker: unsupported mime type %q", mimeType)
	}
}
