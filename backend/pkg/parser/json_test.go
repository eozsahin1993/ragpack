package parser

import (
	"os"
	"strings"
	"testing"
)

func TestJSONParser_Array(t *testing.T) {
	input := `[{"name":"Alice","age":30,"city":"London"},{"name":"Bob","age":25,"city":"Berlin"}]`
	units := collect(t, &JSONParser{}, input)
	if len(units) != 2 {
		t.Fatalf("want 2 units, got %d", len(units))
	}
	if units[0].Kind != UnitKindRow {
		t.Errorf("want kind %q, got %q", UnitKindRow, units[0].Kind)
	}
	if !strings.Contains(units[0].Text, "name: Alice") {
		t.Errorf("first unit missing name: %q", units[0].Text)
	}
	if !strings.Contains(units[1].Text, "name: Bob") {
		t.Errorf("second unit missing name: %q", units[1].Text)
	}
}

func TestJSONParser_RootObject(t *testing.T) {
	input := `{"title":"Intro to Go","author":"Alice"}`
	units := collect(t, &JSONParser{}, input)
	if len(units) != 1 {
		t.Fatalf("want 1 unit, got %d", len(units))
	}
	if !strings.Contains(units[0].Text, "title: Intro to Go") {
		t.Errorf("unit missing title: %q", units[0].Text)
	}
}

func TestJSONParser_NestedObject(t *testing.T) {
	input := `[{"name":"Alice","address":{"street":"123 Main","city":"London"}}]`
	units := collect(t, &JSONParser{}, input)
	if len(units) != 1 {
		t.Fatalf("want 1 unit, got %d", len(units))
	}
	if !strings.Contains(units[0].Text, "address.street: 123 Main") {
		t.Errorf("unit missing nested field: %q", units[0].Text)
	}
	if !strings.Contains(units[0].Text, "address.city: London") {
		t.Errorf("unit missing nested city: %q", units[0].Text)
	}
}

func TestJSONParser_ArrayField(t *testing.T) {
	input := `[{"name":"Alice","tags":["go","python"]}]`
	units := collect(t, &JSONParser{}, input)
	if len(units) != 1 {
		t.Fatalf("want 1 unit, got %d", len(units))
	}
	if !strings.Contains(units[0].Text, "tags.0: go") {
		t.Errorf("unit missing array index field: %q", units[0].Text)
	}
	if !strings.Contains(units[0].Text, "tags.1: python") {
		t.Errorf("unit missing array index field: %q", units[0].Text)
	}
}

func TestJSONParser_NullValuesSkipped(t *testing.T) {
	input := `[{"name":"Alice","bio":null}]`
	units := collect(t, &JSONParser{}, input)
	if len(units) != 1 {
		t.Fatalf("want 1 unit, got %d", len(units))
	}
	if strings.Contains(units[0].Text, "bio") {
		t.Errorf("null field should be skipped: %q", units[0].Text)
	}
}

func TestJSONParser_EmptyArray(t *testing.T) {
	units := collect(t, &JSONParser{}, `[]`)
	if len(units) != 0 {
		t.Fatalf("want 0 units, got %d", len(units))
	}
}

func TestJSONParser_FlatFixture(t *testing.T) {
	file, err := os.Open("testdata/sample_flat.json")
	if err != nil {
		t.Fatalf("open sample_flat.json: %v", err)
	}
	units := collectBytes(t, &JSONParser{}, mustRead(t, file))

	if len(units) != 3 {
		t.Fatalf("want 3 units, got %d", len(units))
	}
	for _, unit := range units {
		if unit.Kind != UnitKindRow {
			t.Errorf("want kind %q, got %q", UnitKindRow, unit.Kind)
		}
	}
	if !strings.HasPrefix(units[0].Text, "id: 1") && !strings.Contains(units[0].Text, "name: Alice Johnson") {
		t.Errorf("first unit unexpected content: %q", units[0].Text)
	}
	if !strings.Contains(units[0].Text, "role: Software Engineer") {
		t.Errorf("first unit missing role: %q", units[0].Text)
	}
	if !strings.Contains(units[2].Text, "name: Carol White") {
		t.Errorf("third unit missing name: %q", units[2].Text)
	}
}

func TestJSONParser_NestedFixture(t *testing.T) {
	file, err := os.Open("testdata/sample_nested.json")
	if err != nil {
		t.Fatalf("open sample_nested.json: %v", err)
	}
	units := collectBytes(t, &JSONParser{}, mustRead(t, file))

	if len(units) != 2 {
		t.Fatalf("want 2 units, got %d", len(units))
	}

	// Nested address fields should be flattened with dot notation
	if !strings.Contains(units[0].Text, "address.street: 12 Baker Street") {
		t.Errorf("first unit missing nested street: %q", units[0].Text)
	}
	if !strings.Contains(units[0].Text, "address.city: London") {
		t.Errorf("first unit missing nested city: %q", units[0].Text)
	}
	if !strings.Contains(units[0].Text, "address.country: UK") {
		t.Errorf("first unit missing nested country: %q", units[0].Text)
	}

	// Array fields should use index notation
	if !strings.Contains(units[0].Text, "skills.0: Go") {
		t.Errorf("first unit missing skills.0: %q", units[0].Text)
	}
	if !strings.Contains(units[0].Text, "skills.1: Python") {
		t.Errorf("first unit missing skills.1: %q", units[0].Text)
	}

	// Nested meta fields
	if !strings.Contains(units[0].Text, "meta.score: 98.5") {
		t.Errorf("first unit missing meta.score: %q", units[0].Text)
	}
	if !strings.Contains(units[1].Text, "address.city: Berlin") {
		t.Errorf("second unit missing nested city: %q", units[1].Text)
	}
}
