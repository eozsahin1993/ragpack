package chunker

import (
	"iter"
	"strings"

	"ragpack/pkg/parser"
)

// SectionChunker emits one chunk per Markdown/HTML section. Each unit already
// carries the heading breadcrumb in its text (prepended by the parser), so
// retrieved chunks retain hierarchy context. Oversized sections are split with
// the sliding window. Used for Markdown and HTML.
type SectionChunker struct{ cfg Config }

func (c *SectionChunker) Chunk(units iter.Seq2[parser.Unit, error]) iter.Seq2[Chunk, error] {
	return func(yield func(Chunk, error) bool) {
		idx := 0
		for unit, err := range units {
			if err != nil {
				yield(Chunk{}, err)
				return
			}
			text := strings.TrimSpace(unit.Text)
			if text == "" {
				continue
			}

			if len([]rune(text)) <= c.cfg.ChunkSize {
				if !yield(Chunk{Text: text, Index: idx}, nil) {
					return
				}
				idx++
			} else {
				var sub []Chunk
				idx = splitOversize(text, idx, c.cfg, &sub)
				for _, ch := range sub {
					if !yield(ch, nil) {
						return
					}
				}
			}
		}
	}
}
