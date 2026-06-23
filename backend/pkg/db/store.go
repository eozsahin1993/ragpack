package db

import (
	"context"
	"time"
)

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
	ChunkText  *string `json:"chunk_text"`
	ExternalId *string `json:"external_id"`
	ExtraJSON  *string `json:"extra_json"`
}

type ChunkQueryResult struct {
	ChunkDbRecord
	Distance   float32 `json:"distance"`
	Similarity float32 `json:"similarity"` // 0-100, cosine-based
}

type VectorDb interface {
	Connect(ctx context.Context, connectionUrl string) error
	CreateTable(ctx context.Context, name string, vectorDim int) error
	DropTable(ctx context.Context, name string) error
	InsertRecord(ctx context.Context, tableName string, record ChunkDbRecord) error
	QuerySimilarVectors(ctx context.Context, tableName string, vector []float32, topK int) ([]ChunkQueryResult, error)
	DeleteChunksByDocument(ctx context.Context, tableName, documentID string) error
}
