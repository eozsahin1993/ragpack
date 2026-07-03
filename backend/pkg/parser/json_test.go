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

	want := []string{
		strings.TrimSpace(`
department: Engineering
email: alice@example.com
id: 1
joined: 2021-03-15
location: London
name: Alice Johnson
role: Software Engineer`),
		strings.TrimSpace(`
department: Product
email: bob@example.com
id: 2
joined: 2020-07-01
location: Berlin
name: Bob Smith
role: Product Manager`),
		strings.TrimSpace(`
department: Engineering
email: carol@example.com
id: 3
joined: 2022-01-10
location: Amsterdam
name: Carol White
role: Data Scientist`),
	}

	for i, unit := range units {
		if unit.Kind != UnitKindRow {
			t.Errorf("unit %d: want kind %q, got %q", i, UnitKindRow, unit.Kind)
		}
		if unit.Text != want[i] {
			t.Errorf("unit %d:\nwant:\n%s\n\ngot:\n%s", i, want[i], unit.Text)
		}
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

	want := []string{
		strings.TrimSpace(`
address.city: London
address.country: UK
address.street: 12 Baker Street
id: 1
meta.active: true
meta.score: 98.5
name: Alice Johnson
role: Software Engineer
skills.0: Go
skills.1: Python
skills.2: Kubernetes`),
		strings.TrimSpace(`
address.city: Berlin
address.country: Germany
address.street: 45 Unter den Linden
id: 2
meta.active: true
meta.score: 91
name: Bob Smith
role: Product Manager
skills.0: Roadmapping
skills.1: SQL
skills.2: Figma`),
	}

	for i, unit := range units {
		if unit.Kind != UnitKindRow {
			t.Errorf("unit %d: want kind %q, got %q", i, UnitKindRow, unit.Kind)
		}
		if unit.Text != want[i] {
			t.Errorf("unit %d:\nwant:\n%s\n\ngot:\n%s", i, want[i], unit.Text)
		}
	}
}
