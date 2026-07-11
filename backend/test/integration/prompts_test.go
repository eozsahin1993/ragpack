//go:build integration

package integration_test

import (
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"

	"ragpack/test/integration/helpers"
)

func TestCreatePrompt(t *testing.T) {
	app := helpers.NewTestApp(t)

	resp, created := helpers.DoJSON(t, app, http.MethodPost, "/admin/prompts", map[string]any{
		"name": "My Prompt", "content": "Answer using {{context}} for {{question}}",
	})
	if resp.StatusCode != fiber.StatusCreated {
		t.Fatalf("create prompt: expected 201, got %d: %v", resp.StatusCode, created)
	}
	if created["slug"] == "" || created["slug"] == nil {
		t.Fatalf("create prompt: no slug in response: %v", created)
	}
	if created["is_system"] != false {
		t.Errorf("expected is_system false for a user-created prompt, got %v", created["is_system"])
	}
}

func TestGetPromptBySlug(t *testing.T) {
	app := helpers.NewTestApp(t)
	slug := helpers.CreatePrompt(t, app, "Get Prompt Test", "content here")

	resp, got := helpers.DoJSON(t, app, http.MethodGet, "/admin/prompts/"+slug, nil)
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("get prompt: got %d", resp.StatusCode)
	}
	if got["name"] != "Get Prompt Test" {
		t.Errorf("expected name %q, got %v", "Get Prompt Test", got["name"])
	}
}

func TestListPromptsIncludesSeededSystemPrompts(t *testing.T) {
	app := helpers.NewTestApp(t)

	resp, list := helpers.DoJSON(t, app, http.MethodGet, "/admin/prompts", nil)
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("list prompts: got %d", resp.StatusCode)
	}
	system, _ := list["system"].([]any)
	if len(system) == 0 {
		t.Errorf("expected seeded system prompts (e.g. basic_rag), got none")
	}
}

func TestUpdatePrompt(t *testing.T) {
	app := helpers.NewTestApp(t)
	slug := helpers.CreatePrompt(t, app, "Update Prompt Test", "original content")

	resp, updated := helpers.DoJSON(t, app, http.MethodPatch, "/admin/prompts/"+slug, map[string]any{
		"content": "updated content",
	})
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("update prompt: got %d", resp.StatusCode)
	}
	if updated["content"] != "updated content" {
		t.Errorf("expected updated content, got %v", updated["content"])
	}
}

func TestDeletePrompt(t *testing.T) {
	app := helpers.NewTestApp(t)
	slug := helpers.CreatePrompt(t, app, "Delete Prompt Test", "content")

	resp, _ := helpers.DoJSON(t, app, http.MethodDelete, "/admin/prompts/"+slug, nil)
	if resp.StatusCode != fiber.StatusOK && resp.StatusCode != fiber.StatusNoContent {
		t.Fatalf("delete prompt: got %d", resp.StatusCode)
	}

	resp, _ = helpers.DoJSON(t, app, http.MethodGet, "/admin/prompts/"+slug, nil)
	if resp.StatusCode != fiber.StatusNotFound {
		t.Errorf("get after delete: expected 404, got %d", resp.StatusCode)
	}
}

// TestSystemPromptIsReadOnly covers meta.ErrSystemReadOnly (pkg/api/prompts/handler.go):
// seeded system prompts (see pkg/meta/sqlite/seeds.go) can't be updated or deleted.
func TestSystemPromptIsReadOnly(t *testing.T) {
	app := helpers.NewTestApp(t)
	const systemSlug = "basic_rag"

	resp, _ := helpers.DoJSON(t, app, http.MethodPatch, "/admin/prompts/"+systemSlug, map[string]any{"content": "hacked"})
	if resp.StatusCode != fiber.StatusForbidden {
		t.Errorf("update system prompt: expected 403, got %d", resp.StatusCode)
	}

	resp, _ = helpers.DoJSON(t, app, http.MethodDelete, "/admin/prompts/"+systemSlug, nil)
	if resp.StatusCode != fiber.StatusForbidden {
		t.Errorf("delete system prompt: expected 403, got %d", resp.StatusCode)
	}
}
