package chunker

import (
	"fmt"
	"iter"
	"strings"

	"ragpack/pkg/parser"
)

// Chunk is a fixed-size piece of text ready for embedding.
type Chunk struct {
	Text   string
	Index  int
	Header *string // optional heading breadcrumb, e.g. "Introduction > Background"
}

// Strategy constants select how parser units are grouped into embeddable chunks.
//
//   - Auto:          picks the strategy below based on the file's MIME type.
//   - Paragraph:     groups blank-line-separated paragraphs up to ChunkSize chars,
//                    carrying Overlap chars of the previous group into the next.
//                    Best for prose: plain text, DOCX.
//   - SlidingWindow: fixed-size rolling window over raw text with Overlap chars
//                    of carry-over. Best for dense continuous text: PDFs.
//   - Section:       one chunk per heading section; oversized sections are split
//                    with the sliding-window fallback. Best for Markdown, HTML.
//   - Unit:          one chunk per parser unit (e.g. one slide). Oversized units
//                    are split. Best for PPTX and other slide-based formats.
//   - RowGroup:      groups spreadsheet rows up to ChunkSize chars, prepending
//                    the header row to each group. Best for XLSX.
const (
	StrategyAuto          = "auto"
	StrategyUnit          = "unit"
	StrategyParagraph     = "paragraph"
	StrategySlidingWindow = "sliding_window"
	StrategySection       = "section"
	StrategyRowGroup      = "row_group"
)

// Config controls chunking behaviour.
type Config struct {
	ChunkSize int    // max characters per chunk
	Overlap   int    // characters carried into the next chunk
	Strategy  string // "auto" (MIME-based) | "unit" | "paragraph" | "sliding_window" | "section" | "row_group"
}

// Chunker groups parser units into embeddable text chunks.
// It receives a lazy stream of units and returns a lazy stream of chunks —
// nothing is materialised until the caller iterates.
type Chunker interface {
	Chunk(units iter.Seq2[parser.Unit, error]) iter.Seq2[Chunk, error]
}

// New returns the appropriate Chunker. When cfg.Strategy is "auto" or empty,
// the strategy is selected from the MIME type; otherwise the explicit strategy wins.
func New(mimeType string, cfg Config) (Chunker, error) {
	strategy := cfg.Strategy
	if strategy == "" || strategy == StrategyAuto {
		strategy = mimeStrategy(mimeType)
		if strategy == "" {
			return nil, fmt.Errorf("chunker: unsupported mime type %q", mimeType)
		}
	}
	switch strategy {
	case StrategySection:
		return &SectionChunker{cfg: cfg}, nil
	case StrategyParagraph:
		return &ParagraphChunker{cfg: cfg}, nil
	case StrategySlidingWindow:
		return &SlidingWindowChunker{cfg: cfg}, nil
	case StrategyUnit:
		return &UnitChunker{cfg: cfg}, nil
	case StrategyRowGroup:
		return &RowGroupChunker{cfg: cfg}, nil
	default:
		return nil, fmt.Errorf("chunker: unknown strategy %q", strategy)
	}
}

// mimeStrategy maps a MIME type to the default strategy name.
func mimeStrategy(mimeType string) string {
	switch {
	case mimeType == "text/markdown" || mimeType == "text/html":
		return StrategySection
	case strings.HasPrefix(mimeType, "text/"):
		return StrategyParagraph
	case mimeType == "application/pdf":
		return StrategySlidingWindow
	case mimeType == "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return StrategyParagraph
	case mimeType == "application/vnd.openxmlformats-officedocument.presentationml.presentation":
		return StrategyUnit
	case mimeType == "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		return StrategyRowGroup
	default:
		return ""
	}
}
