// Package helpers assembles the real app (via pkg/app.New, the same
// sequence cmd/main.go runs at startup) and provides small HTTP/domain
// helpers for the integration suite (backend/test/integration). A real
// package rather than _test.go files, so it compiles like any other
// dependency but is only ever imported by _test.go files — nothing in the
// production build references it.
package helpers

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/gofiber/fiber/v2"

	"ragpack/pkg/app"
	"ragpack/pkg/config"
	lancedbpkg "ragpack/pkg/db/lancedb"
	"ragpack/pkg/embedder"
	"ragpack/pkg/llm"
	"ragpack/pkg/meta"
	sqlitemeta "ragpack/pkg/meta/sqlite"
	"ragpack/test/integration/mocks"
)

// NewFullTestApp assembles the real app via pkg/app.New with real
// SQLite/LanceDB backed by t.TempDir() (auto-removed on test completion,
// pass or fail — see testing.T.TempDir) and fake embedder/LLM registries
// swapped in. Returns both the app bundle (Admin + Public + Ingester) and
// the meta store, for tests that need to create API keys against the
// public (auth-required) surface.
func NewFullTestApp(t *testing.T) (*app.App, meta.MetaStore) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())

	ms, err := sqlitemeta.New(filepath.Join(t.TempDir(), "meta.db"))
	if err != nil {
		t.Fatalf("meta store: %v", err)
	}
	t.Cleanup(func() { ms.Close() })

	vec := lancedbpkg.New()
	if err := vec.Connect(ctx, filepath.Join(t.TempDir(), "lancedb")); err != nil {
		t.Fatalf("vector store: %v", err)
	}
	t.Cleanup(func() { vec.Close() })

	registry := embedder.NewRegistry()
	registry.Register(&mocks.Embedder{Dim: 64, ModelName: "mock-embed"})

	llmRegistry := llm.NewRegistry()
	llmRegistry.Register(mocks.LLM{})

	a := app.New(ctx, app.Deps{
		Meta:      ms,
		Vector:    vec,
		Embedders: registry,
		LLMs:      llmRegistry,
		Config: config.Config{
			Ingester: config.IngesterConfig{
				WorkerCount:    2,
				EmbedRateLimit: 100,
				ChunkSize:      500,
				ChunkOverlap:   50,
				ChunkStrategy:  "auto",
			},
			DefaultPromptSlug: "basic_rag",
			MaxUploadSizeMB:   25,
			// On by default in production (config.Load()); matching that
			// here means the integration suite actually exercises the
			// telemetry write path and the analytics engine/routes
			// (skipped entirely when disabled — see pkg/app.New), not just
			// the ingestion/query business logic around them.
			Telemetry: config.TelemetryConfig{
				Enabled:       true,
				Dir:           filepath.Join(t.TempDir(), "telemetry"),
				RetentionDays: 14,
				MaxSizeMB:     500,
				DuckDB: config.DuckDBConfig{
					MemoryLimit:         "256MB",
					MaxThreads:          2,
					QueryTimeoutSeconds: 10,
				},
			},
		},
	})
	// t.Cleanup runs LIFO: registering cancel after Stop means cancel fires
	// first, unblocking the worker goroutines' ctx.Done() select before
	// Stop's wg.Wait() blocks on them — Stop() itself never cancels
	// anything (see pkg/ingester.WorkerPool.Stop), it only waits, so
	// without this a background-context caller hangs forever.
	//
	// Same shutdown order as cmd/main.go: ingester stops (so its final
	// telemetry events still get flushed) before Telemetry.Close, then
	// Analytics.Close.
	t.Cleanup(func() { a.Analytics.Close() })
	t.Cleanup(func() { a.Telemetry.Close() })
	t.Cleanup(a.Ingester.Stop)
	t.Cleanup(cancel)

	return a, ms
}

// NewTestApp returns just the admin (no-auth) app — what most tests need,
// since they're exercising route/business logic, not the auth middleware
// itself (see auth_test.go for that).
func NewTestApp(t *testing.T) *fiber.App {
	t.Helper()
	a, _ := NewFullTestApp(t)
	return a.Admin
}
