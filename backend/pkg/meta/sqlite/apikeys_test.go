package sqlite

import (
	"context"
	"path/filepath"
	"testing"

	"ragpack/pkg/meta"
)

func newTestStore(t *testing.T) *MetaStore {
	t.Helper()
	ms, err := New(filepath.Join(t.TempDir(), "meta.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { ms.Close() })
	return ms
}

func TestCreateAPIKey_RequiresAtLeastOneGrantOfEitherKind(t *testing.T) {
	ms := newTestStore(t)
	if _, err := ms.CreateAPIKey(context.Background(), "no-grants", "plaintext", nil, nil); err == nil {
		t.Fatal("expected error creating a key with no grants of any kind, got nil")
	}
	if _, err := ms.CreateAPIKey(context.Background(), "empty-grants", "plaintext", []meta.GrantInput{}, nil); err == nil {
		t.Fatal("expected error creating a key with an empty grant slice and no admin grants, got nil")
	}
}

func TestCreateAPIKey_AdminOnlyKeyRequiresNoCollectionGrant(t *testing.T) {
	ms := newTestStore(t)
	ctx := context.Background()

	// Zero collection grants, one admin grant — a pure instance-admin key
	// (e.g. manages prompts, never touches any collection's content) must
	// be creatable without a throwaway collection grant it'll never use.
	key, err := ms.CreateAPIKey(ctx, "admin-only", "plaintext", nil, []meta.AdminGrantInput{
		{ResourceType: meta.ResourcePrompts, Permission: meta.PermissionWrite},
	})
	if err != nil {
		t.Fatalf("create admin-only api key: %v", err)
	}

	grants, err := ms.ListGrants(ctx, key.ID)
	if err != nil {
		t.Fatalf("list grants: %v", err)
	}
	if len(grants) != 0 {
		t.Errorf("expected 0 collection grants, got %d", len(grants))
	}
	adminGrants, err := ms.ListAdminGrants(ctx, key.ID)
	if err != nil {
		t.Fatalf("list admin grants: %v", err)
	}
	if len(adminGrants) != 1 {
		t.Errorf("expected 1 admin grant, got %d", len(adminGrants))
	}
}

func TestCreateAPIKey_RollsBackOnGrantFailure(t *testing.T) {
	ms := newTestStore(t)
	// An invalid permission value violates the CHECK constraint on
	// api_key_collections, which must fail the whole transaction — the key
	// row must not be left behind without its grants.
	_, err := ms.CreateAPIKey(context.Background(), "bad-grant", "plaintext", []meta.GrantInput{
		{Permission: "not-a-real-permission"},
	}, nil)
	if err == nil {
		t.Fatal("expected error creating a key with an invalid grant permission, got nil")
	}
	count, err := ms.CountAPIKeys(context.Background())
	if err != nil {
		t.Fatalf("count api keys: %v", err)
	}
	if count != 0 {
		t.Errorf("expected the failed create to leave no key row behind, found %d", count)
	}
}

func TestCreateAPIKey_RollsBackOnAdminGrantFailure(t *testing.T) {
	ms := newTestStore(t)
	// Same atomicity guarantee, but for a bad admin grant instead of a bad
	// collection grant — the collection grant here is valid, so this proves
	// the transaction covers both tables, not just the first one inserted.
	// resource_type has no CHECK constraint (see meta.ResourceType — the app
	// layer validates it, not the DB), so this uses a bad permission instead,
	// which is still CHECK-constrained on both grant tables.
	_, err := ms.CreateAPIKey(context.Background(), "bad-admin-grant", "plaintext",
		[]meta.GrantInput{{Permission: meta.PermissionRead}},
		[]meta.AdminGrantInput{{ResourceType: meta.ResourceKeys, Permission: "not-a-real-permission"}},
	)
	if err == nil {
		t.Fatal("expected error creating a key with an invalid admin grant resource type, got nil")
	}
	count, err := ms.CountAPIKeys(context.Background())
	if err != nil {
		t.Fatalf("count api keys: %v", err)
	}
	if count != 0 {
		t.Errorf("expected the failed create to leave no key row behind, found %d", count)
	}
}

func TestCreateAPIKey_AdminGrantsAreOptional(t *testing.T) {
	ms := newTestStore(t)
	ctx := context.Background()

	key, err := ms.CreateAPIKey(ctx, "content-only", "plaintext", []meta.GrantInput{{Permission: meta.PermissionRead}}, nil)
	if err != nil {
		t.Fatalf("create api key with no admin grants: %v", err)
	}
	adminGrants, err := ms.ListAdminGrants(ctx, key.ID)
	if err != nil {
		t.Fatalf("list admin grants: %v", err)
	}
	if len(adminGrants) != 0 {
		t.Errorf("expected 0 admin grants for a content-only key, got %d", len(adminGrants))
	}
}

func TestCreateAPIKey_StoresGrantsAndValidates(t *testing.T) {
	ms := newTestStore(t)
	ctx := context.Background()

	key, err := ms.CreateAPIKey(ctx, "test-key", "plaintext-value", []meta.GrantInput{
		{Permission: meta.PermissionRead},
		{CollectionID: strPtr("col-b"), Permission: meta.PermissionWrite},
	}, []meta.AdminGrantInput{
		{ResourceType: meta.ResourcePrompts, Permission: meta.PermissionWrite},
	})
	if err != nil {
		t.Fatalf("create api key: %v", err)
	}

	grants, err := ms.ListGrants(ctx, key.ID)
	if err != nil {
		t.Fatalf("list grants: %v", err)
	}
	if len(grants) != 2 {
		t.Fatalf("expected 2 grants, got %d", len(grants))
	}

	adminGrants, err := ms.ListAdminGrants(ctx, key.ID)
	if err != nil {
		t.Fatalf("list admin grants: %v", err)
	}
	if len(adminGrants) != 1 || adminGrants[0].ResourceType != meta.ResourcePrompts {
		t.Fatalf("expected 1 admin grant on 'prompts', got %+v", adminGrants)
	}

	validated, err := ms.ValidateAPIKey(ctx, "plaintext-value")
	if err != nil {
		t.Fatalf("validate api key: %v", err)
	}
	if validated.ID != key.ID {
		t.Errorf("expected validated key ID %q, got %q", key.ID, validated.ID)
	}

	if _, err := ms.ValidateAPIKey(ctx, "wrong-value"); err == nil {
		t.Error("expected validation to fail for a wrong plaintext, got nil error")
	}
}

func strPtr(s string) *string { return &s }
