package parser

import (
	"context"
	"fmt"
	"io"
	"iter"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// PDFParser extracts one Unit per paragraph using pdftohtml (poppler).
// pdftohtml marks blank lines with &#160;<br/> pairs which we use as
// paragraph separators — no position heuristics needed.
type PDFParser struct{}

func (p *PDFParser) Parse(_ context.Context, r io.ReadCloser) iter.Seq2[Unit, error] {
	return func(yield func(Unit, error) bool) {
		defer r.Close()

		raw, err := io.ReadAll(r)
		if err != nil {
			yield(Unit{}, fmt.Errorf("pdf: read: %w", err))
			return
		}

		units, err := extractParagraphsWithPdftohtml(raw)
		if err != nil {
			yield(Unit{}, err)
			return
		}

		for _, u := range units {
			if !yield(u, nil) {
				return
			}
		}
	}
}

var reTag = regexp.MustCompile(`<[^>]+>`)

// extractParagraphsWithPdftohtml runs pdftohtml and parses its output.
// pdftohtml uses <br/> for line breaks and &#160;<br/> for blank lines;
// consecutive blank lines mark paragraph boundaries.
func extractParagraphsWithPdftohtml(data []byte) ([]Unit, error) {
	tmp, err := os.CreateTemp("", "ragpack-*.pdf")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return nil, err
	}
	tmp.Close()

	out, err := exec.Command("pdftohtml", "-stdout", "-noframes", "-i", "-q", tmp.Name()).Output()
	if err != nil {
		return nil, fmt.Errorf("pdftohtml: %w", err)
	}

	return parsePdftohtmlOutput(string(out)), nil
}

// parsePdftohtmlOutput converts raw pdftohtml HTML into paragraph units.
// pdftohtml uses <br/> for every line break, and &#160;<br/> for blank lines
// that mark paragraph boundaries. We split on <br/> first, so each element is
// one visual line, then group by blank-line separators.
func parsePdftohtmlOutput(src string) []Unit {
	var units []Unit
	var group []string
	sawBlank := false

	for line := range strings.SplitSeq(src, "<br/>") {
		// Strip HTML tags, decode entities, collapse source newlines.
		line = reTag.ReplaceAllString(line, "")
		line = html.UnescapeString(line)
		line = strings.ReplaceAll(line, "\n", " ")
		line = strings.TrimSpace(line)

		if line == "" {
			if len(group) > 0 {
				sawBlank = true
			}
			continue
		}
		if sawBlank {
			units = append(units, Unit{Kind: UnitKindParagraph, Text: strings.Join(group, " ")})
			group = nil
			sawBlank = false
		}
		group = append(group, line)
	}
	if len(group) > 0 {
		units = append(units, Unit{Kind: UnitKindParagraph, Text: strings.Join(group, " ")})
	}
	return units
}

func (p *PDFParser) Title() *string { return nil }
