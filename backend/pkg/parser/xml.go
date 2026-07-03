package parser

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"iter"
	"strings"
)

// XMLParser streams one Unit per direct child of the root element.
// Nested elements are flattened to dot-notation key: value lines,
// consistent with the JSON parser.
type XMLParser struct{}

func (p *XMLParser) Parse(_ context.Context, r io.ReadCloser) iter.Seq2[Unit, error] {
	return func(yield func(Unit, error) bool) {
		defer r.Close()

		dec := xml.NewDecoder(r)

		depth := 0
		var pathStack []string
		var textBuf strings.Builder
		var recordBuf strings.Builder
		var recordElement string
		inRecord := false

		for {
			tok, err := dec.Token()
			if err == io.EOF {
				break
			}
			if err != nil {
				yield(Unit{}, fmt.Errorf("xml: parse: %w", err))
				return
			}

			switch t := tok.(type) {
			case xml.StartElement:
				depth++
				if depth == 2 {
					recordElement = t.Name.Local
					recordBuf.Reset()
					pathStack = pathStack[:0]
					inRecord = true
				} else if depth > 2 && inRecord {
					pathStack = append(pathStack, t.Name.Local)
					textBuf.Reset()
				}

			case xml.EndElement:
				if depth > 2 && inRecord {
					text := strings.TrimSpace(textBuf.String())
					if text != "" && len(pathStack) > 0 {
						recordBuf.WriteString(strings.Join(pathStack, "."))
						recordBuf.WriteString(": ")
						recordBuf.WriteString(text)
						recordBuf.WriteByte('\n')
					}
					if len(pathStack) > 0 {
						pathStack = pathStack[:len(pathStack)-1]
					}
					textBuf.Reset()
				} else if depth == 2 && inRecord {
					text := strings.TrimSpace(recordBuf.String())
					if text != "" {
						meta := map[string]string{"element": recordElement}
						if !yield(Unit{Kind: UnitKindRow, Text: text, Metadata: meta}, nil) {
							return
						}
					}
					inRecord = false
				}
				depth--

			case xml.CharData:
				if inRecord && depth > 2 {
					textBuf.Write(t)
				}
			}
		}
	}
}
