package parser

import (
	"os"
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

// testdata/sample.md is the ragpack documentation used as a realistic fixture:
//
//	# Ragpack              → unit 0:  "# Ragpack"
//	## Architecture        → unit 1:  "# Ragpack > ## Architecture"
//	## Data Model          → (no direct body, skipped)
//	### Collections        → unit 2:  "# Ragpack > ## Data Model > ### Collections"
//	### Documents          → unit 3:  "# Ragpack > ## Data Model > ### Documents"
//	### Chunks             → unit 4:  "# Ragpack > ## Data Model > ### Chunks"
//	## Ingestion Pipeline  → (no direct body, skipped)
//	### Parsing            → unit 5:  "# Ragpack > ## Ingestion Pipeline > ### Parsing"
//	### Chunking           → unit 6:  "# Ragpack > ## Ingestion Pipeline > ### Chunking"
//	### Embedding          → unit 7:  "# Ragpack > ## Ingestion Pipeline > ### Embedding"
//	## Retrieval and RAG   → unit 8:  "# Ragpack > ## Retrieval and RAG"
//	## Deployment          → (no direct body, skipped)
//	### Requirements       → unit 9:  "# Ragpack > ## Deployment > ### Requirements"
//	### Configuration      → unit 10: "# Ragpack > ## Deployment > ### Configuration"
//	### Upgrading          → unit 11: "# Ragpack > ## Deployment > ### Upgrading"
func loadSampleMarkdown(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile("testdata/sample.md")
	if err != nil {
		t.Fatalf("read testdata/sample.md: %v", err)
	}
	return strings.TrimSpace(string(data))
}

func TestMarkdownParser_ComplexDocument_UnitCount(t *testing.T) {
	units := collect(t, &MarkdownParser{}, loadSampleMarkdown(t))
	if len(units) != 12 {
		t.Fatalf("want 12 units, got %d", len(units))
	}
}

func TestMarkdownParser_ComplexDocument_TopLevelUnit(t *testing.T) {
	units := collect(t, &MarkdownParser{}, loadSampleMarkdown(t))
	if units[0].Metadata["heading"] != "# Ragpack" {
		t.Errorf("first unit heading: want %q, got %q", "# Ragpack", units[0].Metadata["heading"])
	}
	if !strings.HasPrefix(units[0].Text, "Ragpack is a self-hostable") {
		t.Errorf("first unit unexpected text start: %q", units[0].Text)
	}
}

func TestMarkdownParser_ComplexDocument_BreadcrumbInheritance(t *testing.T) {
	units := collect(t, &MarkdownParser{}, loadSampleMarkdown(t))

	// unit 2: Collections is under Data Model which is under Ragpack
	collectionsHeading := units[2].Metadata["heading"]
	if !strings.Contains(collectionsHeading, "# Ragpack") {
		t.Errorf("Collections unit missing Ragpack in breadcrumb: %q", collectionsHeading)
	}
	if !strings.Contains(collectionsHeading, "## Data Model") {
		t.Errorf("Collections unit missing Data Model in breadcrumb: %q", collectionsHeading)
	}
	if !strings.Contains(collectionsHeading, "### Collections") {
		t.Errorf("Collections unit missing Collections in breadcrumb: %q", collectionsHeading)
	}

	// unit 6: Chunking is under Ingestion Pipeline which is under Ragpack
	chunkingHeading := units[6].Metadata["heading"]
	if !strings.Contains(chunkingHeading, "# Ragpack") {
		t.Errorf("Chunking unit missing Ragpack in breadcrumb: %q", chunkingHeading)
	}
	if !strings.Contains(chunkingHeading, "## Ingestion Pipeline") {
		t.Errorf("Chunking unit missing Ingestion Pipeline in breadcrumb: %q", chunkingHeading)
	}
	if !strings.Contains(chunkingHeading, "### Chunking") {
		t.Errorf("Chunking unit missing Chunking in breadcrumb: %q", chunkingHeading)
	}
}

func TestMarkdownParser_ComplexDocument_SiblingH2sShareH1(t *testing.T) {
	units := collect(t, &MarkdownParser{}, loadSampleMarkdown(t))

	// unit 9 (Requirements) and unit 10 (Configuration) are both under Deployment
	// and must not carry breadcrumbs from Ingestion Pipeline
	for _, idx := range []int{9, 10, 11} {
		heading := units[idx].Metadata["heading"]
		if strings.Contains(heading, "Ingestion Pipeline") {
			t.Errorf("unit %d breadcrumb should not contain Ingestion Pipeline: %q", idx, heading)
		}
		if !strings.Contains(heading, "## Deployment") {
			t.Errorf("unit %d breadcrumb missing Deployment: %q", idx, heading)
		}
	}
}

func TestMarkdownParser_ComplexDocument_BodyContent(t *testing.T) {
	units := collect(t, &MarkdownParser{}, loadSampleMarkdown(t))

	// unit 2: Collections
	if !strings.HasPrefix(units[2].Text, "A collection is a named group") {
		t.Errorf("Collections unit unexpected text start: %q", units[2].Text)
	}
	// unit 5: Parsing
	if !strings.HasPrefix(units[5].Text, "The parser converts raw file bytes") {
		t.Errorf("Parsing unit unexpected text start: %q", units[5].Text)
	}
	// unit 11: Upgrading
	if !strings.HasPrefix(units[11].Text, "The CLI reads its own version") {
		t.Errorf("Upgrading unit unexpected text start: %q", units[11].Text)
	}
}
