package sqlite

import (
	"context"
	"fmt"
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
		       chunk_strategy, chunk_size, chunk_overlap
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
		       chunk_strategy, chunk_size, chunk_overlap
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
		       chunk_strategy, chunk_size, chunk_overlap
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
