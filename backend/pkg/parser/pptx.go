package parser

import (
	"context"
	"fmt"
	"io"
	"iter"
	"sort"
	"strings"
)

// PptxParser streams one Unit per slide from PowerPoint Open XML files.
type PptxParser struct{}

func (p *PptxParser) Parse(_ context.Context, r io.ReadCloser) iter.Seq2[Unit, error] {
	return func(yield func(Unit, error) bool) {
		zr, err := openXMLReader(r)
		if err != nil {
			yield(Unit{}, fmt.Errorf("pptx: unzip: %w", err))
			return
		}

		var slideNames []string
		for _, f := range zr.File {
			if strings.HasPrefix(f.Name, "ppt/slides/slide") && strings.HasSuffix(f.Name, ".xml") {
				slideNames = append(slideNames, f.Name)
			}
		}
		sort.Strings(slideNames)

		for i, name := range slideNames {
			rc, err := openXMLEntry(zr, name)
			if err != nil {
				continue
			}

			var sb strings.Builder
			// a:t = text run, a:p = paragraph boundary
			err = xmlStreamText(rc, "t", "p", func(text string) bool {
				sb.WriteString(text)
				sb.WriteByte('\n')
				return true
			})
			rc.Close()
			if err != nil {
				yield(Unit{}, fmt.Errorf("pptx: slide %d: %w", i+1, err))
				return
			}

			text := strings.TrimSpace(sb.String())
			if text == "" {
				continue
			}

			meta := map[string]string{"slide": fmt.Sprintf("%d", i+1)}
			if !yield(Unit{Kind: UnitKindSlide, Text: text, Metadata: meta}, nil) {
				return
			}
		}
	}
}
