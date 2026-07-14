//go:build integration

// The other test files all exercise the admin (no-auth) surface, since
// RegisterPublic and RegisterAdmin mount the identical route/handler set
// (see pkg/api/router.go) — a bug in route logic would show up there either
// way. The one thing admin-only tests never touch is the auth middleware
// itself (pkg/api/middleware.Auth), which only wraps the public surface —
// covered here instead.
package integration_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"

	"ragpack/pkg/meta"
	"ragpack/test/integration/helpers"
)

func TestPublicAPIRequiresAuth(t *testing.T) {
	a, _, _ := helpers.NewFullTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/collections", nil)
	resp, err := a.Public.Test(req, -1)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("expected 401 with no Authorization header, got %d", resp.StatusCode)
	}
}

func TestPublicAPIRejectsInvalidKey(t *testing.T) {
	a, _, _ := helpers.NewFullTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/collections", nil)
	req.Header.Set("Authorization", "Bearer not-a-real-key")
	resp, err := a.Public.Test(req, -1)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("expected 401 with an invalid key, got %d", resp.StatusCode)
	}
}

func TestPublicAPIAcceptsValidKey(t *testing.T) {
	a, ms, _ := helpers.NewFullTestApp(t)

	const plaintext = "test-secret-key"
	grants := []meta.GrantInput{{Permission: meta.PermissionBoth}}
	// GET /collections is gated on the "collections" admin resource (see
	// pkg/api/collections/router.go), not a collection grant, so this key
	// needs both to reach it.
	adminGrants := []meta.AdminGrantInput{{ResourceType: meta.ResourceCollections, Permission: meta.PermissionRead}}
	if _, err := ms.CreateAPIKey(context.Background(), "test key", plaintext, grants, adminGrants); err != nil {
		t.Fatalf("create api key: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/collections", nil)
	req.Header.Set("Authorization", "Bearer "+plaintext)
	resp, err := a.Public.Test(req, -1)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200 with a valid key, got %d", resp.StatusCode)
	}
}
