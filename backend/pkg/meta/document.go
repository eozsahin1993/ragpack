package meta

import (
	"context"
	"time"
)

type DocumentStatus string

const (
	DocumentStatusIngesting DocumentStatus = "ingesting"
	DocumentStatusComplete  DocumentStatus = "complete"
	DocumentStatusFailed    DocumentStatus = "failed"
)

type Document struct {
	ID           string         `db:"id"            json:"id"`
	CollectionID string         `db:"collection_id" json:"collection_id"`
	JobID        string         `db:"job_id"        json:"job_id"`
	FileUri      string         `db:"file_uri"      json:"file_uri"`
	MimeType     string         `db:"mime_type"     json:"mime_type"`
	Name         *string        `db:"name"          json:"name,omitempty"`
	ExternalId   *string        `db:"external_id"   json:"external_id,omitempty"`
	ExtraJSON    *string        `db:"extra_json"    json:"extra_json,omitempty"`
	ChunkCount   int            `db:"chunk_count"   json:"chunk_count"`
	Status       DocumentStatus `db:"status"        json:"status"`
	Error        *string        `db:"error"         json:"error,omitempty"`
	CreatedAt    time.Time      `db:"created_at"    json:"created_at"`
	UpdatedAt    time.Time      `db:"updated_at"    json:"updated_at"`
}

// DocumentPatch holds optional fields for a partial document update.
// Only non-nil fields are applied.
type DocumentPatch struct {
	Name      *string
	ExtraJSON *string
}

type DocumentReader interface {
	GetDocument(ctx context.Context, id string) (Document, error)
	ListDocumentsByCollection(ctx context.Context, collectionID string, limit, offset int) ([]Document, error)
	CountDocumentsByCollection(ctx context.Context, collectionID string) (int, error)
	FindDocumentByFileUri(ctx context.Context, collectionID, fileUri string) (*Document, error)
}

type DocumentWriter interface {
	CreateDocument(ctx context.Context, collectionID, jobID, fileUri, mimeType string, extraJSON *string) (Document, error)
	ResetDocument(ctx context.Context, docID, newJobID string) (Document, error)
	UpdateDocument(ctx context.Context, id string, patch DocumentPatch) error
	UpdateDocumentStatus(ctx context.Context, id string, status DocumentStatus, chunkCount int, docError *string) error
	DeleteDocument(ctx context.Context, id string) error
	DeleteDocumentsByCollection(ctx context.Context, collectionID string) error
}

type DocumentStore interface {
	DocumentReader
	DocumentWriter
}
