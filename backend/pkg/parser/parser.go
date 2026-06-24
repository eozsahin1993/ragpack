package parser

import (
	"context"
	"fmt"
	"io"
	"iter"
	"strings"
)

// Unit is one semantic element emitted by a parser — a paragraph, slide, row,
// or section. Metadata carries format-specific context (heading breadcrumb,
// slide number, sheet name, column headers) so chunkers can make smart
// grouping decisions without knowing the source format.
type Unit struct {
	Text     string
	Metadata map[string]string
}

// Parser streams semantic units from a document.
// Implementations read r and yield one Unit at a time; the caller drives
// iteration via range, so only a small buffer is in memory at once.
type Parser interface {
	Parse(ctx context.Context, r io.ReadCloser) iter.Seq2[Unit, error]
}

// New returns the appropriate Parser for the given MIME type.
func New(mimeType string) (Parser, error) {
	switch {
	case mimeType == "text/markdown":
		return &MarkdownParser{}, nil
	case mimeType == "text/html":
		return &HTMLParser{}, nil
	case strings.HasPrefix(mimeType, "text/"):
		return &TextParser{}, nil
	case mimeType == "application/pdf":
		return &PDFParser{}, nil
	case mimeType == "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return &DocxParser{}, nil
	case mimeType == "application/vnd.openxmlformats-officedocument.presentationml.presentation":
		return &PptxParser{}, nil
	case mimeType == "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		return &XlsxParser{}, nil
	default:
		return nil, fmt.Errorf("parser: unsupported mime type %q", mimeType)
	}
}
