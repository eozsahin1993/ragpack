package parser

import (
	"strings"
	"testing"
)

func TestHTMLParser_SimpleText(t *testing.T) {
	input := `<html><body><p>Hello world</p></body></html>`
	units := collect(t, &HTMLParser{}, input)
	if len(units) == 0 {
		t.Fatal("want at least 1 unit, got 0")
	}
	if !strings.Contains(units[0].Text, "Hello world") {
		t.Errorf("expected text to contain 'Hello world', got %q", units[0].Text)
	}
}

func TestHTMLParser_HeadingsBecomeSections(t *testing.T) {
	input := `<html><body><h1>Title</h1><p>Body</p></body></html>`
	units := collect(t, &HTMLParser{}, input)
	if len(units) == 0 {
		t.Fatal("want at least 1 unit")
	}
	if units[0].Metadata["heading"] == "" {
		t.Error("expected heading metadata to be set")
	}
}

func TestHTMLParser_SkipsScript(t *testing.T) {
	input := `<html><body><script>alert('xss')</script><p>Safe content</p></body></html>`
	units := collect(t, &HTMLParser{}, input)
	for _, unit := range units {
		if strings.Contains(unit.Text, "alert") {
			t.Errorf("script content leaked into unit: %q", unit.Text)
		}
	}
}

func TestHTMLParser_SkipsNav(t *testing.T) {
	input := `<html><body><nav>Skip me</nav><p>Keep me</p></body></html>`
	units := collect(t, &HTMLParser{}, input)
	for _, unit := range units {
		if strings.Contains(unit.Text, "Skip me") {
			t.Errorf("nav content leaked into unit: %q", unit.Text)
		}
	}
}

func TestHTMLParser_MultipleParagraphs(t *testing.T) {
	input := `<html><body><p>First paragraph</p><p>Second paragraph</p></body></html>`
	units := collect(t, &HTMLParser{}, input)
	if len(units) != 1 {
		t.Fatalf("want 1 unit, got %d", len(units))
	}
	if !strings.Contains(units[0].Text, "First paragraph") {
		t.Errorf("unit missing 'First paragraph': %q", units[0].Text)
	}
	if !strings.Contains(units[0].Text, "Second paragraph") {
		t.Errorf("unit missing 'Second paragraph': %q", units[0].Text)
	}
}

func TestHTMLParser_Empty(t *testing.T) {
	units := collect(t, &HTMLParser{}, `<html><body></body></html>`)
	if len(units) != 0 {
		t.Fatalf("want 0 units, got %d", len(units))
	}
}
