package parser

import (
	"ragpack/pkg/util"
	"bufio"
	"context"
	"io"
	"iter"
	"strings"
)

// TextParser streams paragraphs from plain text files.
// A blank line signals a paragraph boundary; if no blank lines exist the
// whole file is emitted as a single unit.
type TextParser struct{ title string }

func (p *TextParser) Parse(_ context.Context, r io.ReadCloser) iter.Seq2[Unit, error] {
	return func(yield func(Unit, error) bool) {
		defer r.Close()

		scanner := bufio.NewScanner(r)
		var cur strings.Builder

		flush := func() bool {
			if text := strings.TrimSpace(cur.String()); text != "" {
				cur.Reset()
				return yield(Unit{Kind: UnitKindParagraph, Text: text}, nil)
			}
			cur.Reset()
			return true
		}

		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				if !flush() {
					return
				}
			} else {
				if cur.Len() > 0 {
					cur.WriteByte('\n')
				} else if p.title == "" {
					p.title = line
				}
				cur.WriteString(line)
			}
		}
		flush()

		if err := scanner.Err(); err != nil {
			yield(Unit{}, err)
		}
	}
}

func (p *TextParser) Title() *string { return util.NonEmptyStr(p.title) }
