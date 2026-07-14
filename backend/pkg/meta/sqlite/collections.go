package sqlite

import (
	"context"
	"fmt"
	"strings"
	"time"

	"ragpack/pkg/meta"
)

func (s *MetaStore) CreateCollection(ctx context.Context, input meta.CreateCollectionInput) (meta.Collection, error) {
	id, tableName := newTableName(input.Name)
	c := meta.Collection{
		ID:            id,
		Name:          input.Name,
		Slug:          slugify(input.Name + "_" + input.EmbedModel),
		TableName:     tableName,
		EmbedModel:    input.EmbedModel,
		VectorDim:     input.VectorDim,
		CreatedAt:     time.Now().UTC(),
		ChunkStrategy: input.ChunkStrategy,
		ChunkSize:     input.ChunkSize,
		ChunkOverlap:  input.ChunkOverlap,
	}

	_, err := s.db.NamedExecContext(ctx, `
		INSERT INTO collections
			(id, name, slug, table_name, embed_model, vector_dim, created_at, chunk_strategy, chunk_size, chunk_overlap)
		VALUES
			(:id, :name, :slug, :table_name, :embed_model, :vector_dim, :created_at, :chunk_strategy, :chunk_size, :chunk_overlap)
	`, c)
	if err != nil {
		return meta.Collection{}, fmt.Errorf("sqlite: create collection %q: %w", input.Name, err)
	}
	return c, nil
}

func (s *MetaStore) GetCollectionBySlug(ctx context.Context, slug string) (meta.Collection, error) {
	var c meta.Collection
	err := s.db.GetContext(ctx, &c, `
		SELECT id, name, slug, table_name, embed_model, vector_dim, created_at,
		       chunk_strategy, chunk_size, chunk_overlap, refresh_enabled, refresh_interval_seconds, last_auto_refresh_at
		FROM collections
		WHERE slug = ?
	`, slug)
	if err != nil {
		return meta.Collection{}, fmt.Errorf("sqlite: get collection by slug %q: %w", slug, err)
	}
	return c, nil
}

func (s *MetaStore) GetCollectionByID(ctx context.Context, id string) (meta.Collection, error) {
	var c meta.Collection
	err := s.db.GetContext(ctx, &c, `
		SELECT id, name, slug, table_name, embed_model, vector_dim, created_at,
		       chunk_strategy, chunk_size, chunk_overlap, refresh_enabled, refresh_interval_seconds, last_auto_refresh_at
		FROM collections
		WHERE id = ?
	`, id)
	if err != nil {
		return meta.Collection{}, fmt.Errorf("sqlite: get collection by id %q: %w", id, err)
	}
	return c, nil
}

func (s *MetaStore) ListCollections(ctx context.Context, limit, offset int) ([]meta.Collection, error) {
	var collections []meta.Collection
	err := s.db.SelectContext(ctx, &collections, `
		SELECT id, name, slug, table_name, embed_model, vector_dim, created_at,
		       chunk_strategy, chunk_size, chunk_overlap, refresh_enabled, refresh_interval_seconds, last_auto_refresh_at
		FROM collections
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("sqlite: list collections: %w", err)
	}
	return collections, nil
}

func (s *MetaStore) ListAllCollections(ctx context.Context) ([]meta.Collection, error) {
	var collections []meta.Collection
	err := s.db.SelectContext(ctx, &collections, `
		SELECT id, name, slug, table_name, embed_model, vector_dim, created_at,
		       chunk_strategy, chunk_size, chunk_overlap, refresh_enabled, refresh_interval_seconds, last_auto_refresh_at
		FROM collections
		ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("sqlite: list all collections: %w", err)
	}
	return collections, nil
}

func (s *MetaStore) CountCollections(ctx context.Context) (int, error) {
	var count int
	err := s.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM collections`)
	if err != nil {
		return 0, fmt.Errorf("sqlite: count collections: %w", err)
	}
	return count, nil
}

// UpdateCollection applies a partial update; nil patch fields are untouched.
// Name additionally recomputes slug (embed_model, fetched fresh, is part of it).
func (s *MetaStore) UpdateCollection(ctx context.Context, id string, patch meta.CollectionPatch) (meta.Collection, error) {
	var clauses []string
	var args []any

	if patch.Name != nil {
		var embedModel string
		if err := s.db.GetContext(ctx, &embedModel, `SELECT embed_model FROM collections WHERE id = ?`, id); err != nil {
			return meta.Collection{}, fmt.Errorf("sqlite: update collection %q: %w", id, err)
		}
		clauses = append(clauses, "name = ?", "slug = ?")
		args = append(args, *patch.Name, slugify(*patch.Name+"_"+embedModel))
	}
	if patch.RefreshEnabled != nil {
		clauses = append(clauses, "refresh_enabled = ?")
		args = append(args, *patch.RefreshEnabled)
	}
	if patch.RefreshIntervalSeconds != nil {
		clauses = append(clauses, "refresh_interval_seconds = ?")
		args = append(args, *patch.RefreshIntervalSeconds)
	}
	if len(clauses) == 0 {
		return s.GetCollectionByID(ctx, id)
	}
	args = append(args, id)

	_, err := s.db.ExecContext(ctx, "UPDATE collections SET "+strings.Join(clauses, ", ")+" WHERE id = ?", args...)
	if err != nil {
		return meta.Collection{}, fmt.Errorf("sqlite: update collection %q: %w", id, err)
	}
	return s.GetCollectionByID(ctx, id)
}

// ListCollectionsDueForAutoRefresh compares via strftime unix-epoch-seconds, not julianday — refresh_interval_seconds is already seconds.
func (s *MetaStore) ListCollectionsDueForAutoRefresh(ctx context.Context, now time.Time) ([]meta.Collection, error) {
	var collections []meta.Collection
	err := s.db.SelectContext(ctx, &collections, `
		SELECT id, name, slug, table_name, embed_model, vector_dim, created_at,
		       chunk_strategy, chunk_size, chunk_overlap, refresh_enabled, refresh_interval_seconds, last_auto_refresh_at
		FROM collections
		WHERE refresh_enabled = 1
		  AND refresh_interval_seconds IS NOT NULL
		  AND (
		        last_auto_refresh_at IS NULL
		        OR (strftime('%s', ?) - strftime('%s', last_auto_refresh_at)) >= refresh_interval_seconds
		      )
		ORDER BY created_at ASC
	`, now)
	if err != nil {
		return nil, fmt.Errorf("sqlite: list collections due for auto-refresh: %w", err)
	}
	return collections, nil
}

// TouchCollectionAutoRefreshed stamps last_auto_refresh_at once per check,
// not once per document — see checkCollection in refresh_scheduler.go.
func (s *MetaStore) TouchCollectionAutoRefreshed(ctx context.Context, id string, at time.Time) error {
	_, err := s.db.ExecContext(ctx, `UPDATE collections SET last_auto_refresh_at = ? WHERE id = ?`, at, id)
	if err != nil {
		return fmt.Errorf("sqlite: touch collection auto-refreshed %q: %w", id, err)
	}
	return nil
}

func (s *MetaStore) DeleteCollection(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `
		DELETE FROM collections WHERE id = ?
	`, id)
	if err != nil {
		return fmt.Errorf("sqlite: delete collection %q: %w", id, err)
	}
	return nil
}
