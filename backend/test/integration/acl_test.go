//go:build integration

// Exercises the API-key collection-grant model end to end against the real
// public (auth-required) surface — see pkg/meta/apikey.go for the grant
// model and pkg/api/middleware/collection_access.go for enforcement. Unit
// coverage for the pure allow/deny evaluation lives in
// pkg/api/middleware/collection_access_test.go and
// pkg/meta/sqlite/apikeys_test.go; this file covers the parts only
// observable through real HTTP routing — cross-collection isolation, the
// read/write split, wildcard grants against collections created after the
// key, the admin surface's ACL bypass, the non-slug'd document/job routes
// where the collection isn't known until the handler resolves it, and that
// collection grants (CollectionGrant) never imply admin capability
// (AdminGrant) — see admin_acl_test.go for admin-grant-specific coverage.
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

// doAuth sends a request with an Authorization header against a real app
// (helpers.DoJSON has no auth-header support, since every other suite runs
// against the no-auth admin surface).
func doAuth(t *testing.T, app *fiber.App, method, path, key string) *http.Response {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
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

func collectionID(t *testing.T, ms meta.MetaStore, slug string) string {
	t.Helper()
	col, err := ms.GetCollectionBySlug(context.Background(), slug)
	if err != nil {
		t.Fatalf("get collection %q: %v", slug, err)
	}
	return col.ID
}

func TestACL_KeyScopedToOneCollectionCannotReachAnother(t *testing.T) {
	a, ms := helpers.NewFullTestApp(t)
	slugA := helpers.CreateCollection(t, a.Admin, "ACL Tenant A")
	slugB := helpers.CreateCollection(t, a.Admin, "ACL Tenant B")

	key := helpers.CreateAPIKey(t, ms, "tenant-a-key",
		helpers.CollectionGrant(collectionID(t, ms, slugA), meta.PermissionBoth))

	if resp := doAuth(t, a.Public, http.MethodGet, "/api/v1/collections/"+slugA+"/documents", key); resp.StatusCode != fiber.StatusOK {
		t.Errorf("own collection: expected 200, got %d", resp.StatusCode)
	}
	if resp := doAuth(t, a.Public, http.MethodGet, "/api/v1/collections/"+slugB+"/documents", key); resp.StatusCode != fiber.StatusForbidden {
		t.Errorf("other tenant's collection: expected 403, got %d", resp.StatusCode)
	}
}

func TestACL_ReadGrantCannotIngest(t *testing.T) {
	a, ms := helpers.NewFullTestApp(t)
	slug := helpers.CreateCollection(t, a.Admin, "Read Only Tenant")

	key := helpers.CreateAPIKey(t, ms, "read-only-key",
		helpers.CollectionGrant(collectionID(t, ms, slug), meta.PermissionRead))

	if resp := doAuth(t, a.Public, http.MethodGet, "/api/v1/collections/"+slug+"/documents", key); resp.StatusCode != fiber.StatusOK {
		t.Errorf("read: expected 200, got %d", resp.StatusCode)
	}
	if resp := doAuth(t, a.Public, http.MethodPost, "/api/v1/collections/"+slug+"/ingest", key); resp.StatusCode != fiber.StatusForbidden {
		t.Errorf("ingest with read-only grant: expected 403, got %d", resp.StatusCode)
	}
}

func TestACL_WildcardGrantCoversCollectionsCreatedAfterTheKey(t *testing.T) {
	a, ms := helpers.NewFullTestApp(t)

	key := helpers.CreateAPIKey(t, ms, "wildcard-key", helpers.WildcardGrant(meta.PermissionRead))

	// Collection didn't exist when the key/grant was created.
	slug := helpers.CreateCollection(t, a.Admin, "Created After Key")

	if resp := doAuth(t, a.Public, http.MethodGet, "/api/v1/collections/"+slug+"/documents", key); resp.StatusCode != fiber.StatusOK {
		t.Errorf("wildcard grant on a collection created after the key: expected 200, got %d", resp.StatusCode)
	}
}

func TestACL_AdminSurfaceBypassesACLEntirely(t *testing.T) {
	a, _ := helpers.NewFullTestApp(t)
	slug := helpers.CreateCollection(t, a.Admin, "Admin Bypass Tenant")

	// No Authorization header at all — the admin surface has no auth
	// middleware and enforceACL=false, so this must still succeed.
	req := httptest.NewRequest(http.MethodGet, "/admin/collections/"+slug+"/documents", nil)
	resp, err := a.Admin.Test(req, -1)
	if err != nil {
		t.Fatalf("admin request: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("admin surface: expected 200 with no key at all, got %d", resp.StatusCode)
	}
}

// A collection grant — even an unrestricted wildcard one — must never imply
// admin capability. This is the specific bug the AdminGrant split exists to
// prevent: a key issued for "read access to every collection" (an
// analytics tool, a backup job) must not incidentally be able to create API
// keys or collections just because it happens to carry a wildcard
// CollectionGrant.
func TestACL_CollectionGrantsDoNotImplyAdminCapability(t *testing.T) {
	a, ms := helpers.NewFullTestApp(t)
	slug := helpers.CreateCollection(t, a.Admin, "Scoped Tenant")

	scoped := helpers.CreateAPIKey(t, ms, "scoped-key",
		helpers.CollectionGrant(collectionID(t, ms, slug), meta.PermissionBoth))
	wildcardContent := helpers.CreateAPIKey(t, ms, "wildcard-content-key", helpers.WildcardGrant(meta.PermissionBoth))

	for _, key := range []string{scoped, wildcardContent} {
		if resp := doAuth(t, a.Public, http.MethodPost, "/api/v1/collections", key); resp.StatusCode != fiber.StatusForbidden {
			t.Errorf("key with no admin grant creating a collection: expected 403, got %d", resp.StatusCode)
		}
		if resp := doAuth(t, a.Public, http.MethodPost, "/api/v1/keys", key); resp.StatusCode != fiber.StatusForbidden {
			t.Errorf("key with no admin grant creating another key: expected 403, got %d", resp.StatusCode)
		}
		if resp := doAuth(t, a.Public, http.MethodGet, "/api/v1/collections", key); resp.StatusCode != fiber.StatusForbidden {
			t.Errorf("key with no admin grant listing collections: expected 403, got %d", resp.StatusCode)
		}
	}
}

func TestACL_TopLevelDocumentRouteChecksTheDocumentsOwnCollection(t *testing.T) {
	a, ms := helpers.NewFullTestApp(t)
	slugA := helpers.CreateCollection(t, a.Admin, "Doc Owner Tenant")
	slugB := helpers.CreateCollection(t, a.Admin, "Other Tenant")

	helpers.UploadDoc(t, a.Admin, slugA, "note.txt", "hello world", nil)
	doc := helpers.WaitForDocument(t, a.Admin, slugA, "note.txt")
	docID, _ := doc["id"].(string)
	if docID == "" {
		t.Fatalf("uploaded document has no id: %v", doc)
	}

	ownerKey := helpers.CreateAPIKey(t, ms, "owner-key",
		helpers.CollectionGrant(collectionID(t, ms, slugA), meta.PermissionRead))
	otherKey := helpers.CreateAPIKey(t, ms, "other-key",
		helpers.CollectionGrant(collectionID(t, ms, slugB), meta.PermissionRead))

	// Top-level route (/api/v1/documents/:id, no :slug in the URL) — the
	// collection is only known after the handler resolves doc.CollectionID.
	if resp := doAuth(t, a.Public, http.MethodGet, "/api/v1/documents/"+docID, ownerKey); resp.StatusCode != fiber.StatusOK {
		t.Errorf("owner key via top-level route: expected 200, got %d", resp.StatusCode)
	}
	if resp := doAuth(t, a.Public, http.MethodGet, "/api/v1/documents/"+docID, otherKey); resp.StatusCode != fiber.StatusForbidden {
		t.Errorf("other tenant's key via top-level route: expected 403, got %d", resp.StatusCode)
	}
}
