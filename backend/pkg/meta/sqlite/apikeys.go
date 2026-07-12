package sqlite

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"ragpack/pkg/auth"
	"ragpack/pkg/meta"
)

func (s *MetaStore) CreateAPIKey(ctx context.Context, name, plaintext string, grants []meta.GrantInput, adminGrants []meta.AdminGrantInput) (meta.APIKey, error) {
	if len(grants) == 0 {
		return meta.APIKey{}, fmt.Errorf("sqlite: create api key: at least one grant is required")
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return meta.APIKey{}, fmt.Errorf("sqlite: begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	k := meta.APIKey{
		ID:        uuid.NewString(),
		Name:      name,
		KeyHint:   auth.Hint(plaintext),
		CreatedAt: time.Now().UTC(),
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO api_keys (id, name, key_hash, key_hint, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, k.ID, k.Name, auth.Hash(plaintext), k.KeyHint, k.CreatedAt); err != nil {
		return meta.APIKey{}, fmt.Errorf("sqlite: create api key: %w", err)
	}

	now := time.Now().UTC()
	for _, g := range grants {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO api_key_collections (id, api_key_id, collection_id, permission, created_at)
			VALUES (?, ?, ?, ?, ?)
		`, uuid.NewString(), k.ID, g.CollectionID, string(g.Permission), now); err != nil {
			return meta.APIKey{}, fmt.Errorf("sqlite: create api key grant: %w", err)
		}
	}

	for _, g := range adminGrants {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO api_key_admin_grants (id, api_key_id, resource_type, permission, created_at)
			VALUES (?, ?, ?, ?, ?)
		`, uuid.NewString(), k.ID, string(g.ResourceType), string(g.Permission), now); err != nil {
			return meta.APIKey{}, fmt.Errorf("sqlite: create api key admin grant: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return meta.APIKey{}, fmt.Errorf("sqlite: commit api key: %w", err)
	}
	return k, nil
}

func (s *MetaStore) ValidateAPIKey(ctx context.Context, plaintext string) (meta.APIKey, error) {
	hash := auth.Hash(plaintext)
	var k meta.APIKey
	if err := s.db.GetContext(ctx, &k, `SELECT id, name, key_hint, created_at, last_used_at FROM api_keys WHERE key_hash = ?`, hash); err != nil {
		return meta.APIKey{}, fmt.Errorf("invalid key")
	}
	_, _ = s.db.ExecContext(ctx, `UPDATE api_keys SET last_used_at = ? WHERE id = ?`, time.Now().UTC(), k.ID)
	return k, nil
}

func (s *MetaStore) ListAPIKeys(ctx context.Context) ([]meta.APIKey, error) {
	var keys []meta.APIKey
	if err := s.db.SelectContext(ctx, &keys, `
		SELECT id, name, key_hint, created_at, last_used_at
		FROM api_keys ORDER BY created_at DESC
	`); err != nil {
		return nil, fmt.Errorf("sqlite: list api keys: %w", err)
	}
	return keys, nil
}

func (s *MetaStore) ListGrants(ctx context.Context, apiKeyID string) ([]meta.CollectionGrant, error) {
	var grants []meta.CollectionGrant
	if err := s.db.SelectContext(ctx, &grants, `
		SELECT id, api_key_id, collection_id, permission, created_at
		FROM api_key_collections
		WHERE api_key_id = ?
		ORDER BY created_at ASC
	`, apiKeyID); err != nil {
		return nil, fmt.Errorf("sqlite: list api key grants: %w", err)
	}
	return grants, nil
}

func (s *MetaStore) ListAdminGrants(ctx context.Context, apiKeyID string) ([]meta.AdminGrant, error) {
	var grants []meta.AdminGrant
	if err := s.db.SelectContext(ctx, &grants, `
		SELECT id, api_key_id, resource_type, permission, created_at
		FROM api_key_admin_grants
		WHERE api_key_id = ?
		ORDER BY created_at ASC
	`, apiKeyID); err != nil {
		return nil, fmt.Errorf("sqlite: list api key admin grants: %w", err)
	}
	return grants, nil
}

func (s *MetaStore) DeleteAPIKey(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM api_keys WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("sqlite: delete api key %q: %w", id, err)
	}
	return nil
}

func (s *MetaStore) CountAPIKeys(ctx context.Context) (int, error) {
	var count int
	if err := s.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM api_keys`); err != nil {
		return 0, fmt.Errorf("sqlite: count api keys: %w", err)
	}
	return count, nil
}
