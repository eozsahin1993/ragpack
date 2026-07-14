// Package app assembles the ingester and both Fiber apps (public + admin) —
// the single source of truth for how ragpack's pieces wire together at
// startup, so cmd/main.go and integration tests exercise the exact same
// sequence instead of a hand-duplicated copy that can drift out of sync.
package app

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"ragpack/pkg/analytics"
	"ragpack/pkg/api"
	"ragpack/pkg/api/validate"
	"ragpack/pkg/chunker"
	"ragpack/pkg/config"
	"ragpack/pkg/db"
	"ragpack/pkg/embedder"
	"ragpack/pkg/ingester"
	"ragpack/pkg/llm"
	"ragpack/pkg/meta"
	"ragpack/pkg/telemetry"
)

// App bundles the running pieces so callers can start/stop them uniformly.
type App struct {
	Public    *fiber.App
	Admin     *fiber.App
	Ingester  ingester.Ingester
	Telemetry *telemetry.Recorder
	Analytics *analytics.Engine // nil when telemetry is disabled or engine init failed
}

type Deps struct {
	Meta      meta.MetaStore
	Vector    db.VectorDb
	Embedders *embedder.Registry
	LLMs      *llm.Registry
	Config    config.Config
}

func New(ctx context.Context, d Deps) *App {
	chunkCfg := chunker.Config{
		ChunkSize: d.Config.Ingester.ChunkSize,
		Overlap:   d.Config.Ingester.ChunkOverlap,
		Strategy:  d.Config.Ingester.ChunkStrategy,
	}
	rec := telemetry.New(telemetry.Config{
		Enabled:       d.Config.Telemetry.Enabled,
		Dir:           d.Config.Telemetry.Dir,
		RetentionDays: d.Config.Telemetry.RetentionDays,
		MaxSizeMB:     d.Config.Telemetry.MaxSizeMB,
		RedactText:    d.Config.Telemetry.RedactText,
	})

	// No honest no-op for a read query the way Recorder has for a write (see
	// analytics.Engine's doc comment), so skip constructing the engine
	// entirely when telemetry is off rather than building one that would
	// always return empty results — RegisterAdmin skips mounting the routes
	// when eng is nil, so they 404 instead.
	var eng *analytics.Engine
	if d.Config.Telemetry.Enabled {
		var err error
		eng, err = analytics.New(analytics.Config{
			Dir:          d.Config.Telemetry.Dir,
			MemoryLimit:  d.Config.Telemetry.DuckDB.MemoryLimit,
			MaxThreads:   d.Config.Telemetry.DuckDB.MaxThreads,
			QueryTimeout: time.Duration(d.Config.Telemetry.DuckDB.QueryTimeoutSeconds) * time.Second,
		})
		if err != nil {
			log.Printf("analytics: duckdb engine init failed, analytics endpoints disabled: %v", err)
			eng = nil
		}
	}

	ing := ingester.New(d.Meta, d.Vector, d.Embedders, d.Config.Ingester.WorkerCount, d.Config.Ingester.EmbedRateLimit, chunkCfg, rec)
	ing.Start(ctx, d.Config.Ingester.WorkerCount)

	publicApp := newFiberApp(d.Config.MaxUploadSizeMB)
	api.RegisterPublic(publicApp, d.Meta, d.Vector, d.Embedders, d.LLMs, ing, d.Config.DefaultPromptSlug, d.Config.MaxUploadSizeMB, d.Config.MinCollectionRefreshSeconds, rec)

	adminApp := newFiberApp(d.Config.MaxUploadSizeMB)
	api.RegisterAdmin(adminApp, d.Meta, d.Vector, d.Embedders, d.LLMs, ing, d.Config.DefaultPromptSlug, d.Config.MaxUploadSizeMB, d.Config.MinCollectionRefreshSeconds, rec, eng)

	return &App{Public: publicApp, Admin: adminApp, Ingester: ing, Telemetry: rec, Analytics: eng}
}

func newFiberApp(maxUploadSize int) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      "RagPack Engine v1.0",
		BodyLimit:    maxUploadSize * 1024 * 1024,
		ErrorHandler: handleFiberError,
	})
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PATCH, DELETE, OPTIONS",
	}))
	return app
}

// handleFiberError skips validate.ErrResponseWritten (validate.Body/Query
// already wrote the real error response) so it isn't overwritten by Fiber's
// default handler; anything else falls through to that default.
func handleFiberError(c *fiber.Ctx, err error) error {
	if errors.Is(err, validate.ErrResponseWritten) {
		return nil
	}
	return fiber.DefaultErrorHandler(c, err)
}
