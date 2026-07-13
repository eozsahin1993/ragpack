package query

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"ragpack/pkg/api/middleware"
	"ragpack/pkg/db"
	"ragpack/pkg/llm"
	"ragpack/pkg/meta"
	tquery "ragpack/pkg/telemetry/query"
	"ragpack/pkg/telemetry/trace"
)

// This file is all of Query/Rag's telemetry assembly — handler.go keeps only
// one-line touchpoints at stage boundaries.

// newQueryEvent seeds the telemetry row shared by Query and Rag. Events are
// only born after collection resolution + validation — earlier failures are
// client errors with no collection dimension to slice by.
func (h *Handler) newQueryEvent(c *fiber.Ctx, collection meta.Collection, endpoint, queryText string, topK int, vectorOnly bool, hs db.HybridSettings, filtersJSON []byte) tquery.Event {
	ev := tquery.Event{
		OccurredAt:       time.Now().UTC(),
		CollectionID:     collection.ID,
		CollectionSlug:   collection.Slug,
		Origin:           "admin",
		Endpoint:         endpoint,
		QueryText:        &queryText,
		TopK:             topK,
		VectorSearchOnly: vectorOnly,
		Hybrid: tquery.HybridSettings{
			FullTextWeight: float64(hs.FullTextWeight),
			SemanticWeight: float64(hs.SemanticWeight),
			RRFK:           float64(hs.RRFK),
			FullTextLimit:  int32(hs.FullTextLimit),
		},
		EmbedModel: collection.EmbedModel,
		Status:     "complete",
	}
	if key, ok := c.Locals(middleware.LocalAPIKey).(meta.APIKey); ok {
		ev.APIKeyID = &key.ID
		ev.Origin = "public"
	}
	if len(filtersJSON) > 0 {
		f := string(filtersJSON)
		ev.FiltersJSON = &f
	}
	return ev
}

func (h *Handler) recordFailure(ev *tquery.Event, start time.Time, msg string) {
	ev.Status = "failed"
	ev.Error = &msg
	ev.TotalMs = time.Since(start).Milliseconds()
	h.telemetry.Record(ev)
}

// recordLLMFailure keeps the successful retrieval's metrics and trace (chunks
// + prompt, no answer) on the failed event — those tokens were really spent.
func (h *Handler) recordLLMFailure(ev *tquery.Event, start time.Time, results []db.ChunkQueryResult, formattedPrompt, msg string) {
	ev.ResultCount = len(results)
	ev.Results = resultStats(results)
	h.recordFailure(ev, start, msg)
	h.telemetry.Record(&trace.Event{
		EventID:         ev.EventID,
		OccurredAt:      ev.OccurredAt,
		Chunks:          traceChunks(results),
		FormattedPrompt: &formattedPrompt,
	})
}

// finishQueryEvent records the successful event plus its trace row.
func (h *Handler) finishQueryEvent(ev *tquery.Event, start time.Time, results []db.ChunkQueryResult, formattedPrompt, answer *string) {
	ev.ResultCount = len(results)
	ev.Results = resultStats(results)
	ev.TotalMs = time.Since(start).Milliseconds()
	h.telemetry.Record(ev)
	h.telemetry.Record(&trace.Event{
		EventID:         ev.EventID,
		OccurredAt:      ev.OccurredAt,
		Chunks:          traceChunks(results),
		FormattedPrompt: formattedPrompt,
		Answer:          answer,
	})
}

// setLLMUsage fills token/cost fields. Local model = confirmed $0; unknown
// hosted model = nil (unpriced), never a silent zero.
func (h *Handler) setLLMUsage(ev *tquery.Event, model string, usage llm.Usage) {
	if usage.InputTokens > 0 || usage.OutputTokens > 0 {
		in, out := int64(usage.InputTokens), int64(usage.OutputTokens)
		ev.LLMInputTokens = &in
		ev.LLMOutputTokens = &out
	}
	if h.llmRegistry.IsLocal(model) {
		zero := 0.0
		ev.LLMCostUSD = &zero
	} else if usd, priced := llm.Cost(model, usage); priced {
		ev.LLMCostUSD = &usd
	}
}

func resultStats(results []db.ChunkQueryResult) []tquery.ResultStat {
	stats := make([]tquery.ResultStat, len(results))
	for i, r := range results {
		stats[i] = tquery.ResultStat{
			SourceName: r.SourceName,
			Similarity: float64(r.VectorSimilarity),
			BM25Score:  float64(r.KeywordBM25Score),
			RRFScore:   float64(r.RRFScore),
		}
	}
	return stats
}

func traceChunks(results []db.ChunkQueryResult) []trace.Chunk {
	chunks := make([]trace.Chunk, len(results))
	for i, r := range results {
		text := ""
		if r.ChunkText != nil {
			text = *r.ChunkText
		}
		chunks[i] = trace.Chunk{
			SourceName:  r.SourceName,
			ChunkHeader: r.ChunkHeader,
			ChunkText:   text,
			Similarity:  float64(r.VectorSimilarity),
			BM25Score:   float64(r.KeywordBM25Score),
		}
	}
	return chunks
}
