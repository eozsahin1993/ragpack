package chunker

import (
	"iter"
	"strings"

	"ragpack/pkg/parser"
)

// RowGroupChunker groups spreadsheet row units into chunks, prepending the
// header row (from unit metadata) to each chunk so results are self-contained.
// Used for XLSX.
type RowGroupChunker struct{ cfg Config }

func (c *RowGroupChunker) Chunk(units iter.Seq2[parser.Unit, error]) iter.Seq2[Chunk, error] {
	return func(yield func(Chunk, error) bool) {
		idx := 0

		// Group units by sheet so the header resets per sheet.
		type sheetState struct {
			header string
			buf    strings.Builder
		}
		sheets := map[string]*sheetState{}
		var sheetOrder []string

		for unit, err := range units {
			if err != nil {
				yield(Chunk{}, err)
				return
			}

			sheet := unit.Metadata["sheet"]
			header := unit.Metadata["headers"]

			st, ok := sheets[sheet]
			if !ok {
				st = &sheetState{header: header}
				sheets[sheet] = st
				sheetOrder = append(sheetOrder, sheet)
				st.buf.WriteString(header)
				st.buf.WriteByte('\n')
			}

			line := strings.TrimSpace(unit.Text)
			if line == "" {
				continue
			}

			candidate := st.buf.String() + line + "\n"
			headerLen := len(st.header) + 1
			if st.buf.Len() > headerLen && len([]rune(candidate)) > c.cfg.ChunkSize {
				if text := strings.TrimSpace(st.buf.String()); text != "" {
					if !yield(Chunk{Text: text, Index: idx}, nil) {
						return
					}
					idx++
				}
				st.buf.Reset()
				st.buf.WriteString(st.header)
				st.buf.WriteByte('\n')
			}
			st.buf.WriteString(line)
			st.buf.WriteByte('\n')
		}

		// Flush remaining buffers in sheet order.
		for _, sheet := range sheetOrder {
			st := sheets[sheet]
			if text := strings.TrimSpace(st.buf.String()); text != "" {
				if !yield(Chunk{Text: text, Index: idx}, nil) {
					return
				}
				idx++
			}
		}
	}
}
