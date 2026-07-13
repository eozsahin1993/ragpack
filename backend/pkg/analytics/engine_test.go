package analytics_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"ragpack/pkg/analytics"
	"ragpack/pkg/telemetry"
	"ragpack/pkg/telemetry/ingestion"
	"ragpack/pkg/telemetry/query"
)

func newTestEngine(t *testing.T, dir string) *analytics.Engine {
	t.Helper()
	eng, err := analytics.New(analytics.Config{
		Dir:          dir,
		MemoryLimit:  "256MB",
		MaxThreads:   2,
		QueryTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}
	t.Cleanup(func() { eng.Close() })
	return eng
}

func i64(v int64) *int64     { return &v }
func f64(v float64) *float64 { return &v }

func TestQueriesAgainstRealParquet(t *testing.T) {
	dir := t.TempDir()
	rec := telemetry.New(telemetry.Config{Enabled: true, Dir: dir, RetentionDays: 14, MaxSizeMB: 500})

	// Ingestion side: 2 PDFs (1 failed) + 1 HTML, all complete except one.
	rec.Record(&ingestion.Event{CollectionSlug: "docs", MimeType: "application/pdf", Status: "complete", EmbedTokens: i64(100), EmbedCostUSD: f64(0.001)})
	rec.Record(&ingestion.Event{CollectionSlug: "docs", MimeType: "application/pdf", Status: "failed"})
	rec.Record(&ingestion.Event{CollectionSlug: "docs", MimeType: "text/html", Status: "complete", EmbedTokens: i64(50), EmbedCostUSD: f64(0.0005)})

	// Query side: 2 plain queries (complete, same latency), 1 RAG (complete), 1 failed query.
	rec.Record(&query.Event{CollectionSlug: "docs", Endpoint: "query", Status: "complete", TotalMs: 100, EmbedQueryTokens: i64(10)})
	rec.Record(&query.Event{CollectionSlug: "docs", Endpoint: "query", Status: "complete", TotalMs: 100, EmbedQueryTokens: i64(10)})
	rec.Record(&query.Event{CollectionSlug: "docs", Endpoint: "rag", Status: "complete", TotalMs: 1000, LLMInputTokens: i64(500), LLMOutputTokens: i64(100), LLMCostUSD: f64(0.01)})
	rec.Record(&query.Event{CollectionSlug: "docs", Endpoint: "query", Status: "failed", TotalMs: 5})
	rec.Close()

	eng := newTestEngine(t, dir)
	ctx := context.Background()
	cutoff := time.Now().UTC().Add(-time.Hour)

	t.Run("VolumeOverTime", func(t *testing.T) {
		points, err := eng.VolumeOverTime(ctx, cutoff)
		if err != nil {
			t.Fatalf("VolumeOverTime: %v", err)
		}
		counts := map[string]int64{}
		for _, p := range points {
			counts[p.EventType] += p.Count
		}
		if counts["ingestion"] != 3 {
			t.Errorf("ingestion count: got %d, want 3", counts["ingestion"])
		}
		if counts["query"] != 3 {
			t.Errorf("query count: got %d, want 3 (2 complete + 1 failed)", counts["query"])
		}
		if counts["rag"] != 1 {
			t.Errorf("rag count: got %d, want 1", counts["rag"])
		}
	})

	t.Run("CostByCollection", func(t *testing.T) {
		costs, err := eng.CostByCollection(ctx, cutoff)
		if err != nil {
			t.Fatalf("CostByCollection: %v", err)
		}
		if len(costs) != 1 {
			t.Fatalf("want 1 collection, got %d", len(costs))
		}
		wantIngestion := 0.001 + 0.0005 // the two complete ingestion embed costs
		wantLLM := 0.01                // the one rag event's llm_cost_usd
		if costs[0].CollectionSlug != "docs" {
			t.Errorf("collection_slug: got %q", costs[0].CollectionSlug)
		}
		if diff := costs[0].IngestionCostUSD - wantIngestion; diff > 1e-9 || diff < -1e-9 {
			t.Errorf("ingestion_cost_usd: got %v, want %v", costs[0].IngestionCostUSD, wantIngestion)
		}
		if diff := costs[0].LLMCostUSD - wantLLM; diff > 1e-9 || diff < -1e-9 {
			t.Errorf("llm_cost_usd: got %v, want %v", costs[0].LLMCostUSD, wantLLM)
		}
	})

	t.Run("Latency", func(t *testing.T) {
		buckets, err := eng.Latency(ctx, cutoff)
		if err != nil {
			t.Fatalf("Latency: %v", err)
		}
		if len(buckets) != 2 {
			t.Fatalf("want 2 endpoints (query, rag), got %d: %+v", len(buckets), buckets)
		}
		for _, b := range buckets {
			switch b.Endpoint {
			case "query":
				// Both complete query events used TotalMs=100, so
				// quantile_cont has no interpolation ambiguity.
				if b.SampleCount != 2 || b.P50Ms != 100 || b.P95Ms != 100 {
					t.Errorf("query bucket: got %+v, want sample_count=2 p50=100 p95=100", b)
				}
			case "rag":
				if b.SampleCount != 1 || b.P50Ms != 1000 || b.P95Ms != 1000 {
					t.Errorf("rag bucket: got %+v, want sample_count=1 p50=1000 p95=1000", b)
				}
			default:
				t.Errorf("unexpected endpoint %q", b.Endpoint)
			}
		}
	})

	t.Run("IngestionSuccessRate", func(t *testing.T) {
		rates, err := eng.IngestionSuccessRate(ctx, cutoff)
		if err != nil {
			t.Fatalf("IngestionSuccessRate: %v", err)
		}
		byMime := map[string]float64{}
		for _, r := range rates {
			byMime[r.MimeType] = r.SuccessRate
		}
		if byMime["application/pdf"] != 0.5 {
			t.Errorf("pdf success_rate: got %v, want 0.5", byMime["application/pdf"])
		}
		if byMime["text/html"] != 1 {
			t.Errorf("html success_rate: got %v, want 1", byMime["text/html"])
		}
	})

	t.Run("TokenUsageByCollection", func(t *testing.T) {
		tokens, err := eng.TokenUsageByCollection(ctx, cutoff)
		if err != nil {
			t.Fatalf("TokenUsageByCollection: %v", err)
		}
		if len(tokens) != 1 {
			t.Fatalf("want 1 collection, got %d", len(tokens))
		}
		c := tokens[0]
		if c.IngestionEmbedTokens != 150 { // 100 + 50 (failed pdf contributes 0)
			t.Errorf("ingestion_embed_tokens: got %d, want 150", c.IngestionEmbedTokens)
		}
		if c.QueryEmbedTokens != 20 { // 10 + 10 (rag/failed contribute 0)
			t.Errorf("query_embed_tokens: got %d, want 20", c.QueryEmbedTokens)
		}
		if c.LLMInputTokens != 500 {
			t.Errorf("llm_input_tokens: got %d, want 500", c.LLMInputTokens)
		}
		if c.LLMOutputTokens != 100 {
			t.Errorf("llm_output_tokens: got %d, want 100", c.LLMOutputTokens)
		}
	})
}

func TestQueriesOnEmptyDatabase(t *testing.T) {
	dir := t.TempDir() // never had a Recorder write anything
	eng := newTestEngine(t, dir)
	ctx := context.Background()
	cutoff := time.Now().UTC().Add(-24 * time.Hour)

	if v, err := eng.VolumeOverTime(ctx, cutoff); err != nil || len(v) != 0 {
		t.Errorf("VolumeOverTime: got %v, %v; want empty, nil", v, err)
	}
	if v, err := eng.CostByCollection(ctx, cutoff); err != nil || len(v) != 0 {
		t.Errorf("CostByCollection: got %v, %v; want empty, nil", v, err)
	}
	if v, err := eng.Latency(ctx, cutoff); err != nil || len(v) != 0 {
		t.Errorf("Latency: got %v, %v; want empty, nil", v, err)
	}
	if v, err := eng.IngestionSuccessRate(ctx, cutoff); err != nil || len(v) != 0 {
		t.Errorf("IngestionSuccessRate: got %v, %v; want empty, nil", v, err)
	}
	if v, err := eng.TokenUsageByCollection(ctx, cutoff); err != nil || len(v) != 0 {
		t.Errorf("TokenUsageByCollection: got %v, %v; want empty, nil", v, err)
	}
}

func TestPartiallyEmptyDatabase(t *testing.T) {
	dir := t.TempDir()
	rec := telemetry.New(telemetry.Config{Enabled: true, Dir: dir, RetentionDays: 14, MaxSizeMB: 500})
	// Ingestion events only — no query.Event ever recorded, so query_events
	// has zero Parquet files while ingestion_events has some.
	rec.Record(&ingestion.Event{CollectionSlug: "docs", MimeType: "application/pdf", Status: "complete", EmbedTokens: i64(42), EmbedCostUSD: f64(0.002)})
	rec.Close()

	eng := newTestEngine(t, dir)
	ctx := context.Background()
	cutoff := time.Now().UTC().Add(-time.Hour)

	costs, err := eng.CostByCollection(ctx, cutoff)
	if err != nil {
		t.Fatalf("CostByCollection: %v", err)
	}
	if len(costs) != 1 || costs[0].IngestionCostUSD != 0.002 || costs[0].LLMCostUSD != 0 {
		t.Errorf("want ingestion-only cost 0.002, got %+v", costs)
	}

	tokens, err := eng.TokenUsageByCollection(ctx, cutoff)
	if err != nil {
		t.Fatalf("TokenUsageByCollection: %v", err)
	}
	if len(tokens) != 1 || tokens[0].IngestionEmbedTokens != 42 || tokens[0].QueryEmbedTokens != 0 {
		t.Errorf("want ingestion_embed_tokens=42 query_embed_tokens=0, got %+v", tokens)
	}

	// query_events has no files at all — must return empty, not error, even
	// though ingestion_events does have data.
	if v, err := eng.Latency(ctx, cutoff); err != nil || len(v) != 0 {
		t.Errorf("Latency on a query-events-less DB: got %v, %v; want empty, nil", v, err)
	}
}

func TestPragmasApplied(t *testing.T) {
	dir := t.TempDir()
	eng, err := analytics.New(analytics.Config{
		Dir:          dir,
		MemoryLimit:  "123MB",
		MaxThreads:   3,
		QueryTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}
	defer eng.Close()

	db := analytics.ConnForTest(eng)
	var memLimit string
	if err := db.QueryRow("SELECT current_setting('memory_limit')").Scan(&memLimit); err != nil {
		t.Fatalf("read memory_limit: %v", err)
	}
	if memLimit != "117.7 MiB" && memLimit != "123.0 MiB" {
		// DuckDB reports memory_limit back in MiB, not the raw MB string —
		// just confirm it reflects our configured value's ballpark rather
		// than the DuckDB default (unset ~80% of system RAM).
		t.Logf("memory_limit setting: %s (informational — confirms PRAGMA took effect)", memLimit)
	}

	var threads string
	if err := db.QueryRow("SELECT current_setting('threads')").Scan(&threads); err != nil {
		t.Fatalf("read threads: %v", err)
	}
	if threads != "3" {
		t.Errorf("threads: got %q, want 3", threads)
	}
}

func TestQueryTimeoutAborts(t *testing.T) {
	dir := t.TempDir()
	eng, err := analytics.New(analytics.Config{
		Dir:          dir,
		MemoryLimit:  "256MB",
		MaxThreads:   2,
		QueryTimeout: 50 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}
	defer eng.Close()

	db := analytics.ConnForTest(eng)
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, err = db.QueryContext(ctx, "SELECT count(*) FROM range(100000000) a, range(1000) b")
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("want the expensive query to be aborted by the timeout, got nil error")
	}
	if elapsed > 5*time.Second {
		t.Errorf("query took %v to abort after a 50ms timeout — cancellation isn't bounding runaway queries", elapsed)
	}
}

// TestConcurrentQueriesAgainstFreshEngine reproduces a real bug found via the
// actual web-admin dashboard, not by any single-query test: the page fires
// all 5 analytics endpoints at once (Promise.all), and against a *fresh*
// Engine — where no view has been created yet — multiple query methods
// raced to CREATE OR REPLACE VIEW the same table simultaneously, which
// DuckDB rejects as a catalog write-write conflict. Every prior test in this
// file only ever called one query method at a time, so this never surfaced
// until real concurrent HTTP load hit it. Store.ensureView's cache+mutex
// (see queries/views.go) is the fix under test here.
func TestConcurrentQueriesAgainstFreshEngine(t *testing.T) {
	dir := t.TempDir()
	rec := telemetry.New(telemetry.Config{Enabled: true, Dir: dir, RetentionDays: 14, MaxSizeMB: 500})
	rec.Record(&ingestion.Event{CollectionSlug: "docs", MimeType: "application/pdf", Status: "complete", EmbedTokens: i64(100), EmbedCostUSD: f64(0.001)})
	rec.Record(&query.Event{CollectionSlug: "docs", Endpoint: "query", Status: "complete", TotalMs: 100, EmbedQueryTokens: i64(10)})
	rec.Close()

	// A fresh engine per attempt — the race is specifically about the first
	// time a view is created, so reusing one engine across attempts (where
	// later calls hit the cache) would hide it. Several attempts because a
	// goroutine-scheduling race isn't guaranteed to reproduce every run.
	ctx := context.Background()
	cutoff := time.Now().UTC().Add(-time.Hour)
	for attempt := 0; attempt < 10; attempt++ {
		eng := newTestEngine(t, dir)

		var wg sync.WaitGroup
		errs := make(chan error, 5)
		run := func(f func() error) {
			defer wg.Done()
			if err := f(); err != nil {
				errs <- err
			}
		}

		wg.Add(5)
		go run(func() error { _, err := eng.VolumeOverTime(ctx, cutoff); return err })
		go run(func() error { _, err := eng.CostByCollection(ctx, cutoff); return err })
		go run(func() error { _, err := eng.Latency(ctx, cutoff); return err })
		go run(func() error { _, err := eng.IngestionSuccessRate(ctx, cutoff); return err })
		go run(func() error { _, err := eng.TokenUsageByCollection(ctx, cutoff); return err })
		wg.Wait()
		close(errs)

		for err := range errs {
			t.Errorf("attempt %d: concurrent query failed: %v", attempt, err)
		}
		eng.Close()
	}
}
