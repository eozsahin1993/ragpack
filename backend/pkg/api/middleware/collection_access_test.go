package middleware

import (
	"testing"

	"ragpack/pkg/meta"
)

func strPtr(s string) *string { return &s }

func TestHasAccess_NoGrantsFailsClosed(t *testing.T) {
	if hasAccess(nil, "col-a", meta.PermissionRead) {
		t.Error("expected no access with zero grants (fail closed), got access")
	}
	if hasAccess([]meta.CollectionGrant{}, "col-a", meta.PermissionRead) {
		t.Error("expected no access with empty grant slice, got access")
	}
}

func TestHasAccess_SpecificGrantOnlyCoversItsOwnCollection(t *testing.T) {
	grants := []meta.CollectionGrant{
		{CollectionID: strPtr("col-a"), Permission: meta.PermissionBoth},
	}
	if !hasAccess(grants, "col-a", meta.PermissionRead) {
		t.Error("expected access to col-a, got none")
	}
	if hasAccess(grants, "col-b", meta.PermissionRead) {
		t.Error("expected no access to col-b (not granted), got access")
	}
}

func TestHasAccess_WildcardCoversAnyCollection(t *testing.T) {
	grants := []meta.CollectionGrant{
		{CollectionID: nil, Permission: meta.PermissionRead},
	}
	for _, col := range []string{"col-a", "col-b", "brand-new-collection-not-yet-created"} {
		if !hasAccess(grants, col, meta.PermissionRead) {
			t.Errorf("expected wildcard grant to cover %q, got no access", col)
		}
	}
}

func TestHasAccess_PermissionTiers(t *testing.T) {
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
			grants := []meta.CollectionGrant{{CollectionID: strPtr("col-a"), Permission: tc.granted}}
			if got := hasAccess(grants, "col-a", tc.required); got != tc.want {
				t.Errorf("granted=%s required=%s: got %v, want %v", tc.granted, tc.required, got, tc.want)
			}
		})
	}
}

func TestHasAccess_MultipleGrantsAnyMatchWins(t *testing.T) {
	grants := []meta.CollectionGrant{
		{CollectionID: strPtr("col-a"), Permission: meta.PermissionRead},
		{CollectionID: strPtr("col-b"), Permission: meta.PermissionWrite},
	}
	if !hasAccess(grants, "col-a", meta.PermissionRead) {
		t.Error("expected read access to col-a")
	}
	if hasAccess(grants, "col-a", meta.PermissionWrite) {
		t.Error("expected no write access to col-a (only granted read)")
	}
	if !hasAccess(grants, "col-b", meta.PermissionWrite) {
		t.Error("expected write access to col-b")
	}
	if hasAccess(grants, "col-c", meta.PermissionRead) {
		t.Error("expected no access to col-c (no matching grant)")
	}
}

func TestCovers(t *testing.T) {
	if !covers(meta.PermissionBoth, meta.PermissionRead) {
		t.Error("both should cover read")
	}
	if !covers(meta.PermissionBoth, meta.PermissionWrite) {
		t.Error("both should cover write")
	}
	if !covers(meta.PermissionRead, meta.PermissionRead) {
		t.Error("read should cover read")
	}
	if covers(meta.PermissionRead, meta.PermissionWrite) {
		t.Error("read should not cover write")
	}
}
