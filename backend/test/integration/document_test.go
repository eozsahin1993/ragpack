//go:build integration

package integration_test

import (
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"

	"ragpack/test/integration/helpers"
)

func TestPatchDocument(t *testing.T) {
	app := helpers.NewTestApp(t)
	slug := helpers.CreateCollection(t, app, "Patch Test")

	helpers.UploadDoc(t, app, slug, "doc.txt", "A short document about gardening.", nil)
	doc := helpers.WaitForDocument(t, app, slug, "doc.txt")
	docID, _ := doc["id"].(string)

	resp, patched := helpers.DoJSON(t, app, http.MethodPatch, "/admin/documents/"+docID, map[string]any{"name": "Renamed"})
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("patch document: got %d", resp.StatusCode)
	}
	if patched["name"] != "Renamed" {
		t.Errorf("expected name Renamed, got %v", patched["name"])
	}
}

func TestDeleteDocument(t *testing.T) {
	app := helpers.NewTestApp(t)
	slug := helpers.CreateCollection(t, app, "Delete Doc Test")

	helpers.UploadDoc(t, app, slug, "doc.txt", "A short document about gardening.", nil)
	doc := helpers.WaitForDocument(t, app, slug, "doc.txt")
	docID, _ := doc["id"].(string)

	resp, _ := helpers.DoJSON(t, app, http.MethodDelete, "/admin/documents/"+docID, nil)
	if resp.StatusCode != fiber.StatusOK && resp.StatusCode != fiber.StatusNoContent {
		t.Fatalf("delete document: got %d", resp.StatusCode)
	}

	resp, _ = helpers.DoJSON(t, app, http.MethodGet, "/admin/documents/"+docID, nil)
	if resp.StatusCode != fiber.StatusNotFound {
		t.Errorf("get after delete: expected 404, got %d", resp.StatusCode)
	}
}

// TestDeleteDocumentScopedToCollection covers the /collections/:slug/documents/:id
// route specifically — documents.Register is mounted both top-level (slug-less,
// covered by TestDeleteDocument) and under /collections/:slug (see
// pkg/api/router.go and CLAUDE.md's "two API surfaces" notes); both need to work.
func TestDeleteDocumentScopedToCollection(t *testing.T) {
	app := helpers.NewTestApp(t)
	slug := helpers.CreateCollection(t, app, "Scoped Delete Test")

	helpers.UploadDoc(t, app, slug, "doc.txt", "A short document about gardening.", nil)
	doc := helpers.WaitForDocument(t, app, slug, "doc.txt")
	docID, _ := doc["id"].(string)

	scopedPath := "/admin/collections/" + slug + "/documents/" + docID

	resp, got := helpers.DoJSON(t, app, http.MethodGet, scopedPath, nil)
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("get via collection-scoped route: got %d", resp.StatusCode)
	}
	if got["id"] != docID {
		t.Errorf("expected id %q, got %v", docID, got["id"])
	}

	resp, _ = helpers.DoJSON(t, app, http.MethodDelete, scopedPath, nil)
	if resp.StatusCode != fiber.StatusOK && resp.StatusCode != fiber.StatusNoContent {
		t.Fatalf("delete via collection-scoped route: got %d", resp.StatusCode)
	}

	resp, _ = helpers.DoJSON(t, app, http.MethodGet, scopedPath, nil)
	if resp.StatusCode != fiber.StatusNotFound {
		t.Errorf("get after delete: expected 404, got %d", resp.StatusCode)
	}
}
