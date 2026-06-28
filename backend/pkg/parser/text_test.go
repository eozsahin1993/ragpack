package parser

import "testing"

func TestTextParser_SingleParagraph(t *testing.T) {
	units := collect(t, &TextParser{}, "Hello world")
	if len(units) != 1 {
		t.Fatalf("want 1 unit, got %d", len(units))
	}
	if units[0].Text != "Hello world" {
		t.Errorf("want %q, got %q", "Hello world", units[0].Text)
	}
}

func TestTextParser_MultipleParagraphs(t *testing.T) {
	input := "First paragraph\n\nSecond paragraph\n\nThird paragraph"
	units := collect(t, &TextParser{}, input)
	if len(units) != 3 {
		t.Fatalf("want 3 units, got %d", len(units))
	}
	if units[0].Text != "First paragraph" {
		t.Errorf("want %q, got %q", "First paragraph", units[0].Text)
	}
	if units[2].Text != "Third paragraph" {
		t.Errorf("want %q, got %q", "Third paragraph", units[2].Text)
	}
}

func TestTextParser_MultipleBlankLines(t *testing.T) {
	input := "First\n\n\n\nSecond"
	units := collect(t, &TextParser{}, input)
	if len(units) != 2 {
		t.Fatalf("want 2 units, got %d: %v", len(units), units)
	}
}

func TestTextParser_Empty(t *testing.T) {
	units := collect(t, &TextParser{}, "")
	if len(units) != 0 {
		t.Fatalf("want 0 units, got %d", len(units))
	}
}

func TestTextParser_WhitespaceOnly(t *testing.T) {
	units := collect(t, &TextParser{}, "   \n\n   \n")
	if len(units) != 0 {
		t.Fatalf("want 0 units, got %d", len(units))
	}
}

func TestTextParser_TrimsWhitespace(t *testing.T) {
	units := collect(t, &TextParser{}, "  hello  \n\n  world  ")
	if len(units) != 2 {
		t.Fatalf("want 2 units, got %d", len(units))
	}
	if units[0].Text != "hello" {
		t.Errorf("want %q, got %q", "hello", units[0].Text)
	}
}
