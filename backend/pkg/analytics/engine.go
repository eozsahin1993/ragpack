package analytics

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/duckdb/duckdb-go/v2"

	"ragpack/pkg/analytics/queries"
)

type Config struct {
	Dir          string        // same dir telemetry.Config.Dir points at
	MemoryLimit  string        // PRAGMA memory_limit, e.g. "256MB"
	MaxThreads   int           // PRAGMA threads
	QueryTimeout time.Duration // context deadline applied to every query
}

// Engine is a thin wrapper over an in-memory DuckDB catalog whose "tables"
// are views over external Parquet files — DuckDB's own storage holds
// nothing. It owns connection lifecycle and the safety caps only; the five
// named queries (what to actually ask the data) live in
// pkg/analytics/queries, one file each, so it's clear at a glance which
// files are analytics setup and which are dashboard questions.
//
// A nil *Engine is valid for Close (mirrors telemetry.Recorder); unlike
// Recorder there is no honest no-op for a read query, so callers (pkg/app)
// skip constructing one at all when telemetry is disabled, and
// pkg/api/router skips registering the routes in that case.
type Engine struct {
	db  *sql.DB
	cfg Config
}

func New(cfg Config) (*Engine, error) {
	db, err := sql.Open("duckdb", "") // in-memory catalog; data lives in external Parquet
	if err != nil {
		return nil, fmt.Errorf("analytics: open duckdb: %w", err)
	}

	// Two connections: confirmed via spike that pooled duckdb-go connections
	// share one underlying in-memory database (a table/view created on one
	// connection is visible from another), and that memory_limit/threads are
	// GLOBAL DuckDB settings, not per-connection — so this is not "2 x the
	// configured budget." Whatever MemoryLimit/MaxThreads is set to remains
	// one shared ceiling that concurrent queries contend for, however many
	// connections are open; raising this just buys a little dashboard
	// parallelism (see pkg/api/analytics) without changing the cap itself.
	db.SetMaxOpenConns(2)

	init := []string{
		fmt.Sprintf("PRAGMA memory_limit='%s'", cfg.MemoryLimit),
		fmt.Sprintf("PRAGMA threads=%d", cfg.MaxThreads),
	}
	for _, stmt := range init {
		if _, err := db.ExecContext(context.Background(), stmt); err != nil {
			db.Close()
			return nil, fmt.Errorf("analytics: %s: %w", stmt, err)
		}
	}
	return &Engine{db: db, cfg: cfg}, nil
}

// Close is nil-safe, matching telemetry.Recorder's convention.
func (e *Engine) Close() error {
	if e == nil {
		return nil
	}
	return e.db.Close()
}

// withTimeout bounds one query. Cancellation is checked between DuckDB's
// internal execution steps, not preemptively mid-step — confirmed by a
// spike where a 50ms timeout took ~550ms to actually abort a deliberately
// expensive query. Fine against the ~10s production budget, but a query
// timeout here means "aborted within a bounded margin of the deadline," not
// "aborted the instant the deadline passed."
func (e *Engine) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, e.cfg.QueryTimeout)
}

func (e *Engine) VolumeOverTime(ctx context.Context, cutoff time.Time) ([]queries.VolumePoint, error) {
	ctx, cancel := e.withTimeout(ctx)
	defer cancel()
	return queries.VolumeOverTime(ctx, e.db, e.cfg.Dir, cutoff)
}

func (e *Engine) CostByCollection(ctx context.Context, cutoff time.Time) ([]queries.CollectionCost, error) {
	ctx, cancel := e.withTimeout(ctx)
	defer cancel()
	return queries.CostByCollection(ctx, e.db, e.cfg.Dir, cutoff)
}

func (e *Engine) Latency(ctx context.Context, cutoff time.Time) ([]queries.LatencyBucket, error) {
	ctx, cancel := e.withTimeout(ctx)
	defer cancel()
	return queries.Latency(ctx, e.db, e.cfg.Dir, cutoff)
}

func (e *Engine) IngestionFailureRate(ctx context.Context, cutoff time.Time) ([]queries.MimeFailureRate, error) {
	ctx, cancel := e.withTimeout(ctx)
	defer cancel()
	return queries.IngestionFailureRate(ctx, e.db, e.cfg.Dir, cutoff)
}

func (e *Engine) TokenUsageByCollection(ctx context.Context, cutoff time.Time) ([]queries.CollectionTokens, error) {
	ctx, cancel := e.withTimeout(ctx)
	defer cancel()
	return queries.TokenUsageByCollection(ctx, e.db, e.cfg.Dir, cutoff)
}
