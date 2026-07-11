//go:build integration

package integration_test

import (
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"

	"ragpack/test/integration/helpers"
)

func TestCreateCollection(t *testing.T) {
	app := helpers.NewTestApp(t)

	resp, created := helpers.DoJSON(t, app, http.MethodPost, "/admin/collections", map[string]any{"name": "CRUD Test"})
	if resp.StatusCode != fiber.StatusOK && resp.StatusCode != fiber.StatusCreated {
		t.Fatalf("create collection: expected 200/201, got %d: %v", resp.StatusCode, created)
	}
	if created["slug"] == "" || created["slug"] == nil {
		t.Fatalf("create collection: no slug in response: %v", created)
	}
	if created["name"] != "CRUD Test" {
		t.Errorf("create collection: expected name %q, got %v", "CRUD Test", created["name"])
	}
}

func TestGetCollectionBySlug(t *testing.T) {
	app := helpers.NewTestApp(t)
	slug := helpers.CreateCollection(t, app, "Get Test")

	resp, got := helpers.DoJSON(t, app, http.MethodGet, "/admin/collections/"+slug, nil)
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("get by slug: got %d", resp.StatusCode)
	}
	if got["name"] != "Get Test" {
		t.Errorf("get by slug: expected name %q, got %v", "Get Test", got["name"])
	}
}

func TestListCollections(t *testing.T) {
	app := helpers.NewTestApp(t)
	helpers.CreateCollection(t, app, "List Test A")
	helpers.CreateCollection(t, app, "List Test B")

	resp, list := helpers.DoJSON(t, app, http.MethodGet, "/admin/collections", nil)
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("list collections: got %d", resp.StatusCode)
	}
	cols, _ := list["collections"].([]any)
	if len(cols) != 2 {
		t.Errorf("list collections: expected 2, got %d", len(cols))
	}
}

func TestDeleteCollection(t *testing.T) {
	app := helpers.NewTestApp(t)
	slug := helpers.CreateCollection(t, app, "Delete Test")

	resp, _ := helpers.DoJSON(t, app, http.MethodDelete, "/admin/collections/"+slug, nil)
	if resp.StatusCode != fiber.StatusOK && resp.StatusCode != fiber.StatusNoContent {
		t.Fatalf("delete collection: got %d", resp.StatusCode)
	}

	resp, _ = helpers.DoJSON(t, app, http.MethodGet, "/admin/collections/"+slug, nil)
	if resp.StatusCode != fiber.StatusNotFound {
		t.Errorf("get after delete: expected 404, got %d", resp.StatusCode)
	}
}
