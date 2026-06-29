package parser

import (
	"os"
	"strings"
	"testing"
)

func TestPDFParser_RealDocument(t *testing.T) {
	file, err := os.Open("testdata/sample.pdf")
	if err != nil {
		t.Fatalf("open sample.pdf: %v", err)
	}
	units := collectBytes(t, &PDFParser{}, mustRead(t, file))

	if len(units) != 1 {
		t.Fatalf("want 1 unit (single page), got %d", len(units))
	}

	unit := units[0]

	if unit.Metadata["page"] != "1" {
		t.Errorf("want page=1, got %q", unit.Metadata["page"])
	}

	if strings.TrimSpace(unit.Text) == "" {
		t.Error("unit text is empty")
	}

	if !strings.HasPrefix(unit.Text, "Paragraphs") {
		t.Errorf("unit text unexpected start: %q", unit.Text[:min(50, len(unit.Text))])
	}

	mustContainPrefix := []string{
		"A paragraph is an organized set",
		"Long, unbroken blocks of text",
		"To be effect", // PDF extraction splits "effective" across lines
		"There is no minimum or maximum length",
	}
	for _, prefix := range mustContainPrefix {
		if !strings.Contains(unit.Text, prefix) {
			t.Errorf("unit text missing expected content starting with %q", prefix)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
