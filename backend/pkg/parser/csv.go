package parser

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"iter"
	"strings"
)

// CSVParser streams one Unit per data row from CSV files.
// The header row is stored in each unit's metadata so chunkers can prepend it
// without needing to track state across units.
type CSVParser struct{}

func (p *CSVParser) Parse(_ context.Context, r io.ReadCloser) iter.Seq2[Unit, error] {
	return func(yield func(Unit, error) bool) {
		defer r.Close()

		reader := csv.NewReader(r)
		reader.TrimLeadingSpace = true

		headerRow, err := reader.Read()
		if err == io.EOF {
			return
		}
		if err != nil {
			yield(Unit{}, fmt.Errorf("csv: read header: %w", err))
			return
		}
		header := strings.Join(headerRow, "\t")

		for {
			row, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				yield(Unit{}, fmt.Errorf("csv: read row: %w", err))
				return
			}
			text := strings.Join(row, "\t")
			if strings.TrimSpace(text) == "" {
				continue
			}
			meta := map[string]string{
				"headers": header,
				"sheet":   "csv",
			}
			if !yield(Unit{Kind: UnitKindRow, Text: text, Metadata: meta}, nil) {
				return
			}
		}
	}
}
