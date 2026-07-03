package parser

import (
	"os"
	"strings"
	"testing"
)

func TestXMLParser_BasicRecords(t *testing.T) {
	input := `<items><item><name>Alice</name><role>Engineer</role></item><item><name>Bob</name><role>Manager</role></item></items>`
	units := collect(t, &XMLParser{}, input)
	if len(units) != 2 {
		t.Fatalf("want 2 units, got %d", len(units))
	}
	if units[0].Kind != UnitKindRow {
		t.Errorf("want kind %q, got %q", UnitKindRow, units[0].Kind)
	}
	if units[0].Metadata["element"] != "item" {
		t.Errorf("want element=item, got %q", units[0].Metadata["element"])
	}
	if !strings.Contains(units[0].Text, "name: Alice") {
		t.Errorf("first unit missing name: %q", units[0].Text)
	}
	if !strings.Contains(units[1].Text, "name: Bob") {
		t.Errorf("second unit missing name: %q", units[1].Text)
	}
}

func TestXMLParser_NestedElements(t *testing.T) {
	input := `<root><record><name>Alice</name><address><city>London</city><country>UK</country></address></record></root>`
	units := collect(t, &XMLParser{}, input)
	if len(units) != 1 {
		t.Fatalf("want 1 unit, got %d", len(units))
	}
	if !strings.Contains(units[0].Text, "address.city: London") {
		t.Errorf("unit missing nested city: %q", units[0].Text)
	}
	if !strings.Contains(units[0].Text, "address.country: UK") {
		t.Errorf("unit missing nested country: %q", units[0].Text)
	}
}

func TestXMLParser_Empty(t *testing.T) {
	units := collect(t, &XMLParser{}, `<root></root>`)
	if len(units) != 0 {
		t.Fatalf("want 0 units, got %d", len(units))
	}
}

func TestXMLParser_Fixture(t *testing.T) {
	file, err := os.Open("testdata/sample.xml")
	if err != nil {
		t.Fatalf("open sample.xml: %v", err)
	}
	units := collectBytes(t, &XMLParser{}, mustRead(t, file))

	if len(units) != 3 {
		t.Fatalf("want 3 units, got %d", len(units))
	}

	want := []string{
		strings.TrimSpace(`
id: 1
name: Alice Johnson
role: Software Engineer
department: Engineering
address.street: 12 Baker Street
address.city: London
address.country: UK
skills.skill: Go
skills.skill: Python
skills.skill: Kubernetes`),
		strings.TrimSpace(`
id: 2
name: Bob Smith
role: Product Manager
department: Product
address.street: 45 Unter den Linden
address.city: Berlin
address.country: Germany
skills.skill: Roadmapping
skills.skill: SQL
skills.skill: Figma`),
		strings.TrimSpace(`
id: 3
name: Carol White
role: Data Scientist
department: Engineering
address.street: 8 Prinsengracht
address.city: Amsterdam
address.country: Netherlands
skills.skill: Python
skills.skill: SQL
skills.skill: PyTorch`),
	}

	for i, unit := range units {
		if unit.Metadata["element"] != "employee" {
			t.Errorf("unit %d: want element=employee, got %q", i, unit.Metadata["element"])
		}
		if unit.Text != want[i] {
			t.Errorf("unit %d:\nwant:\n%s\n\ngot:\n%s", i, want[i], unit.Text)
		}
	}
}
