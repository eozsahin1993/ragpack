package ingester

import (
	"time"

	"ragpack/pkg/embedder"
	"ragpack/pkg/meta"
	"ragpack/pkg/telemetry/ingestion"
)

// This file is all of the ingest pipeline's telemetry assembly — worker.go
// keeps only stage timers feeding ingestStats.

// ingestStats accumulates stage timings/usage during process().
type ingestStats struct {
	fetchMs     int64
	loopMs      int64
	waitMs      int64
	embedMs     int64
	insertMs    int64
	optimizeMs  int64
	embedTokens int
}

// recordIngestion emits the ingestion_events row for one process() attempt.
// Cost semantics: local model = confirmed $0; unknown hosted model = nil
// (unpriced), never a silent zero.
func (wp *WorkerPool) recordIngestion(job meta.Job, docID string, collection meta.Collection, chunkCount int, processErr error, totalMs int64, s ingestStats) {
	status := "complete"
	var errStr *string
	if processErr != nil {
		status = "failed"
		msg := processErr.Error()
		errStr = &msg
		chunkCount = 0
	}

	var tokens *int64
	if s.embedTokens > 0 {
		v := int64(s.embedTokens)
		tokens = &v
	}
	var cost *float64
	if wp.registry.IsLocal(collection.EmbedModel) {
		zero := 0.0
		cost = &zero
	} else if usd, priced := embedder.Cost(collection.EmbedModel, embedder.Usage{TotalTokens: s.embedTokens}); priced {
		cost = &usd
	}

	// parse/chunk aren't separately timeable (fused streaming loop) — derived
	// from loop wall time minus the three timed calls inside flush().
	parseChunkMs := s.loopMs - s.waitMs - s.embedMs - s.insertMs
	if parseChunkMs < 0 {
		parseChunkMs = 0
	}

	wp.telemetry.Record(&ingestion.Event{
		OccurredAt:      time.Now().UTC(),
		JobID:           job.ID,
		DocumentID:      docID,
		CollectionID:    collection.ID,
		CollectionSlug:  collection.Slug,
		FileUri:         job.FileUri,
		MimeType:        job.MimeType,
		Intent:          string(job.Intent),
		Status:          status,
		Error:           errStr,
		ChunkCount:      chunkCount,
		FetchMs:         s.fetchMs,
		ParseChunkMs:    parseChunkMs,
		RateLimitWaitMs: s.waitMs,
		EmbedMs:         s.embedMs,
		InsertMs:        s.insertMs,
		OptimizeIndexMs: s.optimizeMs,
		TotalMs:         totalMs,
		EmbedModel:      collection.EmbedModel,
		EmbedTokens:     tokens,
		EmbedCostUSD:    cost,
	})
}
