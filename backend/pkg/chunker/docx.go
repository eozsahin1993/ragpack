package chunker

import (
	"context"
	"fmt"
	"io"
	"strings"
)

// DOCXChunker chunks Word documents paragraph-by-paragraph.
// Paragraphs are accumulated until cfg.ChunkSize is reached; the last
// cfg.Overlap characters worth of paragraphs are carried into the next chunk
// to preserve cross-boundary context.
type DOCXChunker struct{ cfg Config }

func (c *DOCXChunker) Chunk(_ context.Context, r io.ReadCloser) ([]Chunk, error) {
	zr, err := openXMLReader(r)
	if err != nil {
		return nil, fmt.Errorf("docx: unzip: %w", err)
	}
	rc, err := openXMLEntry(zr, "word/document.xml")
	if err != nil {
		return nil, fmt.Errorf("docx: %w", err)
	}
	defer rc.Close()

	// w:t = text run, w:p = paragraph boundary
	raw, err := xmlText(rc, "t", "p")
	if err != nil {
		return nil, fmt.Errorf("docx: parse: %w", err)
	}

	return chunkByParagraphs(raw, c.cfg), nil
}

// chunkByParagraphs groups newline-delimited paragraphs into chunks of up to
// cfg.ChunkSize characters, carrying cfg.Overlap characters of trailing
// paragraphs into the next chunk.
func chunkByParagraphs(text string, cfg Config) []Chunk {
	var paras []string
	for _, p := range strings.Split(text, "\n") {
		if p = strings.TrimSpace(p); p != "" {
			paras = append(paras, p)
		}
	}
	if len(paras) == 0 {
		return nil
	}

	var chunks []Chunk
	start, idx := 0, 0

	for start < len(paras) {
		var sb strings.Builder
		i := start
		for i < len(paras) {
			candidate := sb.String()
			if sb.Len() > 0 {
				candidate += "\n"
			}
			candidate += paras[i]
			if sb.Len() > 0 && len([]rune(candidate)) > cfg.ChunkSize {
				break
			}
			if sb.Len() > 0 {
				sb.WriteByte('\n')
			}
			sb.WriteString(paras[i])
			i++
		}

		if chunkText := strings.TrimSpace(sb.String()); chunkText != "" {
			if len([]rune(chunkText)) > cfg.ChunkSize {
				// Single paragraph exceeded the size limit — split it with the
				// sliding-window chunker so we don't emit an oversized chunk.
				sub, err := (&TextChunker{cfg: cfg}).Chunk(
					context.Background(), io.NopCloser(strings.NewReader(chunkText)),
				)
				if err == nil {
					for _, ch := range sub {
						chunks = append(chunks, Chunk{Text: ch.Text, Index: idx})
						idx++
					}
				}
			} else {
				chunks = append(chunks, Chunk{Text: chunkText, Index: idx})
				idx++
			}
		}

		// Carry back enough trailing paragraphs to cover cfg.Overlap characters.
		overlap, next := 0, i
		for next > start+1 {
			next--
			overlap += len([]rune(paras[next])) + 1
			if overlap >= cfg.Overlap {
				break
			}
		}
		if next == start {
			next = start + 1
		}
		start = next
	}

	return chunks
}
