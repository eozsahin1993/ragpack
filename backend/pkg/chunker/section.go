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

			var header *string
			if h, ok := unit.Metadata["heading"]; ok && h != "" {
				formatted := formatHeading(h)
				header = &formatted
			}

			if len([]rune(text)) <= c.cfg.ChunkSize {
				if !yield(Chunk{Text: text, Index: idx, Header: header}, nil) {
					return
				}
				idx++
			} else {
				var sub []Chunk
				idx = splitOversize(text, idx, c.cfg, header, &sub)
				for _, ch := range sub {
					if !yield(ch, nil) {
						return
					}
				}
			}
		}
	}
}

// formatHeading converts a Markdown breadcrumb like "# H1\n## H2\n### H3"
// into "H1 > H2 > H3".
func formatHeading(breadcrumb string) string {
	lines := strings.Split(breadcrumb, "\n")
	parts := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimLeft(line, "# ")
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return strings.Join(parts, " > ")
}
