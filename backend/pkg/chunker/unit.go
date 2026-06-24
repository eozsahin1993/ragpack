package chunker

import (
	"iter"
	"strings"

	"ragpack/pkg/parser"
)

// UnitChunker emits one chunk per unit. If a unit exceeds cfg.ChunkSize it
// is split with the sliding window so no content is lost. Used for PPTX where
// each slide is the natural semantic boundary.
type UnitChunker struct{ cfg Config }

func (c *UnitChunker) Chunk(units iter.Seq2[parser.Unit, error]) iter.Seq2[Chunk, error] {
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
