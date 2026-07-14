package meta

import (
	"context"
	"errors"
	"time"
)

// ErrDocumentAlreadyIngesting: ResetDocument's conditional UPDATE (WHERE status != 'ingesting') affected zero
// rows — another job (e.g. a concurrent manual + auto-refresh) already won the atomic claim on this document.
var ErrDocumentAlreadyIngesting = errors.New("document is already being ingested")

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
	Name         *string        `db:"name"          json:"name,omitempty"   sort:"true"`
	ExternalId   *string        `db:"external_id"   json:"external_id,omitempty"`
	ExtraJSON    *string        `db:"extra_json"    json:"extra_json,omitempty"`
	ChunkCount   int            `db:"chunk_count"   json:"chunk_count"`
	Status       DocumentStatus `db:"status"        json:"status"           sort:"true"`
	Error        *string        `db:"error"         json:"error,omitempty"`
	CreatedAt    time.Time      `db:"created_at"    json:"created_at"       sort:"true"`
	UpdatedAt    time.Time      `db:"updated_at"    json:"updated_at"       sort:"default"`

	// LastETag is nil if never checked, or if the source sends no ETag.
	LastETag *string `db:"last_etag" json:"last_etag,omitempty"`
}

// DocumentPatch: nil fields are untouched, except Error==nil is ambiguous with "clear it", so ClearError disambiguates.
type DocumentPatch struct {
	Name      *string
	ExtraJSON *string

	Status     *DocumentStatus
	ChunkCount *int
	Error      *string
	ClearError bool
	LastETag   *string
}

// DocumentFilter holds optional predicates for listing/counting documents.
// Nil fields are ignored (no filtering on that column).
type DocumentFilter struct {
	CollectionID *string
	Status       *DocumentStatus
}

// DocumentSortSpec is derived from the `sort` tags on the Document struct.
var DocumentSortSpec = Sortable[Document]()

type DocumentSort struct {
	Field string
	Dir   SortDir
}

type DocumentReader interface {
	GetDocument(ctx context.Context, id string) (Document, error)
	ListDocuments(ctx context.Context, filter DocumentFilter, sort DocumentSort, limit, offset int) ([]Document, error)
	CountDocuments(ctx context.Context, filter DocumentFilter) (int, error)
	FindDocumentByFileUri(ctx context.Context, collectionID, fileUri string) (*Document, error)
}

type DocumentWriter interface {
	CreateDocument(ctx context.Context, collectionID, jobID, fileUri, mimeType string, extraJSON *string) (Document, error)
	ResetDocument(ctx context.Context, docID, newJobID string) (Document, error)
	UpdateDocument(ctx context.Context, id string, patch DocumentPatch) error
	DeleteDocument(ctx context.Context, id string) error
	DeleteDocumentsByCollection(ctx context.Context, collectionID string) error
}

type DocumentStore interface {
	DocumentReader
	DocumentWriter
}
