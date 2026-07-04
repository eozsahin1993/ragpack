package parser

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// wantParagraphs are the 5 expected paragraphs from testdata/sample.pdf.
var wantParagraphs = []string{
	"A paragraph is an organized set of sentences that deal with a single topic.",
	"Long, unbroken blocks of text often appear daunting to the reader.",
	"To be effective, paragraphs should contain certain elements.",
	"A paragraph also should be coherent.",
	"There is no minimum or maximum length for paragraphs.",
}

func TestPDFParser_SamplePDF(t *testing.T) {
	if _, err := exec.LookPath("pdftohtml"); err != nil {
		t.Skip("pdftohtml not in PATH — install poppler-utils to run this test")
	}

	file, err := os.Open("testdata/sample.pdf")
	if err != nil {
		t.Fatalf("open sample.pdf: %v", err)
	}
	units := collectBytes(t, &PDFParser{}, mustRead(t, file))

	if len(units) < len(wantParagraphs) {
		t.Fatalf("want at least %d paragraph units, got %d:\n%s",
			len(wantParagraphs), len(units), strings.Join(unitTexts(units), "\n"))
	}
	for _, u := range units {
		if u.Kind != UnitKindParagraph {
			t.Errorf("want kind %q, got %q", UnitKindParagraph, u.Kind)
		}
	}

	for _, want := range wantParagraphs {
		found := false
		for _, u := range units {
			if strings.HasPrefix(u.Text, want) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("no unit starts with: %q", want)
		}
	}
}

func unitTexts(units []Unit) []string {
	out := make([]string, len(units))
	for i, u := range units {
		out[i] = u.Text
	}
	return out
}
