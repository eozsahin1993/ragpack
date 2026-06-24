package chunker

import (
	"iter"
	"strings"

	"ragpack/pkg/parser"
)

// SlidingWindowChunker applies a fixed-size sliding window over the stream of
// units without loading the full document. It maintains a rolling rune buffer
// of at most ChunkSize+Overlap+one_unit characters; once the buffer fills past
// ChunkSize it emits a chunk and discards everything before the overlap tail.
type SlidingWindowChunker struct{ cfg Config }

func (c *SlidingWindowChunker) Chunk(units iter.Seq2[parser.Unit, error]) iter.Seq2[Chunk, error] {
	return func(yield func(Chunk, error) bool) {
		step := c.cfg.ChunkSize - c.cfg.Overlap
		if step <= 0 {
			yield(Chunk{}, errBadConfig(c.cfg.ChunkSize, c.cfg.Overlap))
			return
		}

		var buf []rune
		idx := 0

		// flush emits chunks while the buffer has enough content, then
		// slides the window forward. Returns false if the caller stopped.
		flush := func(final bool) bool {
			for {
				if len(buf) < c.cfg.ChunkSize && !final {
					return true
				}
				end := c.cfg.ChunkSize
				if end > len(buf) {
					end = len(buf)
				}
				text := strings.TrimSpace(string(buf[:end]))
				if text != "" {
					if !yield(Chunk{Text: text, Index: idx}, nil) {
						return false
					}
					idx++
				}
				if end == len(buf) {
					buf = buf[:0]
					return true
				}
				// Copy the overlap tail into a fresh slice so the old backing
				// array (potentially megabytes) can be garbage collected.
				tail := buf[step:]
				next := make([]rune, len(tail))
				copy(next, tail)
				buf = next
			}
		}

		for unit, err := range units {
			if err != nil {
				yield(Chunk{}, err)
				return
			}
			if t := strings.TrimSpace(unit.Text); t != "" {
				if len(buf) > 0 {
					buf = append(buf, '\n')
				}
				buf = append(buf, []rune(t)...)
				if !flush(false) {
					return
				}
			}
		}
		flush(true)
	}
}
