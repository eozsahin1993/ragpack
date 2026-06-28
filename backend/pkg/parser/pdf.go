package parser

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"iter"
	"strings"

	"github.com/ledongthuc/pdf"
)

// PDFParser streams one Unit per page.
// PDF is a layout format — page boundaries are the only reliable structural
// unit available after text extraction.
type PDFParser struct{}

func (p *PDFParser) Parse(_ context.Context, r io.ReadCloser) iter.Seq2[Unit, error] {
	return func(yield func(Unit, error) bool) {
		defer r.Close()

		raw, err := io.ReadAll(r)
		if err != nil {
			yield(Unit{}, fmt.Errorf("pdf: read: %w", err))
			return
		}

		reader, err := pdf.NewReader(bytes.NewReader(raw), int64(len(raw)))
		if err != nil {
			yield(Unit{}, fmt.Errorf("pdf: parse: %w", err))
			return
		}

		for i := 1; i <= reader.NumPage(); i++ {
			page := reader.Page(i)
			if page.V.IsNull() {
				continue
			}
			text, err := page.GetPlainText(nil)
			if err != nil {
				continue
			}
			text = strings.TrimSpace(text)
			if text == "" {
				continue
			}
			meta := map[string]string{"page": fmt.Sprintf("%d", i)}
			if !yield(Unit{Kind: UnitKindPage, Text: text, Metadata: meta}, nil) {
				return
			}
		}
	}
}
