package sqlite

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"ragpack/pkg/auth"
	"ragpack/pkg/meta"
)

func (s *MetaStore) CreateAPIKey(ctx context.Context, name, plaintext string) (meta.APIKey, error) {
	k := meta.APIKey{
		ID:        uuid.NewString(),
		Name:      name,
		KeyHint:   auth.Hint(plaintext),
		CreatedAt: time.Now().UTC(),
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO api_keys (id, name, key_hash, key_hint, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, k.ID, k.Name, auth.Hash(plaintext), k.KeyHint, k.CreatedAt.Unix())
	if err != nil {
		return meta.APIKey{}, fmt.Errorf("sqlite: create api key: %w", err)
	}
	return k, nil
}

func (s *MetaStore) ValidateAPIKey(ctx context.Context, plaintext string) error {
	hash := auth.Hash(plaintext)
	var id string
	if err := s.db.GetContext(ctx, &id, `SELECT id FROM api_keys WHERE key_hash = ?`, hash); err != nil {
		return fmt.Errorf("invalid key")
	}
	_, _ = s.db.ExecContext(ctx, `UPDATE api_keys SET last_used_at = ? WHERE id = ?`, time.Now().UTC().Unix(), id)
	return nil
}

func (s *MetaStore) ListAPIKeys(ctx context.Context) ([]meta.APIKey, error) {
	type row struct {
		ID          string  `db:"id"`
		Name        string  `db:"name"`
		KeyHint     string  `db:"key_hint"`
		CreatedAt   int64   `db:"created_at"`
		LastUsedAt  *int64  `db:"last_used_at"`
	}
	var rows []row
	if err := s.db.SelectContext(ctx, &rows, `
		SELECT id, name, key_hint, created_at, last_used_at
		FROM api_keys ORDER BY created_at DESC
	`); err != nil {
		return nil, fmt.Errorf("sqlite: list api keys: %w", err)
	}
	keys := make([]meta.APIKey, len(rows))
	for i, r := range rows {
		keys[i] = meta.APIKey{
			ID:        r.ID,
			Name:      r.Name,
			KeyHint:   r.KeyHint,
			CreatedAt: time.Unix(r.CreatedAt, 0).UTC(),
		}
		if r.LastUsedAt != nil {
			t := time.Unix(*r.LastUsedAt, 0).UTC()
			keys[i].LastUsedAt = &t
		}
	}
	return keys, nil
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
