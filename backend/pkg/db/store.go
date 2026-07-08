package db

import (
	"context"
	"fmt"
	"time"
)

// MetadataSlotColumn returns the LanceDB column name for a metadata slot.
// fieldType is one of "str", "num", "bool", "date", "arr"; slot is 1-indexed.
func MetadataSlotColumn(fieldType string, slot int) string {
	return fmt.Sprintf("metadata_%s_%d", fieldType, slot)
}

type ChunkDbRecord struct {
	// Required
	ID         string    `json:"id"`
	DocumentID string    `json:"document_id"`
	ChunkHash  string    `json:"chunk_hash"`
	ChunkIndex int       `json:"chunk_index"`
	Vector     []float32 `json:"vector"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	MimeType   string    `json:"mime_type"`
	FileUri    string    `json:"file_uri"`
	SourceName string    `json:"source_name"`

	// Optional
	ChunkText   *string `json:"chunk_text"`
	ChunkHeader *string `json:"chunk_header"`
	ExternalId  *string `json:"external_id"`
	ExtraJSON   *string `json:"extra_json"`

	// Metadata filter slots — index n = slot n+1 (e.g. MetadataStr[0] = metadata_str_1)
	MetadataStr  [20]*string  `json:"-"`
	MetadataNum  [10]*float64 `json:"-"`
	MetadataBool [10]*bool    `json:"-"`
	MetadataDate [10]*int64   `json:"-"` // Unix timestamp seconds
	MetadataArr  [10][]string `json:"-"` // nil slice = NULL
}

type ChunkQueryResult struct {
	ChunkDbRecord
	Distance   float32                `json:"distance"`
	Similarity float32                `json:"similarity"` // 0-100, cosine-based
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

type VectorDb interface {
	Connect(ctx context.Context, connectionUrl string) error
	CreateTable(ctx context.Context, name string, collectionID string, vectorDim int) error
	DropTable(ctx context.Context, name string) error
	InsertBatch(ctx context.Context, tableName string, records []ChunkDbRecord) error
	QuerySimilarVectors(ctx context.Context, tableName string, vector []float32, topK int, filter string) ([]ChunkQueryResult, error)
	DeleteChunksByDocument(ctx context.Context, tableName, documentID string) error
	ListChunksByDocument(ctx context.Context, tableName, documentID string) ([]ChunkDbRecord, error)
	UpdateChunksExtraJSON(ctx context.Context, tableName, documentID string, extraJSON *string) error
	OptimizeIndex(ctx context.Context, tableName string) error
	// CreateMetadataIndex creates a scalar index on a metadata slot column.
	// fieldType must be "str", "num", or "arr".
	CreateMetadataIndex(ctx context.Context, tableName, colName, fieldType string) error
	// DropMetadataIndex removes the named index from a table.
	DropMetadataIndex(ctx context.Context, tableName, indexName string) error
	// NullMetadataSlot sets every row's value in colName to NULL (used when deleting a metadata field).
	NullMetadataSlot(ctx context.Context, tableName, colName string) error
}
