package db

import "time"
import "context"

type ChunkDbRecord struct {
	// Required
	ID         string    `json:"id"`
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

type VectorDb interface {
	Connect(ctx context.Context, connectionUrl string) error
	CreateTable(ctx context.Context, name string) error
	InsertRecord(ctx context.Context, tableName string, record ChunkDbRecord) error
	QuerySimilarVectors(ctx context.Context, tableName string, vector []float32, topK int) ([]ChunkDbRecord, error)
}
