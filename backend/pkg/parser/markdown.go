package parser

import (
	"bufio"
	"context"
	"io"
	"iter"
	"strings"
)

// MarkdownParser streams sections from Markdown documents.
// Each section spans from one ATX heading to the next; the full breadcrumb of
// parent headings is attached as metadata so retrieved chunks retain hierarchy context.
type MarkdownParser struct{}

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
			if breadcrumb == "" && b == "" {
				return true
			}
			text := strings.TrimSpace(breadcrumb + "\n" + b)
			if text == "" {
				return true
			}
			meta := map[string]string{"heading": breadcrumb}
			return yield(Unit{Text: text, Metadata: meta}, nil)
		}

		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			line := scanner.Text()
			level := markdownHeaderLevel(line)
			if level > 0 {
				if !flush() {
					return
				}
				title := strings.TrimSpace(line[level+1:])
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
