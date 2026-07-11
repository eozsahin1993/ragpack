//go:build integration

package integration_test

import (
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"

	"ragpack/test/integration/helpers"
)

func TestMetadataFilterMatchesTaggedDocument(t *testing.T) {
	app := helpers.NewTestApp(t)
	slug := helpers.CreateCollection(t, app, "Filter Match Test")
	helpers.RegisterMetadataField(t, app, slug, "category", "str")

	helpers.UploadDoc(t, app, slug, "tagged.txt", "A document about machine learning fundamentals.", map[string]string{"category": "ml"})
	doc := helpers.WaitForDocument(t, app, slug, "tagged.txt")
	if doc["status"] != "complete" {
		t.Fatalf("ingest failed: %v", doc["status"])
	}

	resp, matched := helpers.DoJSON(t, app, http.MethodPost, "/admin/collections/"+slug+"/query", map[string]any{
		"query": "machine learning", "top_k": 5,
		"filters": map[string]any{"category": map[string]string{"$eq": "ml"}},
	})
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("filtered query: got %d: %v", resp.StatusCode, matched)
	}
	if items, _ := matched["results"].([]any); len(items) == 0 {
		t.Errorf("expected a match for category=ml, got none")
	}
}

func TestMetadataFilterExcludesNonMatchingValue(t *testing.T) {
	app := helpers.NewTestApp(t)
	slug := helpers.CreateCollection(t, app, "Filter Exclude Test")
	helpers.RegisterMetadataField(t, app, slug, "category", "str")

	helpers.UploadDoc(t, app, slug, "tagged.txt", "A document about machine learning fundamentals.", map[string]string{"category": "ml"})
	doc := helpers.WaitForDocument(t, app, slug, "tagged.txt")
	if doc["status"] != "complete" {
		t.Fatalf("ingest failed: %v", doc["status"])
	}

	resp, excluded := helpers.DoJSON(t, app, http.MethodPost, "/admin/collections/"+slug+"/query", map[string]any{
		"query": "machine learning", "top_k": 5,
		"filters": map[string]any{"category": map[string]string{"$eq": "other"}},
	})
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("excluding filtered query: got %d: %v", resp.StatusCode, excluded)
	}
	if items, _ := excluded["results"].([]any); len(items) != 0 {
		t.Errorf("expected no matches, got %d", len(items))
	}
}
