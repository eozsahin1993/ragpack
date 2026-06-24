package chunker

import (
	"context"
	"fmt"
	"io"
	"strings"
)

type MarkdownChunker struct {
	cfg Config
}

type mdSection struct {
	breadcrumb string // accumulated parent headers, e.g. "# H1\n## H2"
	body       string // content beneath this header
}

func (c *MarkdownChunker) Chunk(ctx context.Context, r io.ReadCloser) ([]Chunk, error) {
	defer r.Close()

	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("markdown chunker: read: %w", err)
	}

	sections := splitMarkdownSections(strings.TrimSpace(string(raw)))

	var chunks []Chunk
	idx := 0

	for _, sec := range sections {
		full := strings.TrimSpace(sec.breadcrumb + "\n" + sec.body)
		if full == "" {
			continue
		}

		if len([]rune(full)) <= c.cfg.ChunkSize {
			chunks = append(chunks, Chunk{Text: full, Index: idx})
			idx++
			continue
		}

		// Section too large: accumulate paragraphs up to chunk size.
		paragraphs := strings.Split(full, "\n\n")
		var buf strings.Builder
		for _, para := range paragraphs {
			para = strings.TrimSpace(para)
			if para == "" {
				continue
			}
			pending := len([]rune(buf.String())) + len([]rune(para))
			if pending > c.cfg.ChunkSize && buf.Len() > 0 {
				chunks = append(chunks, Chunk{Text: strings.TrimSpace(buf.String()), Index: idx})
				idx++
				buf.Reset()
			}
			// Single paragraph larger than chunk size: use sliding window.
			if len([]rune(para)) > c.cfg.ChunkSize {
				tc := &TextChunker{cfg: c.cfg}
				sub, _ := tc.Chunk(ctx, io.NopCloser(strings.NewReader(para)))
				for _, sc := range sub {
					chunks = append(chunks, Chunk{Text: sc.Text, Index: idx})
					idx++
				}
				continue
			}
			if buf.Len() > 0 {
				buf.WriteString("\n\n")
			}
			buf.WriteString(para)
		}
		if buf.Len() > 0 {
			chunks = append(chunks, Chunk{Text: strings.TrimSpace(buf.String()), Index: idx})
			idx++
		}
	}

	return chunks, nil
}

// splitMarkdownSections splits a Markdown document into sections at each header
// boundary. Each section carries its full breadcrumb of parent headers so
// retrieved chunks retain context about where in the document they came from.
func splitMarkdownSections(content string) []mdSection {
	type header struct {
		level int
		title string
	}

	var sections []mdSection
	var stack []header
	var body strings.Builder
	breadcrumb := ""

	flush := func() {
		b := strings.TrimSpace(body.String())
		if breadcrumb != "" || b != "" {
			sections = append(sections, mdSection{breadcrumb: breadcrumb, body: b})
		}
		body.Reset()
	}

	for _, line := range strings.Split(content, "\n") {
		level := headerLevel(line)
		if level > 0 {
			flush()
			title := strings.TrimSpace(line[level+1:])
			// Pop headers at same or deeper level.
			for len(stack) > 0 && stack[len(stack)-1].level >= level {
				stack = stack[:len(stack)-1]
			}
			stack = append(stack, header{level: level, title: title})
			// Build breadcrumb from current stack.
			parts := make([]string, len(stack))
			for i, h := range stack {
				parts[i] = strings.Repeat("#", h.level) + " " + h.title
			}
			breadcrumb = strings.Join(parts, "\n")
		} else {
			body.WriteString(line)
			body.WriteString("\n")
		}
	}
	flush()

	return sections
}

// headerLevel returns the ATX header level (1–6) of a line, or 0 if not a header.
func headerLevel(line string) int {
	if len(line) < 2 {
		return 0
	}
	level := 0
	for _, ch := range line {
		if ch == '#' {
			level++
		} else {
			break
		}
	}
	if level == 0 || level > 6 {
		return 0
	}
	if len(line) <= level || line[level] != ' ' {
		return 0
	}
	return level
}
