package meta

import (
	"context"
	"time"
)

type APIKey struct {
	ID         string     `db:"id"           json:"id"`
	Name       string     `db:"name"         json:"name"`
	KeyHint    string     `db:"key_hint"     json:"key_hint"`
	CreatedAt  time.Time  `db:"created_at"   json:"created_at"`
	LastUsedAt *time.Time `db:"last_used_at" json:"last_used_at,omitempty"`
}

type APIKeyStore interface {
	CreateAPIKey(ctx context.Context, name, plaintext string) (APIKey, error)
	ValidateAPIKey(ctx context.Context, plaintext string) error
	ListAPIKeys(ctx context.Context) ([]APIKey, error)
	DeleteAPIKey(ctx context.Context, id string) error
	CountAPIKeys(ctx context.Context) (int, error)
}
