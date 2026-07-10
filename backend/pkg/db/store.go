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
	// Populated when this result came from the vector channel.
	VectorDistance   float32 `json:"vector_distance"`
	VectorSimilarity float32 `json:"vector_similarity"` // 0-100, cosine-based
	// Raw BM25 score; populated when from the keyword channel. Unnormalized.
	KeywordBM25Score float32 `json:"keyword_bm25_score,omitempty"`
	// RRFScoreNormalized is RRFScore mapped to 0-100; RRFScore itself is only comparable within one query's own weights/k.
	RRFScoreNormalized float32        `json:"rrf_score_normalized,omitempty"`
	RRFScore           float32        `json:"rrf_score,omitempty"`
	Metadata           map[string]any `json:"metadata,omitempty"`
}

// HybridSettings configures the weighted RRF merge.
type HybridSettings struct {
	FullTextWeight float32
	SemanticWeight float32
	RRFK           float32
	FullTextLimit  int // caps FTS candidates; FTS search has no native limit
}

func DefaultHybridSettings() HybridSettings {
	return HybridSettings{
		FullTextWeight: 0.3,
		SemanticWeight: 0.7,
		RRFK:           60,
		FullTextLimit:  200,
	}
}

type VectorDb interface {
	Connect(ctx context.Context, connectionUrl string) error
	CreateTable(ctx context.Context, name string, collectionID string, vectorDim int) error
	DropTable(ctx context.Context, name string) error
	InsertBatch(ctx context.Context, tableName string, records []ChunkDbRecord) error
	// keywordQuery empty = vector-only; non-empty fuses in an FTS pass (hybrid search).
	QuerySimilarVectors(ctx context.Context, tableName string, vector []float32, topK int, filter string, keywordQuery string, hybrid HybridSettings) ([]ChunkQueryResult, error)
	DeleteChunksByDocument(ctx context.Context, tableName, documentID string) error
	ListChunksByDocument(ctx context.Context, tableName, documentID string) ([]ChunkDbRecord, error)
	UpdateChunks(ctx context.Context, tableName, documentID string, patch ChunkPatch) error
	OptimizeIndex(ctx context.Context, tableName string) error
	// fieldType must be "str", "num", or "arr".
	CreateMetadataIndex(ctx context.Context, tableName, colName, fieldType string) error
	DropMetadataIndex(ctx context.Context, tableName, indexName string) error
	// NullMetadataSlot sets every row's value in colName to NULL (field deletion).
	NullMetadataSlot(ctx context.Context, tableName, colName string) error
}
