package sqlite

import (
	"context"
	"fmt"
	"time"

	"ragpack/pkg/meta"
)

func (s *MetaStore) CreateCollection(ctx context.Context, name, embedModel string, vectorDim int) (meta.Collection, error) {
	id, tableName := newTableName(name)
	c := meta.Collection{
		ID:         id,
		Name:       name,
		Slug:       slugify(name + "_" + embedModel),
		TableName:  tableName,
		EmbedModel: embedModel,
		VectorDim:  vectorDim,
		CreatedAt:  time.Now().UTC(),
	}

	_, err := s.db.NamedExecContext(ctx, `
		INSERT INTO collections (id, name, slug, table_name, embed_model, vector_dim, created_at)
		VALUES (:id, :name, :slug, :table_name, :embed_model, :vector_dim, :created_at)
	`, c)
	if err != nil {
		return meta.Collection{}, fmt.Errorf("sqlite: create collection %q: %w", name, err)
	}
	return c, nil
}

func (s *MetaStore) GetCollectionBySlug(ctx context.Context, slug string) (meta.Collection, error) {
	var c meta.Collection
	err := s.db.GetContext(ctx, &c, `
		SELECT id, name, slug, table_name, embed_model, vector_dim, created_at
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
		SELECT id, name, slug, table_name, embed_model, vector_dim, created_at
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
		SELECT id, name, slug, table_name, embed_model, vector_dim, created_at
		FROM collections
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("sqlite: list collections: %w", err)
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

func (s *MetaStore) UpdateCollectionName(ctx context.Context, id, name string) (meta.Collection, error) {
	var embedModel string
	err := s.db.GetContext(ctx, &embedModel, `SELECT embed_model FROM collections WHERE id = ?`, id)
	if err != nil {
		return meta.Collection{}, fmt.Errorf("sqlite: update collection name %q: %w", id, err)
	}
	newSlug := slugify(name + "_" + embedModel)

	_, err = s.db.ExecContext(ctx, `
		UPDATE collections SET name = ?, slug = ? WHERE id = ?
	`, name, newSlug, id)
	if err != nil {
		return meta.Collection{}, fmt.Errorf("sqlite: update collection name %q: %w", id, err)
	}
	return s.GetCollectionByID(ctx, id)
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
