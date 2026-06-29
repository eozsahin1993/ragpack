package parser

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func mustRead(t *testing.T, file *os.File) []byte {
	t.Helper()
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	return data
}

func makeDocx(paragraphs []string) []byte {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	contentTypes := `<?xml version="1.0" encoding="UTF-8"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`
	entry, _ := zw.Create("[Content_Types].xml")
	entry.Write([]byte(contentTypes))

	var sb strings.Builder
	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	sb.WriteString(`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">`)
	sb.WriteString(`<w:body>`)
	for _, paragraph := range paragraphs {
		sb.WriteString(`<w:p><w:r><w:t>`)
		sb.WriteString(paragraph)
		sb.WriteString(`</w:t></w:r></w:p>`)
	}
	sb.WriteString(`</w:body></w:document>`)

	entry, _ = zw.Create("word/document.xml")
	entry.Write([]byte(sb.String()))
	zw.Close()
	return buf.Bytes()
}

func TestDocxParser_Paragraphs(t *testing.T) {
	data := makeDocx([]string{"First paragraph", "Second paragraph", "Third paragraph"})
	units := collectBytes(t, &DocxParser{}, data)
	if len(units) != 3 {
		t.Fatalf("want 3 units, got %d", len(units))
	}
	if units[0].Text != "First paragraph" {
		t.Errorf("want %q, got %q", "First paragraph", units[0].Text)
	}
}

func TestDocxParser_Empty(t *testing.T) {
	data := makeDocx([]string{})
	units := collectBytes(t, &DocxParser{}, data)
	if len(units) != 0 {
		t.Fatalf("want 0 units, got %d", len(units))
	}
}

func TestDocxParser_RealDocument(t *testing.T) {
	file, err := os.Open("testdata/sample.docx")
	if err != nil {
		t.Fatalf("open sample.docx: %v", err)
	}
	units := collectBytes(t, &DocxParser{}, mustRead(t, file))
	if len(units) < 5 {
		t.Fatalf("want at least 5 units from real docx, got %d", len(units))
	}

	for i, unit := range units {
		if strings.TrimSpace(unit.Text) == "" {
			t.Errorf("unit %d has empty text", i)
		}
	}

	mustContainPrefix := []string{
		"Blog",
		"How is climate change affecting our planet?",
		"Abstract",
		"What is climate change?",
		"How is climate change affecting the environment?",
		"How is climate change affecting life forms?",
		"How can we act against climate change?",
		"Summary",
		"References",
	}
	for _, prefix := range mustContainPrefix {
		found := false
		for _, unit := range units {
			if strings.HasPrefix(unit.Text, prefix) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("no unit starts with %q", prefix)
		}
	}
}
