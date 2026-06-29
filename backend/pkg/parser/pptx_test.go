package parser

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"
)

func makePptx(slides []string) []byte {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	entry, _ := zw.Create("[Content_Types].xml")
	entry.Write([]byte(`<?xml version="1.0"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"></Types>`))

	for i, text := range slides {
		var sb strings.Builder
		sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
		sb.WriteString(`<p:sld xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main" xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main">`)
		sb.WriteString(`<p:cSld><p:spTree><p:sp><p:txBody><a:p><a:r><a:t>`)
		sb.WriteString(text)
		sb.WriteString(`</a:t></a:r></a:p></p:txBody></p:sp></p:spTree></p:cSld></p:sld>`)
		slideEntry, _ := zw.Create("ppt/slides/slide" + string(rune('1'+i)) + ".xml")
		slideEntry.Write([]byte(sb.String()))
	}
	zw.Close()
	return buf.Bytes()
}

func TestPptxParser_Slides(t *testing.T) {
	data := makePptx([]string{"Slide one content", "Slide two content"})
	units := collectBytes(t, &PptxParser{}, data)
	if len(units) != 2 {
		t.Fatalf("want 2 units, got %d", len(units))
	}
	if units[0].Metadata["slide"] != "1" {
		t.Errorf("want slide=1, got %q", units[0].Metadata["slide"])
	}
	if !strings.Contains(units[0].Text, "Slide one") {
		t.Errorf("unexpected text: %q", units[0].Text)
	}
}

func TestPptxParser_Empty(t *testing.T) {
	data := makePptx([]string{})
	units := collectBytes(t, &PptxParser{}, data)
	if len(units) != 0 {
		t.Fatalf("want 0 units, got %d", len(units))
	}
}
