// Package query is the query_events telemetry table: one row per Query or
// Rag call.
package query

import (
	"time"

	"ragpack/pkg/telemetry/schema"
)

// HybridSettings mirrors db.HybridSettings for the struct column.
type HybridSettings struct {
	FullTextWeight float64
	SemanticWeight float64
	RRFK           float64
	FullTextLimit  int32
}

// ResultStat is one entry of the results column: scores only, no chunk text
// (that lives in the trace package).
type ResultStat struct {
	SourceName string
	Similarity float64
	BM25Score  float64
	RRFScore   float64
}

type Event struct {
	EventID          string
	OccurredAt       time.Time
	CollectionID     string
	CollectionSlug   string
	APIKeyID         *string // nil on the admin surface
	Origin           string  // public | admin
	Endpoint         string  // query | rag
	QueryText        *string
	TopK             int
	VectorSearchOnly bool
	Hybrid           HybridSettings
	FiltersJSON      *string
	EmbedModel       string
	EmbedQueryTokens *int64
	EmbedMs          int64
	VectorSearchMs   int64
	ResultCount      int
	Results          []ResultStat
	Status           string // complete | failed
	Error            *string
	TotalMs          int64

	// RAG-only; nil on plain queries.
	PromptSlug      *string
	LLMModel        *string
	LLMInputTokens  *int64
	LLMOutputTokens *int64
	LLMCostUSD      *float64
	LLMMs           *int64
}

// Prepare stamps EventID/OccurredAt if unset and blanks QueryText under
// redaction — unlike a trace, a redacted query event is still recorded (just
// without the raw question text).
func (e *Event) Prepare(redact bool) bool {
	schema.Stamp(&e.EventID, &e.OccurredAt)
	if redact {
		e.QueryText = nil
	}
	return true
}
