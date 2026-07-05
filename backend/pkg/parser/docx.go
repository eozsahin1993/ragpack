package parser

import (
	"ragpack/pkg/util"
	"context"
	"fmt"
	"io"
	"iter"
)

// DocxParser streams one Unit per paragraph from Word Open XML documents.
type DocxParser struct{ title string }

func (p *DocxParser) Parse(_ context.Context, r io.ReadCloser) iter.Seq2[Unit, error] {
	return func(yield func(Unit, error) bool) {
		zr, err := openXMLReader(r)
		if err != nil {
			yield(Unit{}, fmt.Errorf("docx: unzip: %w", err))
			return
		}

		rc, err := openXMLEntry(zr, "word/document.xml")
		if err != nil {
			yield(Unit{}, fmt.Errorf("docx: %w", err))
			return
		}
		defer rc.Close()

		// w:t = text run, w:p = paragraph boundary
		if err := xmlStreamText(rc, "t", "p", func(text string) bool {
			if p.title == "" {
				p.title = text
			}
			return yield(Unit{Kind: UnitKindParagraph, Text: text}, nil)
		}); err != nil {
			yield(Unit{}, fmt.Errorf("docx: parse: %w", err))
		}
	}
}


func (p *DocxParser) Title() *string { return util.NonEmptyStr(p.title) }
