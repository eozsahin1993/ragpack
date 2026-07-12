//go:build integration

// Exercises the AdminGrant model (pkg/meta/apikey.go, resource_type
// 'keys'/'prompts'/'collections'/'*') end to end — that each resource type
// is independently grantable, the "*" wildcard covers all of them, and the
// read/write split is enforced. Complements acl_test.go, which covers that
// CollectionGrant never implies admin capability.
package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"

	"ragpack/pkg/meta"
	"ragpack/test/integration/helpers"
)

func doAuthJSON(t *testing.T, app *fiber.App, method, path, key string, body any) *http.Response {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if key != "" {
		req.Header.Set("Authorization", "Bearer "+key)
	}
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	return resp
}

func TestAdminACL_ResourceTypesAreIndependentlyGranted(t *testing.T) {
	a, ms := helpers.NewFullTestApp(t)

	// A key with write access to prompts only — must be able to manage
	// prompts, but still locked out of keys/collections management, proving
	// the three resource types don't leak into each other.
	promptsOnly := helpers.CreateAdminAPIKey(t, ms, "prompts-admin",
		[]meta.AdminGrantInput{helpers.AdminGrant(meta.ResourcePrompts, meta.PermissionWrite)},
		helpers.WildcardGrant(meta.PermissionRead))

	createPrompt := map[string]any{"name": "greeting", "content": "Hello, {{query}}"}
	if resp := doAuthJSON(t, a.Public, http.MethodPost, "/api/v1/prompts", promptsOnly, createPrompt); resp.StatusCode != fiber.StatusCreated {
		t.Errorf("prompts-write key creating a prompt: expected 201, got %d", resp.StatusCode)
	}
	if resp := doAuth(t, a.Public, http.MethodPost, "/api/v1/keys", promptsOnly); resp.StatusCode != fiber.StatusForbidden {
		t.Errorf("prompts-write key creating another key: expected 403 (no keys grant), got %d", resp.StatusCode)
	}
	if resp := doAuth(t, a.Public, http.MethodPost, "/api/v1/collections", promptsOnly); resp.StatusCode != fiber.StatusForbidden {
		t.Errorf("prompts-write key creating a collection: expected 403 (no collections grant), got %d", resp.StatusCode)
	}
}

func TestAdminACL_ReadDoesNotCoverWrite(t *testing.T) {
	a, ms := helpers.NewFullTestApp(t)

	readOnly := helpers.CreateAdminAPIKey(t, ms, "collections-reader",
		[]meta.AdminGrantInput{helpers.AdminGrant(meta.ResourceCollections, meta.PermissionRead)},
		helpers.WildcardGrant(meta.PermissionRead))

	if resp := doAuth(t, a.Public, http.MethodGet, "/api/v1/collections", readOnly); resp.StatusCode != fiber.StatusOK {
		t.Errorf("collections-read key listing collections: expected 200, got %d", resp.StatusCode)
	}
	if resp := doAuthJSON(t, a.Public, http.MethodPost, "/api/v1/collections", readOnly, map[string]any{"name": "nope"}); resp.StatusCode != fiber.StatusForbidden {
		t.Errorf("collections-read key creating a collection: expected 403, got %d", resp.StatusCode)
	}
}

func TestAdminACL_WildcardResourceTypeCoversEverything(t *testing.T) {
	a, ms := helpers.NewFullTestApp(t)

	root := helpers.CreateAdminAPIKey(t, ms, "root-admin",
		[]meta.AdminGrantInput{helpers.AdminGrant(meta.ResourceAll, meta.PermissionBoth)},
		helpers.WildcardGrant(meta.PermissionBoth))

	if resp := doAuthJSON(t, a.Public, http.MethodPost, "/api/v1/collections", root, map[string]any{"name": "root-created"}); resp.StatusCode != fiber.StatusCreated {
		t.Errorf("* admin key creating a collection: expected 201, got %d", resp.StatusCode)
	}
	if resp := doAuthJSON(t, a.Public, http.MethodPost, "/api/v1/prompts", root, map[string]any{"name": "p", "content": "c"}); resp.StatusCode != fiber.StatusCreated {
		t.Errorf("* admin key creating a prompt: expected 201, got %d", resp.StatusCode)
	}
	createKeyBody := map[string]any{"name": "child-key", "grants": []map[string]any{{"permission": "read"}}}
	if resp := doAuthJSON(t, a.Public, http.MethodPost, "/api/v1/keys", root, createKeyBody); resp.StatusCode != fiber.StatusCreated {
		t.Errorf("* admin key creating another key: expected 201, got %d", resp.StatusCode)
	}
}
