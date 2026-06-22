package db

import "time"
import "context"

type ChunkMetadata struct {
	ExternalId string `json:"external_id"`
	FileUri    string `json:"file_uri"`
	SourceName string `json:"source_name"`
	MimeType   string `json:"mime_type"`
	SourceType string `json:"source_type"`
	RawJson    string `json:"raw_json"`
}

type ChunkDbRecord struct {
	ID         string          `json:"id"`
	Chunk      []byte          `json:"chunk"`
	ChunkHash  string          `json:"chunk_hash"`
	ChunkIndex int             `json:"chunk_index"`
	Vector     []float32       `json:"vector"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
	Metadata   ChunkDbMetadata `json:"metadata"`
}

type VectorDb interface {
	connect(ctx context.Context, connectionUrl string) error
	createTable(ctx context.Context, name string) error
	insertRecord(ctx context.Context, tableName string, record ChunkDbRecord) error
	querySimilarVectors(ctx context.Context, tableName string, vector []float32, topK int) ([]ChunkDbRecord, error)
}
