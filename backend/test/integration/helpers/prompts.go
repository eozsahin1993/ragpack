package helpers

import (
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// CreatePrompt creates a prompt and returns its slug, failing the test on error.
func CreatePrompt(t *testing.T, app *fiber.App, name, content string) string {
	t.Helper()
	resp, created := DoJSON(t, app, http.MethodPost, "/admin/prompts", map[string]any{
		"name": name, "content": content,
	})
	if resp.StatusCode != fiber.StatusCreated {
		t.Fatalf("create prompt: expected 201, got %d: %v", resp.StatusCode, created)
	}
	slug, _ := created["slug"].(string)
	if slug == "" {
		t.Fatalf("create prompt: no slug in response: %v", created)
	}
	return slug
}
