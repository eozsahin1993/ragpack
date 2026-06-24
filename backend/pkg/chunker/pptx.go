package chunker

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
)

// PPTXChunker treats each slide as one chunk, preserving slide boundaries as
// semantic units. Slides that exceed cfg.ChunkSize are split with the standard
// sliding-window text chunker so no content is lost.
type PPTXChunker struct{ cfg Config }

func (c *PPTXChunker) Chunk(_ context.Context, r io.ReadCloser) ([]Chunk, error) {
	zr, err := openXMLReader(r)
	if err != nil {
		return nil, fmt.Errorf("pptx: unzip: %w", err)
	}

	var slideNames []string
	for _, f := range zr.File {
		if strings.HasPrefix(f.Name, "ppt/slides/slide") && strings.HasSuffix(f.Name, ".xml") {
			slideNames = append(slideNames, f.Name)
		}
	}
	sort.Strings(slideNames)

	var chunks []Chunk
	for slideNum, name := range slideNames {
		rc, err := openXMLEntry(zr, name)
		if err != nil {
			continue
		}
		// a:t = text run, a:p = paragraph boundary
		text, err := xmlText(rc, "t", "p")
		rc.Close()
		if err != nil || strings.TrimSpace(text) == "" {
			continue
		}

		if len([]rune(text)) <= c.cfg.ChunkSize {
			// Whole slide fits in one chunk — preserve slide as a unit.
			chunks = append(chunks, Chunk{Text: text, Index: slideNum})
		} else {
			// Slide is unusually long; split within the slide boundary.
			sub, err := (&TextChunker{cfg: c.cfg}).Chunk(
				context.Background(), io.NopCloser(strings.NewReader(text)),
			)
			if err == nil {
				for _, ch := range sub {
					chunks = append(chunks, Chunk{Text: ch.Text, Index: slideNum})
				}
			}
		}
	}

	return chunks, nil
}
