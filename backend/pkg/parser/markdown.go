package parser

import (
	"ragpack/pkg/util"
	"bufio"
	"context"
	"io"
	"iter"
	"strings"
)

// MarkdownParser streams sections from Markdown documents.
// Each section spans from one ATX heading to the next; the heading breadcrumb
// is attached as metadata["heading"] and the body text is emitted separately.
type MarkdownParser struct{ title string }

func (p *MarkdownParser) Parse(_ context.Context, r io.ReadCloser) iter.Seq2[Unit, error] {
	return func(yield func(Unit, error) bool) {
		defer r.Close()

		type header struct {
			level int
			title string
		}

		var stack []header
		var body strings.Builder
		breadcrumb := ""

		flush := func() bool {
			b := strings.TrimSpace(body.String())
			body.Reset()
			if b == "" {
				return true
			}
			meta := map[string]string{"heading": breadcrumb}
			return yield(Unit{Kind: UnitKindSection, Text: b, Metadata: meta}, nil)
		}

		scanner := bufio.NewScanner(r)
		// Raise past the 64KB default; a single line can legitimately exceed it.
		scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
		for scanner.Scan() {
			line := scanner.Text()
			level := markdownHeaderLevel(line)
			if level > 0 {
				if !flush() {
					return
				}
				title := strings.TrimSpace(line[level+1:])
				if p.title == "" && level == 1 {
					p.title = title
				}
				for len(stack) > 0 && stack[len(stack)-1].level >= level {
					stack = stack[:len(stack)-1]
				}
				stack = append(stack, header{level: level, title: title})
				parts := make([]string, len(stack))
				for i, h := range stack {
					parts[i] = strings.Repeat("#", h.level) + " " + h.title
				}
				breadcrumb = strings.Join(parts, "\n")
			} else {
				body.WriteString(line)
				body.WriteByte('\n')
			}
		}
		flush()

		if err := scanner.Err(); err != nil {
			yield(Unit{}, err)
		}
	}
}

// markdownHeaderLevel returns the ATX heading level (1–6) of line, or 0.
func markdownHeaderLevel(line string) int {
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


func (p *MarkdownParser) Title() *string { return util.NonEmptyStr(p.title) }
