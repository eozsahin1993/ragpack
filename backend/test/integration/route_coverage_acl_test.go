//go:build integration

// Systematic coverage of every ACL-gated route in pkg/api/router.go — the
// gap acl_test.go and admin_acl_test.go didn't close on their own. Those
// files prove the *mechanism* is correct (cross-tenant isolation, read/write
// split, admin-grant independence) via a representative sample of routes;
// this file exists because sampling can't catch a route that was simply
// never wired to any check at all — a missing requireWrite/adminMW(...) at
// a call site compiles fine and passes every other test, since none of them
// touch that specific route. Every route below gets two requests: one with
// a key that has no relevant grant (must be 403) and one with a key that
// has exactly the right grant (must not be 403 — whatever happens after the
// ACL gate, like a 400 for a missing body field or a 404 for a made-up ID,
// is irrelevant to what this file checks).
package integration_test

import (
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"

	"ragpack/pkg/meta"
	"ragpack/test/integration/helpers"
)

func TestACL_EveryRouteEnforcesAccessControl(t *testing.T) {
	a, ms, _ := helpers.NewFullTestApp(t)

	slugA := helpers.CreateCollection(t, a.Admin, "Route Coverage A")
	slugB := helpers.CreateCollection(t, a.Admin, "Route Coverage B")

	// Shared, read-only fixture — every non-destructive case below reads or
	// renames this same document/job, which is safe regardless of test
	// order since neither operation removes it.
	job := helpers.UploadDoc(t, a.Admin, slugA, "route-cov.txt", "hello world", nil)
	doc := helpers.WaitForDocument(t, a.Admin, slugA, "route-cov.txt")
	docID, _ := doc["id"].(string)
	jobID, _ := job["id"].(string)
	if docID == "" || jobID == "" {
		t.Fatalf("fixture setup failed: docID=%q jobID=%q", docID, jobID)
	}

	// Scoped only to B, no admin grants at all — rejected by every route
	// below regardless of category, since none of them grant anything on A
	// or any admin resource type.
	otherKey := helpers.CreateAPIKey(t, ms, "route-cov-other",
		helpers.CollectionGrant(collectionID(t, ms, slugB), meta.PermissionBoth))

	// Full (both) access to A — clears every collection-scoped route below.
	collectionKey := helpers.CreateAPIKey(t, ms, "route-cov-collection",
		helpers.CollectionGrant(collectionID(t, ms, slugA), meta.PermissionBoth))

	// Wildcard collection read (covers the two top-level "list everything"
	// routes) plus "*" admin (both) (covers every admin-gated route).
	adminKey := helpers.CreateAdminAPIKey(t, ms, "route-cov-admin",
		[]meta.AdminGrantInput{helpers.AdminGrant(meta.ResourceAll, meta.PermissionBoth)},
		helpers.WildcardGrant(meta.PermissionRead))

	renameBody := map[string]any{"name": "renamed"}

	type routeCase struct {
		method   string
		path     string
		body     any
		category string // "collection" | "collection-wildcard" | "admin"
	}

	cases := map[string]routeCase{
		// top-level jobs (read-only)
		"GET /jobs":     {http.MethodGet, "/api/v1/jobs", nil, "collection-wildcard"},
		"GET /jobs/:id": {http.MethodGet, "/api/v1/jobs/" + jobID, nil, "collection"},

		// top-level documents (read-only / rename, never deletes)
		"GET /documents":              {http.MethodGet, "/api/v1/documents", nil, "collection-wildcard"},
		"GET /documents/:id":          {http.MethodGet, "/api/v1/documents/" + docID, nil, "collection"},
		"GET /documents/:id/chunks":   {http.MethodGet, "/api/v1/documents/" + docID + "/chunks", nil, "collection"},
		"GET /documents/:id/metadata": {http.MethodGet, "/api/v1/documents/" + docID + "/metadata", nil, "collection"},
		"PATCH /documents/:id":        {http.MethodPatch, "/api/v1/documents/" + docID, renameBody, "collection"},

		// keys
		"GET /keys":        {http.MethodGet, "/api/v1/keys", nil, "admin"},
		"POST /keys":       {http.MethodPost, "/api/v1/keys", nil, "admin"},
		"DELETE /keys/:id": {http.MethodDelete, "/api/v1/keys/nonexistent", nil, "admin"},

		// prompts
		"GET /prompts":          {http.MethodGet, "/api/v1/prompts", nil, "admin"},
		"POST /prompts":         {http.MethodPost, "/api/v1/prompts", nil, "admin"},
		"GET /prompts/:slug":    {http.MethodGet, "/api/v1/prompts/nonexistent", nil, "admin"},
		"PATCH /prompts/:slug":  {http.MethodPatch, "/api/v1/prompts/nonexistent", nil, "admin"},
		"DELETE /prompts/:slug": {http.MethodDelete, "/api/v1/prompts/nonexistent", nil, "admin"},

		// collection lifecycle
		"POST /collections":          {http.MethodPost, "/api/v1/collections", nil, "admin"},
		"GET /collections":           {http.MethodGet, "/api/v1/collections", nil, "admin"},
		"GET /collections/id/:id":    {http.MethodGet, "/api/v1/collections/id/nonexistent", nil, "admin"},
		"PATCH /collections/id/:id":  {http.MethodPatch, "/api/v1/collections/id/nonexistent", nil, "admin"},
		"DELETE /collections/id/:id": {http.MethodDelete, "/api/v1/collections/id/nonexistent", nil, "admin"},
		"GET /collections/:slug":     {http.MethodGet, "/api/v1/collections/" + slugA, nil, "admin"},
		"DELETE /collections/:slug":  {http.MethodDelete, "/api/v1/collections/nonexistent-slug", nil, "admin"},

		// collection-scoped jobs (read-only)
		"GET /collections/:slug/jobs":     {http.MethodGet, "/api/v1/collections/" + slugA + "/jobs", nil, "collection"},
		"GET /collections/:slug/jobs/:id": {http.MethodGet, "/api/v1/collections/" + slugA + "/jobs/" + jobID, nil, "collection"},

		// ingest / query / rag
		"POST /collections/:slug/ingest": {http.MethodPost, "/api/v1/collections/" + slugA + "/ingest", nil, "collection"},
		"POST /collections/:slug/query":  {http.MethodPost, "/api/v1/collections/" + slugA + "/query", nil, "collection"},
		"POST /collections/:slug/rag":    {http.MethodPost, "/api/v1/collections/" + slugA + "/rag", nil, "collection"},

		// collection-scoped documents (read-only / rename)
		"GET /collections/:slug/documents":              {http.MethodGet, "/api/v1/collections/" + slugA + "/documents", nil, "collection"},
		"GET /collections/:slug/documents/:id":          {http.MethodGet, "/api/v1/collections/" + slugA + "/documents/" + docID, nil, "collection"},
		"GET /collections/:slug/documents/:id/chunks":   {http.MethodGet, "/api/v1/collections/" + slugA + "/documents/" + docID + "/chunks", nil, "collection"},
		"GET /collections/:slug/documents/:id/metadata": {http.MethodGet, "/api/v1/collections/" + slugA + "/documents/" + docID + "/metadata", nil, "collection"},
		"PATCH /collections/:slug/documents/:id":        {http.MethodPatch, "/api/v1/collections/" + slugA + "/documents/" + docID, renameBody, "collection"},

		// metadata-fields
		"GET /collections/:slug/metadata-fields":          {http.MethodGet, "/api/v1/collections/" + slugA + "/metadata-fields", nil, "collection"},
		"POST /collections/:slug/metadata-fields":         {http.MethodPost, "/api/v1/collections/" + slugA + "/metadata-fields", nil, "collection"},
		"DELETE /collections/:slug/metadata-fields/:name": {http.MethodDelete, "/api/v1/collections/" + slugA + "/metadata-fields/nonexistent", nil, "collection"},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if resp := doAuthJSON(t, a.Public, tc.method, tc.path, otherKey, tc.body); resp.StatusCode != fiber.StatusForbidden {
				t.Errorf("key with no relevant grant: expected 403, got %d", resp.StatusCode)
			}

			goodKey := collectionKey
			if tc.category != "collection" {
				goodKey = adminKey
			}
			if resp := doAuthJSON(t, a.Public, tc.method, tc.path, goodKey, tc.body); resp.StatusCode == fiber.StatusForbidden {
				t.Errorf("key with the right %s grant: expected to clear the ACL gate, got 403", tc.category)
			}
		})
	}

	// The four delete-a-specific-resource routes (document/job x top-level/
	// collection-scoped) each get their own disposable fixture — sharing one
	// with the read-only cases above would let whichever runs first (map
	// order is random) actually delete it out from under the others, and
	// once gone, GetDocument/GetJob 404s *before* the ACL check ever runs
	// (see documents.go/jobs.go's inline checkAccess), which would make
	// otherKey's "should be 403" assertion see 404 instead — a false
	// failure caused by fixture reuse, not a real ACL gap.
	disposableDoc := func(t *testing.T, name string) string {
		t.Helper()
		helpers.UploadDoc(t, a.Admin, slugA, name, "disposable", nil)
		doc := helpers.WaitForDocument(t, a.Admin, slugA, name)
		id, _ := doc["id"].(string)
		if id == "" {
			t.Fatalf("disposable document fixture failed for %q", name)
		}
		return id
	}
	disposableJob := func(t *testing.T, name string) string {
		t.Helper()
		job := helpers.UploadDoc(t, a.Admin, slugA, name, "disposable", nil)
		helpers.WaitForDocument(t, a.Admin, slugA, name)
		id, _ := job["id"].(string)
		if id == "" {
			t.Fatalf("disposable job fixture failed for %q", name)
		}
		return id
	}

	deleteCases := []struct {
		name string
		path func(id string) string
		id   string
	}{
		{"DELETE /documents/:id", func(id string) string { return "/api/v1/documents/" + id }, disposableDoc(t, "route-cov-del-1.txt")},
		{"DELETE /collections/:slug/documents/:id", func(id string) string { return "/api/v1/collections/" + slugA + "/documents/" + id }, disposableDoc(t, "route-cov-del-2.txt")},
		{"DELETE /jobs/:id", func(id string) string { return "/api/v1/jobs/" + id }, disposableJob(t, "route-cov-del-3.txt")},
		{"DELETE /collections/:slug/jobs/:id", func(id string) string { return "/api/v1/collections/" + slugA + "/jobs/" + id }, disposableJob(t, "route-cov-del-4.txt")},
	}
	for _, dc := range deleteCases {
		t.Run(dc.name, func(t *testing.T) {
			path := dc.path(dc.id)
			if resp := doAuthJSON(t, a.Public, http.MethodDelete, path, otherKey, nil); resp.StatusCode != fiber.StatusForbidden {
				t.Errorf("key with no relevant grant: expected 403, got %d", resp.StatusCode)
			}
			if resp := doAuthJSON(t, a.Public, http.MethodDelete, path, collectionKey, nil); resp.StatusCode == fiber.StatusForbidden {
				t.Errorf("key with collection grant: expected to clear the ACL gate, got 403")
			}
		})
	}
}
