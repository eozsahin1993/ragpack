package sqlite

import (
	"context"
	"fmt"
	"time"

	"ragpack/backend/pkg/meta"
)

func (s *MetaStore) CreateCollection(ctx context.Context, name, embedModel string, vectorDim int) (meta.Collection, error) {
	id, tableName := newTableName(name)
	c := meta.Collection{
		ID:         id,
		Name:       name,
		TableName:  tableName,
		EmbedModel: embedModel,
		VectorDim:  vectorDim,
		CreatedAt:  time.Now().UTC(),
	}

	_, err := s.db.NamedExecContext(ctx, `
		INSERT INTO collections (id, name, table_name, embed_model, vector_dim, created_at)
		VALUES (:id, :name, :table_name, :embed_model, :vector_dim, :created_at)
	`, c)
	if err != nil {
		return meta.Collection{}, fmt.Errorf("sqlite: create collection %q: %w", name, err)
	}
	return c, nil
}

func (s *MetaStore) GetCollectionByName(ctx context.Context, name string) (meta.Collection, error) {
	var c meta.Collection
	err := s.db.GetContext(ctx, &c, `
		SELECT id, name, table_name, embed_model, vector_dim, created_at
		FROM collections
		WHERE name = ?
	`, name)
	if err != nil {
		return meta.Collection{}, fmt.Errorf("sqlite: get collection by name %q: %w", name, err)
	}
	return c, nil
}

func (s *MetaStore) GetCollectionByID(ctx context.Context, id string) (meta.Collection, error) {
	var c meta.Collection
	err := s.db.GetContext(ctx, &c, `
		SELECT id, name, table_name, embed_model, vector_dim, created_at
		FROM collections
		WHERE id = ?
	`, id)
	if err != nil {
		return meta.Collection{}, fmt.Errorf("sqlite: get collection by id %q: %w", id, err)
	}
	return c, nil
}

func (s *MetaStore) ListCollections(ctx context.Context) ([]meta.Collection, error) {
	var collections []meta.Collection
	err := s.db.SelectContext(ctx, &collections, `
		SELECT id, name, table_name, embed_model, vector_dim, created_at
		FROM collections
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("sqlite: list collections: %w", err)
	}
	return collections, nil
}

func (s *MetaStore) DeleteCollection(ctx context.Context, name string) error {
	_, err := s.db.ExecContext(ctx, `
		DELETE FROM collections
		WHERE name = ?
	`, name)
	if err != nil {
		return fmt.Errorf("sqlite: delete collection %q: %w", name, err)
	}
	return nil
}
