// Package trace is the query_traces telemetry table: the heavy text sibling
// of a query.Event, joined by EventID.
package trace

import "time"

// Chunk is one retrieved chunk with its text, for the drill-down view.
type Chunk struct {
	SourceName  string
	ChunkHeader *string
	ChunkText   string
	Similarity  float64
	BM25Score   float64
}

type Event struct {
	EventID         string
	OccurredAt      time.Time
	Chunks          []Chunk
	FormattedPrompt *string // rag only
	Answer          *string // rag only
}

// Prepare drops the event under redaction or if EventID (a foreign key to
// query_events, never auto-generated) was left unset — an orphan trace is
// worse than a missing one.
func (e *Event) Prepare(redact bool) bool {
	if redact || e.EventID == "" {
		return false
	}
	if e.OccurredAt.IsZero() {
		e.OccurredAt = time.Now().UTC()
	}
	return true
}
