package chunker

import (
	"iter"
	"strings"

	"ragpack/pkg/parser"
)

// ParagraphChunker accumulates units (paragraphs) until cfg.ChunkSize is
// reached, then emits a chunk. Only the overlap tail is kept between chunks —
// the full document is never held in memory.
type ParagraphChunker struct{ cfg Config }

func (c *ParagraphChunker) Chunk(units iter.Seq2[parser.Unit, error]) iter.Seq2[Chunk, error] {
	return func(yield func(Chunk, error) bool) {
		var window []string // paragraphs in the current chunk window
		windowSize := 0    // total rune count in window (with separators)
		idx := 0

		emit := func() bool {
			if len(window) == 0 {
				return true
			}
			text := strings.TrimSpace(strings.Join(window, "\n"))
			if text == "" {
				return true
			}
			if len([]rune(text)) > c.cfg.ChunkSize {
				// Single paragraph exceeded the limit — split with sliding window.
				var sub []Chunk
				idx = splitOversize(text, idx, c.cfg, nil, &sub)
				for _, ch := range sub {
					if !yield(ch, nil) {
						return false
					}
				}
			} else {
				if !yield(Chunk{Text: text, Index: idx}, nil) {
					return false
				}
				idx++
			}
			return true
		}

		// carryOverlap trims the window down to just the trailing paragraphs
		// that cover cfg.Overlap characters, ready for the next chunk.
		carryOverlap := func() {
			overlap := 0
			start := len(window)
			for start > 0 {
				start--
				overlap += len([]rune(window[start])) + 1
				if overlap >= c.cfg.Overlap {
					break
				}
			}
			window = append([]string(nil), window[start:]...) // copy, free old backing array
			windowSize = 0
			for _, p := range window {
				windowSize += len([]rune(p)) + 1
			}
		}

		for unit, err := range units {
			if err != nil {
				yield(Chunk{}, err)
				return
			}
			para := strings.TrimSpace(unit.Text)
			if para == "" {
				continue
			}
			paraSize := len([]rune(para)) + 1 // +1 for \n separator

			if windowSize > 0 && windowSize+paraSize > c.cfg.ChunkSize {
				if !emit() {
					return
				}
				carryOverlap()
			}

			window = append(window, para)
			windowSize += paraSize
		}

		emit()
	}
}
