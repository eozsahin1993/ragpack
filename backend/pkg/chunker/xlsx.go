package chunker

import (
	"archive/zip"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"sort"
	"strings"
)

// XLSXChunker groups spreadsheet rows into chunks. The header row (row 0 of
// each sheet) is prepended to every chunk so retrieved results are self-contained
// without needing to look up the original sheet.
type XLSXChunker struct{ cfg Config }

func (c *XLSXChunker) Chunk(_ context.Context, r io.ReadCloser) ([]Chunk, error) {
	zr, err := openXMLReader(r)
	if err != nil {
		return nil, fmt.Errorf("xlsx: unzip: %w", err)
	}

	shared, _ := xlsxSharedStrings(zr) // nil is safe; xlsxRows handles it

	var sheetNames []string
	for _, f := range zr.File {
		if strings.HasPrefix(f.Name, "xl/worksheets/sheet") && strings.HasSuffix(f.Name, ".xml") {
			sheetNames = append(sheetNames, f.Name)
		}
	}
	sort.Strings(sheetNames)

	var chunks []Chunk
	globalIdx := 0

	for _, name := range sheetNames {
		rc, err := openXMLEntry(zr, name)
		if err != nil {
			continue
		}
		rows, err := xlsxRows(rc, shared)
		rc.Close()
		if err != nil || len(rows) == 0 {
			continue
		}

		header := strings.Join(rows[0], "\t")
		data := rows[1:]

		if len(data) == 0 {
			// Sheet has only a header row — emit it as its own chunk.
			chunks = append(chunks, Chunk{Text: header, Index: globalIdx})
			globalIdx++
			continue
		}

		var sb strings.Builder
		sb.WriteString(header)
		sb.WriteByte('\n')

		for _, row := range data {
			line := strings.Join(row, "\t")
			candidate := sb.String() + line + "\n"
			if sb.Len() > len(header)+1 && len([]rune(candidate)) > c.cfg.ChunkSize {
				// Flush current chunk and start a new one with the header.
				if text := strings.TrimSpace(sb.String()); text != "" {
					chunks = append(chunks, Chunk{Text: text, Index: globalIdx})
					globalIdx++
				}
				sb.Reset()
				sb.WriteString(header)
				sb.WriteByte('\n')
			}
			sb.WriteString(line)
			sb.WriteByte('\n')
		}
		if text := strings.TrimSpace(sb.String()); text != "" {
			chunks = append(chunks, Chunk{Text: text, Index: globalIdx})
			globalIdx++
		}
	}

	return chunks, nil
}

// xlsxSharedStrings parses xl/sharedStrings.xml and returns the ordered string
// table. Excel stores most cell strings here and references them by index.
func xlsxSharedStrings(zr *zip.Reader) ([]string, error) {
	rc, err := openXMLEntry(zr, "xl/sharedStrings.xml")
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	dec := xml.NewDecoder(rc)
	var table []string
	var cur strings.Builder
	capture := false

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "si" {
				cur.Reset()
			} else if t.Name.Local == "t" {
				capture = true
			}
		case xml.EndElement:
			if t.Name.Local == "t" {
				capture = false
			} else if t.Name.Local == "si" {
				table = append(table, cur.String())
			}
		case xml.CharData:
			if capture {
				cur.Write(t)
			}
		}
	}
	return table, nil
}

// xlsxRows parses a worksheet XML and returns each row as a slice of cell
// strings. Shared-string indices are resolved against the provided table.
func xlsxRows(r io.Reader, shared []string) ([][]string, error) {
	dec := xml.NewDecoder(r)
	var rows [][]string
	var curRow []string
	var curVal strings.Builder
	var cellType string
	inV := false

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "row":
				curRow = nil
			case "c":
				cellType = ""
				for _, a := range t.Attr {
					if a.Name.Local == "t" {
						cellType = a.Value
					}
				}
				curVal.Reset()
			case "v", "t":
				inV = true
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "v", "t":
				inV = false
			case "c":
				val := curVal.String()
				if cellType == "s" && shared != nil {
					var idx int
					fmt.Sscan(val, &idx)
					if idx >= 0 && idx < len(shared) {
						val = shared[idx]
					}
				}
				curRow = append(curRow, val)
			case "row":
				if len(curRow) > 0 {
					rows = append(rows, curRow)
				}
			}
		case xml.CharData:
			if inV {
				curVal.Write(t)
			}
		}
	}
	return rows, nil
}
