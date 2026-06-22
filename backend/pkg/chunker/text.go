package chunker

import (
	"context"
	"fmt"
	"io"
	"strings"
)

type TextChunker struct {
	cfg Config
}

func (c *TextChunker) Chunk(_ context.Context, r io.ReadCloser) ([]Chunk, error) {
	defer r.Close()

	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("text chunker: read: %w", err)
	}

	content := strings.TrimSpace(string(raw))
	if len(content) == 0 {
		return nil, nil
	}

	runes := []rune(content)
	total := len(runes)
	size := c.cfg.ChunkSize
	overlap := c.cfg.Overlap
	step := size - overlap

	if step <= 0 {
		return nil, fmt.Errorf("text chunker: overlap (%d) must be less than chunk size (%d)", overlap, size)
	}

	var chunks []Chunk
	for i, idx := 0, 0; idx < total; i, idx = i+1, idx+step {
		end := idx + size
		if end > total {
			end = total
		}

		text := string(runes[idx:end])

		// skip chunks that are only whitespace
		if strings.TrimSpace(text) == "" {
			continue
		}

		chunks = append(chunks, Chunk{
			Text:  text,
			Index: i,
		})

		if end == total {
			break
		}
	}

	return chunks, nil
}
