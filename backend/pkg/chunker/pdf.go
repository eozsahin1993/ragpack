package chunker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/ledongthuc/pdf"
)

type PDFChunker struct {
	cfg Config
}

func (c *PDFChunker) Chunk(_ context.Context, r io.ReadCloser) ([]Chunk, error) {
	defer r.Close()

	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("pdf chunker: read: %w", err)
	}

	reader, err := pdf.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		return nil, fmt.Errorf("pdf chunker: parse: %w", err)
	}

	var sb strings.Builder
	for i := 1; i <= reader.NumPage(); i++ {
		page := reader.Page(i)
		if page.V.IsNull() {
			continue
		}
		text, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}
		sb.WriteString(text)
		sb.WriteString("\n")
	}

	content := strings.TrimSpace(sb.String())
	if content == "" {
		return nil, nil
	}

	// Reuse the text sliding-window chunker on the extracted content.
	tc := &TextChunker{cfg: c.cfg}
	return tc.Chunk(context.Background(), io.NopCloser(strings.NewReader(content)))
}
