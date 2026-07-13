// Package ingestion is the ingestion_events telemetry table: one row per
// document per job attempt.
package ingestion

import (
	"time"

	"ragpack/pkg/telemetry/schema"
)

type Event struct {
	EventID         string
	OccurredAt      time.Time
	JobID           string
	DocumentID      string
	CollectionID    string
	CollectionSlug  string
	FileUri         string
	MimeType        string
	Intent          string // ingest | refresh
	Status          string // complete | failed
	Error           *string
	ChunkCount      int
	FetchMs         int64
	ParseChunkMs    int64 // derived: loop wall time minus wait/embed/insert
	RateLimitWaitMs int64
	EmbedMs         int64
	InsertMs        int64
	OptimizeIndexMs int64
	TotalMs         int64
	EmbedModel      string
	EmbedTokens     *int64   // nil = provider reported no usage
	EmbedCostUSD    *float64 // nil = unpriced model; 0 = confirmed local/free
}

// Prepare stamps EventID/OccurredAt if unset. Ingestion events have no drop
// condition — redact only affects query text/traces.
func (e *Event) Prepare(redact bool) bool {
	schema.Stamp(&e.EventID, &e.OccurredAt)
	return true
}
