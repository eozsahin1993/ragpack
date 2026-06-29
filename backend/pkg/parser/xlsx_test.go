package parser

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"
)

func makeXlsx(headers []string, rows [][]string) []byte {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	entry, _ := zw.Create("[Content_Types].xml")
	entry.Write([]byte(`<?xml version="1.0"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"></Types>`))

	var sb strings.Builder
	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	sb.WriteString(`<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData>`)

	writeRow := func(cells []string) {
		sb.WriteString(`<row>`)
		for _, cell := range cells {
			sb.WriteString(`<c t="inlineStr"><is><t>`)
			sb.WriteString(cell)
			sb.WriteString(`</t></is></c>`)
		}
		sb.WriteString(`</row>`)
	}

	writeRow(headers)
	for _, row := range rows {
		writeRow(row)
	}
	sb.WriteString(`</sheetData></worksheet>`)

	sheetEntry, _ := zw.Create("xl/worksheets/sheet1.xml")
	sheetEntry.Write([]byte(sb.String()))
	zw.Close()
	return buf.Bytes()
}

func TestXlsxParser_Rows(t *testing.T) {
	headers := []string{"Name", "Age"}
	rows := [][]string{
		{"Alice", "30"},
		{"Bob", "25"},
	}
	data := makeXlsx(headers, rows)
	units := collectBytes(t, &XlsxParser{}, data)
	if len(units) != 2 {
		t.Fatalf("want 2 units, got %d", len(units))
	}
	if !strings.Contains(units[0].Text, "Alice") {
		t.Errorf("want Alice in first row, got %q", units[0].Text)
	}
	if units[0].Metadata["headers"] != "Name\tAge" {
		t.Errorf("want headers %q, got %q", "Name\tAge", units[0].Metadata["headers"])
	}
	if units[0].Metadata["sheet"] != "sheet1" {
		t.Errorf("want sheet=sheet1, got %q", units[0].Metadata["sheet"])
	}
}

func TestXlsxParser_HeaderOnly(t *testing.T) {
	data := makeXlsx([]string{"Name", "Age"}, nil)
	units := collectBytes(t, &XlsxParser{}, data)
	if len(units) != 0 {
		t.Fatalf("want 0 data units (header only), got %d", len(units))
	}
}
