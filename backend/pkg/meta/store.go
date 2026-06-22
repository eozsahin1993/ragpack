package meta

import (
	"context"
	"time"
)

type Collection struct {
	ID         string    `db:"id"`
	Name       string    `db:"name"`
	TableName  string    `db:"table_name"`
	EmbedModel string    `db:"embed_model"`
	VectorDim  int       `db:"vector_dim"`
	CreatedAt  time.Time `db:"created_at"`
}

type CollectionReader interface {
	GetCollectionByName(ctx context.Context, name string) (Collection, error)
	GetCollectionByID(ctx context.Context, id string) (Collection, error)
	ListCollections(ctx context.Context) ([]Collection, error)
}

type CollectionWriter interface {
	CreateCollection(ctx context.Context, name, embedModel string, vectorDim int) (Collection, error)
	DeleteCollection(ctx context.Context, name string) error
}

type CollectionStore interface {
	CollectionReader
	CollectionWriter
}

type MetaStore interface {
	CollectionStore
	Close() error
}
