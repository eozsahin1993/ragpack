// Package app assembles the ingester and both Fiber apps (public + admin) —
// the single source of truth for how ragpack's pieces wire together at
// startup, so cmd/main.go and integration tests exercise the exact same
// sequence instead of a hand-duplicated copy that can drift out of sync.
package app

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"

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

	ing := ingester.New(d.Meta, d.Vector, d.Embedders, d.Config.Ingester.WorkerCount, d.Config.Ingester.EmbedRateLimit, chunkCfg, rec)
	ing.Start(ctx, d.Config.Ingester.WorkerCount)

	publicApp := newFiberApp(d.Config.MaxUploadSizeMB)
	api.RegisterPublic(publicApp, d.Meta, d.Vector, d.Embedders, d.LLMs, ing, d.Config.DefaultPromptSlug, d.Config.MaxUploadSizeMB, rec)

	adminApp := newFiberApp(d.Config.MaxUploadSizeMB)
	api.RegisterAdmin(adminApp, d.Meta, d.Vector, d.Embedders, d.LLMs, ing, d.Config.DefaultPromptSlug, d.Config.MaxUploadSizeMB, rec)

	return &App{Public: publicApp, Admin: adminApp, Ingester: ing, Telemetry: rec}
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
