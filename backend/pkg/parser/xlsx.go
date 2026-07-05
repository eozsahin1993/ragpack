package parser

import (
	"archive/zip"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"iter"
	"sort"
	"strings"
)

// XlsxParser streams one Unit per data row from Excel Open XML files.
// The header row is stored in each unit's metadata so chunkers can prepend it
// without needing to track state across units.
type XlsxParser struct{}

func (p *XlsxParser) Parse(_ context.Context, r io.ReadCloser) iter.Seq2[Unit, error] {
	return func(yield func(Unit, error) bool) {
		zr, err := openXMLReader(r)
		if err != nil {
			yield(Unit{}, fmt.Errorf("xlsx: unzip: %w", err))
			return
		}

		shared, _ := xlsxSharedStrings(zr) // nil is safe

		var sheetNames []string
		for _, f := range zr.File {
			if strings.HasPrefix(f.Name, "xl/worksheets/sheet") && strings.HasSuffix(f.Name, ".xml") {
				sheetNames = append(sheetNames, f.Name)
			}
		}
		sort.Strings(sheetNames)

		for _, name := range sheetNames {
			rc, err := openXMLEntry(zr, name)
			if err != nil {
				continue
			}

			sheetName := strings.TrimSuffix(strings.TrimPrefix(name, "xl/worksheets/"), ".xml")
			rows, err := xlsxRows(rc, shared)
			rc.Close()
			if err != nil || len(rows) == 0 {
				continue
			}

			header := strings.Join(rows[0], "\t")
			for _, row := range rows[1:] {
				text := strings.Join(row, "\t")
				meta := map[string]string{
					"headers": header,
					"sheet":   sheetName,
				}
				if !yield(Unit{Kind: UnitKindRow, Text: text, Metadata: meta}, nil) {
					return
				}
			}
		}
	}
}

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

func (p *XlsxParser) Title() *string { return nil }
