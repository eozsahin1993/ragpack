//go:build integration

package integration_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"

	"ragpack/test/integration/helpers"
)

func TestHybridQueryRanksRelevantDocumentFirst(t *testing.T) {
	app := helpers.NewTestApp(t)
	slug := helpers.CreateCollection(t, app, "Hybrid Query Test")

	helpers.UploadDoc(t, app, slug, "platner.txt", "Graham Platner is a candidate whose campaign gained attention.", nil)
	helpers.UploadDoc(t, app, slug, "unrelated.txt", "The weather in Portland was cloudy with occasional rain this week.", nil)
	helpers.WaitForDocument(t, app, slug, "platner.txt")
	helpers.WaitForDocument(t, app, slug, "unrelated.txt")

	resp, results := helpers.DoJSON(t, app, http.MethodPost, "/admin/collections/"+slug+"/query", map[string]any{
		"query": "Platner campaign", "top_k": 5,
	})
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("query: got %d: %v", resp.StatusCode, results)
	}
	items, _ := results["results"].([]any)
	if len(items) == 0 {
		t.Fatalf("query: expected results, got none")
	}
	top, _ := items[0].(map[string]any)
	if !strings.Contains(fmt.Sprint(top["source"]), "platner") {
		t.Errorf("expected platner.txt to rank first, got %v", top["source"])
	}
	if _, ok := top["rrf_score"]; !ok {
		t.Errorf("expected rrf_score on a hybrid result, got %v", top)
	}
}

func TestVectorSearchOnlyOmitsHybridFields(t *testing.T) {
	app := helpers.NewTestApp(t)
	slug := helpers.CreateCollection(t, app, "Vector Only Test")

	helpers.UploadDoc(t, app, slug, "platner.txt", "Graham Platner is a candidate whose campaign gained attention.", nil)
	helpers.WaitForDocument(t, app, slug, "platner.txt")

	resp, results := helpers.DoJSON(t, app, http.MethodPost, "/admin/collections/"+slug+"/query", map[string]any{
		"query": "Platner campaign", "top_k": 5, "vector_search_only": true,
	})
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("vector-only query: got %d", resp.StatusCode)
	}
	items, _ := results["results"].([]any)
	if len(items) == 0 {
		t.Fatalf("vector-only query: expected results")
	}
	top, _ := items[0].(map[string]any)
	if top["rrf_score"] != nil {
		t.Errorf("expected no rrf_score on a vector-only result, got %v", top["rrf_score"])
	}
}
