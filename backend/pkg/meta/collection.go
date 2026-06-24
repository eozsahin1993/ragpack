package meta

import (
	"context"
	"time"
)

type Collection struct {
	ID         string    `db:"id"`
	Name       string    `db:"name"`
	Slug       string    `db:"slug"`
	TableName  string    `db:"table_name"`
	EmbedModel string    `db:"embed_model"`
	VectorDim  int       `db:"vector_dim"`
	CreatedAt  time.Time `db:"created_at"`

	ChunkStrategy *string `db:"chunk_strategy"`
	ChunkSize     *int    `db:"chunk_size"`
	ChunkOverlap  *int    `db:"chunk_overlap"`
}

// CreateCollectionInput carries all parameters for creating a new collection.
// Chunk fields are optional; nil means use the server defaults.
type CreateCollectionInput struct {
	Name          string
	EmbedModel    string
	VectorDim     int
	ChunkStrategy *string
	ChunkSize     *int
	ChunkOverlap  *int
}

type CollectionReader interface {
	GetCollectionByID(ctx context.Context, id string) (Collection, error)
	GetCollectionBySlug(ctx context.Context, slug string) (Collection, error)
	ListCollections(ctx context.Context, limit, offset int) ([]Collection, error)
	CountCollections(ctx context.Context) (int, error)
}

type CollectionWriter interface {
	CreateCollection(ctx context.Context, input CreateCollectionInput) (Collection, error)
	UpdateCollectionName(ctx context.Context, id, name string) (Collection, error)
	DeleteCollection(ctx context.Context, id string) error
}

type CollectionStore interface {
	CollectionReader
	CollectionWriter
}
