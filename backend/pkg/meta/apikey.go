package meta

import (
	"context"
	"time"
)

type Permission string

const (
	PermissionRead  Permission = "read"
	PermissionWrite Permission = "write"
	PermissionBoth  Permission = "both"
)

// ResourceType is an instance-administration resource, decoupled from any CollectionGrant. "*" matches every type, current and future.
type ResourceType string

const (
	ResourceKeys        ResourceType = "keys"
	ResourcePrompts     ResourceType = "prompts"
	ResourceCollections ResourceType = "collections"
	ResourceAll         ResourceType = "*"
)

type APIKey struct {
	ID        string    `db:"id"         json:"id"`
	Name      string    `db:"name"       json:"name"`
	KeyHint   string    `db:"key_hint"   json:"key_hint"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// CollectionGrant grants access to one collection, or every collection (current and future) when CollectionID is nil.
type CollectionGrant struct {
	ID           string     `db:"id"            json:"id"`
	APIKeyID     string     `db:"api_key_id"    json:"api_key_id"`
	CollectionID *string    `db:"collection_id" json:"collection_id,omitempty"`
	Permission   Permission `db:"permission"    json:"permission"`
	CreatedAt    time.Time  `db:"created_at"    json:"created_at"`
}

// GrantInput describes a grant to attach to a key at creation time. See CollectionGrant.
type GrantInput struct {
	CollectionID *string
	Permission   Permission
}

// AdminGrant grants a capability over an instance-administration resource type. See ResourceType.
type AdminGrant struct {
	ID           string       `db:"id"            json:"id"`
	APIKeyID     string       `db:"api_key_id"    json:"api_key_id"`
	ResourceType ResourceType `db:"resource_type" json:"resource_type"`
	Permission   Permission   `db:"permission"    json:"permission"`
	CreatedAt    time.Time    `db:"created_at"    json:"created_at"`
}

// AdminGrantInput describes an admin grant to attach to a key at creation time. Optional — most keys have none.
type AdminGrantInput struct {
	ResourceType ResourceType
	Permission   Permission
}

type APIKeyStore interface {
	// CreateAPIKey creates a key and its grants atomically. At least one of
	// grants/adminGrants must be non-empty — a key with neither can access
	// nothing at all, collection or admin.
	CreateAPIKey(ctx context.Context, name, plaintext string, grants []GrantInput, adminGrants []AdminGrantInput) (APIKey, error)
	// ValidateAPIKey returns the matching key record (caller needs its ID to check grants).
	ValidateAPIKey(ctx context.Context, plaintext string) (APIKey, error)
	ListAPIKeys(ctx context.Context) ([]APIKey, error)
	ListGrants(ctx context.Context, apiKeyID string) ([]CollectionGrant, error)
	ListAdminGrants(ctx context.Context, apiKeyID string) ([]AdminGrant, error)
	DeleteAPIKey(ctx context.Context, id string) error
	CountAPIKeys(ctx context.Context) (int, error)
}
