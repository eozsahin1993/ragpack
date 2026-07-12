package middleware

import (
	"testing"

	"ragpack/pkg/meta"
)

func TestHasAdminAccess_NoGrantsFailsClosed(t *testing.T) {
	if hasAdminAccess(nil, meta.ResourceKeys, meta.PermissionRead) {
		t.Error("expected no access with zero grants (fail closed), got access")
	}
	if hasAdminAccess([]meta.AdminGrant{}, meta.ResourceKeys, meta.PermissionRead) {
		t.Error("expected no access with empty grant slice, got access")
	}
}

func TestHasAdminAccess_SpecificResourceTypeOnlyCoversItsOwnType(t *testing.T) {
	grants := []meta.AdminGrant{
		{ResourceType: meta.ResourcePrompts, Permission: meta.PermissionBoth},
	}
	if !hasAdminAccess(grants, meta.ResourcePrompts, meta.PermissionRead) {
		t.Error("expected access to prompts, got none")
	}
	if hasAdminAccess(grants, meta.ResourceKeys, meta.PermissionRead) {
		t.Error("expected no access to keys (not granted), got access")
	}
	if hasAdminAccess(grants, meta.ResourceCollections, meta.PermissionRead) {
		t.Error("expected no access to collections (not granted), got access")
	}
}

func TestHasAdminAccess_WildcardCoversAnyResourceType_IncludingUnknownFutureOnes(t *testing.T) {
	grants := []meta.AdminGrant{
		{ResourceType: meta.ResourceAll, Permission: meta.PermissionRead},
	}
	for _, rt := range []meta.ResourceType{meta.ResourceKeys, meta.ResourcePrompts, meta.ResourceCollections, "webhooks"} {
		if !hasAdminAccess(grants, rt, meta.PermissionRead) {
			t.Errorf("expected \"*\" grant to cover resource type %q, got no access", rt)
		}
	}
}

func TestHasAdminAccess_PermissionTiers(t *testing.T) {
	cases := []struct {
		name     string
		granted  meta.Permission
		required meta.Permission
		want     bool
	}{
		{"read covers read", meta.PermissionRead, meta.PermissionRead, true},
		{"read does not cover write", meta.PermissionRead, meta.PermissionWrite, false},
		{"write covers write", meta.PermissionWrite, meta.PermissionWrite, true},
		{"write does not cover read", meta.PermissionWrite, meta.PermissionRead, false},
		{"both covers read", meta.PermissionBoth, meta.PermissionRead, true},
		{"both covers write", meta.PermissionBoth, meta.PermissionWrite, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			grants := []meta.AdminGrant{{ResourceType: meta.ResourceKeys, Permission: tc.granted}}
			if got := hasAdminAccess(grants, meta.ResourceKeys, tc.required); got != tc.want {
				t.Errorf("granted=%s required=%s: got %v, want %v", tc.granted, tc.required, got, tc.want)
			}
		})
	}
}

func TestHasAdminAccess_MultipleGrantsAreIndependentPerResourceType(t *testing.T) {
	grants := []meta.AdminGrant{
		{ResourceType: meta.ResourceKeys, Permission: meta.PermissionRead},
		{ResourceType: meta.ResourcePrompts, Permission: meta.PermissionWrite},
	}
	if !hasAdminAccess(grants, meta.ResourceKeys, meta.PermissionRead) {
		t.Error("expected read access to keys")
	}
	if hasAdminAccess(grants, meta.ResourceKeys, meta.PermissionWrite) {
		t.Error("expected no write access to keys (only granted read)")
	}
	if !hasAdminAccess(grants, meta.ResourcePrompts, meta.PermissionWrite) {
		t.Error("expected write access to prompts")
	}
	if hasAdminAccess(grants, meta.ResourceCollections, meta.PermissionRead) {
		t.Error("expected no access to collections (no matching grant)")
	}
}
