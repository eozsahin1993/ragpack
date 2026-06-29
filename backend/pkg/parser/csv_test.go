package parser

import (
	"strings"
	"testing"
)

func TestCSVParser_Rows(t *testing.T) {
	input := "Name,Age,City\nAlice,30,London\nBob,25,Berlin"
	units := collect(t, &CSVParser{}, input)
	if len(units) != 2 {
		t.Fatalf("want 2 units, got %d", len(units))
	}
	if units[0].Metadata["headers"] != "Name\tAge\tCity" {
		t.Errorf("want headers %q, got %q", "Name\tAge\tCity", units[0].Metadata["headers"])
	}
	if !strings.HasPrefix(units[0].Text, "Alice") {
		t.Errorf("first row unexpected text: %q", units[0].Text)
	}
	if units[0].Kind != UnitKindRow {
		t.Errorf("want kind %q, got %q", UnitKindRow, units[0].Kind)
	}
}

func TestCSVParser_HeaderOnly(t *testing.T) {
	input := "Name,Age,City"
	units := collect(t, &CSVParser{}, input)
	if len(units) != 0 {
		t.Fatalf("want 0 units (header only), got %d", len(units))
	}
}

func TestCSVParser_Empty(t *testing.T) {
	units := collect(t, &CSVParser{}, "")
	if len(units) != 0 {
		t.Fatalf("want 0 units, got %d", len(units))
	}
}

func TestCSVParser_QuotedFields(t *testing.T) {
	input := "Name,Bio\nAlice,\"Software engineer, Go enthusiast\"\nBob,\"Writer, editor\""
	units := collect(t, &CSVParser{}, input)
	if len(units) != 2 {
		t.Fatalf("want 2 units, got %d", len(units))
	}
	if !strings.Contains(units[0].Text, "Software engineer") {
		t.Errorf("quoted field missing from unit text: %q", units[0].Text)
	}
}
