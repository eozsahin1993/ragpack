package helpers

import (
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// CreateCollection creates a collection and returns its slug, failing the test on error.
func CreateCollection(t *testing.T, app *fiber.App, name string) string {
	t.Helper()
	resp, created := DoJSON(t, app, http.MethodPost, "/admin/collections", map[string]any{"name": name})
	if resp.StatusCode != fiber.StatusOK && resp.StatusCode != fiber.StatusCreated {
		t.Fatalf("create collection: expected 200/201, got %d: %v", resp.StatusCode, created)
	}
	slug, _ := created["slug"].(string)
	if slug == "" {
		t.Fatalf("create collection: no slug in response: %v", created)
	}
	return slug
}
