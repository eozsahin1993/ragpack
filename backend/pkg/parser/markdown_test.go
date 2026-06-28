package parser

import (
	"strings"
	"testing"
)

func TestMarkdownParser_NoHeadings(t *testing.T) {
	input := "Just some text\nspanning lines"
	units := collect(t, &MarkdownParser{}, input)
	if len(units) != 1 {
		t.Fatalf("want 1 unit, got %d", len(units))
	}
	if units[0].Metadata["heading"] != "" {
		t.Errorf("want empty heading, got %q", units[0].Metadata["heading"])
	}
}

func TestMarkdownParser_SingleSection(t *testing.T) {
	input := "# Title\nBody text here"
	units := collect(t, &MarkdownParser{}, input)
	if len(units) != 1 {
		t.Fatalf("want 1 unit, got %d", len(units))
	}
	if units[0].Metadata["heading"] != "# Title" {
		t.Errorf("want %q, got %q", "# Title", units[0].Metadata["heading"])
	}
	if !strings.Contains(units[0].Text, "Body text") {
		t.Errorf("unit text missing body: %q", units[0].Text)
	}
}

func TestMarkdownParser_MultipleSections(t *testing.T) {
	input := "# First\nFirst body\n\n# Second\nSecond body"
	units := collect(t, &MarkdownParser{}, input)
	if len(units) != 2 {
		t.Fatalf("want 2 units, got %d", len(units))
	}
	if units[0].Metadata["heading"] != "# First" {
		t.Errorf("first heading: want %q got %q", "# First", units[0].Metadata["heading"])
	}
	if units[1].Metadata["heading"] != "# Second" {
		t.Errorf("second heading: want %q got %q", "# Second", units[1].Metadata["heading"])
	}
}

func TestMarkdownParser_NestedHeadings(t *testing.T) {
	input := "# H1\n## H2\n### H3\nDeep body"
	units := collect(t, &MarkdownParser{}, input)
	if len(units) != 1 {
		t.Fatalf("want 1 unit, got %d", len(units))
	}
	heading := units[0].Metadata["heading"]
	if !strings.Contains(heading, "# H1") {
		t.Errorf("breadcrumb missing H1: %q", heading)
	}
	if !strings.Contains(heading, "## H2") {
		t.Errorf("breadcrumb missing H2: %q", heading)
	}
	if !strings.Contains(heading, "### H3") {
		t.Errorf("breadcrumb missing H3: %q", heading)
	}
}

func TestMarkdownParser_HeadingResetsDepth(t *testing.T) {
	input := "# H1\n## H2\nBody\n# New H1\nNew body"
	units := collect(t, &MarkdownParser{}, input)
	if len(units) != 2 {
		t.Fatalf("want 2 units, got %d", len(units))
	}
	heading := units[1].Metadata["heading"]
	if strings.Contains(heading, "H2") {
		t.Errorf("second section breadcrumb should not contain H2: %q", heading)
	}
}

func TestMarkdownParser_HeadingWithNoBody(t *testing.T) {
	input := "# Title\n# Another"
	units := collect(t, &MarkdownParser{}, input)
	if len(units) != 0 {
		t.Fatalf("want 0 units, got %d", len(units))
	}
}

func TestMarkdownParser_NotAHeading(t *testing.T) {
	input := "#notaheading\ntext"
	units := collect(t, &MarkdownParser{}, input)
	if len(units) != 1 {
		t.Fatalf("want 1 unit, got %d", len(units))
	}
}
