package helpers

import (
	"context"
	"testing"

	"ragpack/pkg/meta"
)

// CreateAPIKey creates a content-access-only API key (no admin grants)
// directly against the meta store (bypassing the /keys endpoint, which
// itself requires an admin grant) and returns its plaintext value for use in
// an Authorization header.
func CreateAPIKey(t *testing.T, ms meta.MetaStore, name string, grants ...meta.GrantInput) string {
	t.Helper()
	return CreateAdminAPIKey(t, ms, name, nil, grants...)
}

// CreateAdminAPIKey is like CreateAPIKey but also attaches adminGrants
// (instance-administration capability — see meta.AdminGrant).
func CreateAdminAPIKey(t *testing.T, ms meta.MetaStore, name string, adminGrants []meta.AdminGrantInput, grants ...meta.GrantInput) string {
	t.Helper()
	const plaintext = "test-key-"
	if _, err := ms.CreateAPIKey(context.Background(), name, plaintext+name, grants, adminGrants); err != nil {
		t.Fatalf("create api key %q: %v", name, err)
	}
	return plaintext + name
}

// WildcardGrant returns a grant covering every collection (current and
// future) at the given permission level.
func WildcardGrant(perm meta.Permission) meta.GrantInput {
	return meta.GrantInput{Permission: perm}
}

// CollectionGrant returns a grant scoped to one collection ID.
func CollectionGrant(collectionID string, perm meta.Permission) meta.GrantInput {
	return meta.GrantInput{CollectionID: &collectionID, Permission: perm}
}

// AdminGrant returns an admin grant on the given resource type.
func AdminGrant(resourceType meta.ResourceType, perm meta.Permission) meta.AdminGrantInput {
	return meta.AdminGrantInput{ResourceType: resourceType, Permission: perm}
}
