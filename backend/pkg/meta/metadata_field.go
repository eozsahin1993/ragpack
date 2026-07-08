package meta

import (
	"context"
	"time"
)

type MetadataField struct {
	ID           string    `db:"id"            json:"id"`
	CollectionID string    `db:"collection_id" json:"collection_id"`
	Name         string    `db:"name"          json:"name"`
	Type         string    `db:"type"          json:"type"` // "str" | "num" | "arr"
	Slot         int       `db:"slot"          json:"slot"`
	CreatedAt    time.Time `db:"created_at"    json:"created_at"`
}

type MetadataFieldInput struct {
	Name string
	Type string // "str" | "num" | "arr"
}

type MetadataFieldStore interface {
	// RegisterMetadataFields registers multiple fields in a single transaction,
	// auto-assigning the next available slot per type.
	RegisterMetadataFields(ctx context.Context, collectionID string, fields []MetadataFieldInput) ([]MetadataField, error)
	ListMetadataFields(ctx context.Context, collectionID string) ([]MetadataField, error)
	GetMetadataFieldByName(ctx context.Context, collectionID, name string) (MetadataField, error)
	DeleteMetadataField(ctx context.Context, collectionID, name string) (MetadataField, error)
}
