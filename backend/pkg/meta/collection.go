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
}

type CollectionReader interface {
	GetCollectionByID(ctx context.Context, id string) (Collection, error)
	GetCollectionBySlug(ctx context.Context, slug string) (Collection, error)
	ListCollections(ctx context.Context) ([]Collection, error)
}

type CollectionWriter interface {
	CreateCollection(ctx context.Context, name, embedModel string, vectorDim int) (Collection, error)
	UpdateCollectionName(ctx context.Context, id, name string) (Collection, error)
	DeleteCollection(ctx context.Context, id string) error
}

type CollectionStore interface {
	CollectionReader
	CollectionWriter
}
