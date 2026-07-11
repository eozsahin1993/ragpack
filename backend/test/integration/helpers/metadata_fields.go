package helpers

import (
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// RegisterMetadataField registers a single metadata field on a collection.
func RegisterMetadataField(t *testing.T, app *fiber.App, slug, name, fieldType string) {
	t.Helper()
	resp, _ := DoJSON(t, app, http.MethodPost, "/admin/collections/"+slug+"/metadata-fields", map[string]any{
		"fields": []map[string]string{{"name": name, "type": fieldType}},
	})
	if resp.StatusCode != fiber.StatusOK && resp.StatusCode != fiber.StatusCreated {
		t.Fatalf("register metadata field %q: got %d", name, resp.StatusCode)
	}
}
