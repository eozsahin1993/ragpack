package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"ragpack/pkg/api"
	"ragpack/pkg/api/validate"
	"ragpack/pkg/chunker"
	"ragpack/pkg/config"
	"ragpack/pkg/db"
	lancedbpkg "ragpack/pkg/db/lancedb"
	"ragpack/pkg/embedder"
	"ragpack/pkg/ingester"
	"ragpack/pkg/llm"
	"ragpack/pkg/meta"
	sqlitemeta "ragpack/pkg/meta/sqlite"
)

func main() {
	cfg := config.Load()

	ms, err := sqlitemeta.New(cfg.SqlitePath)
	if err != nil {
		log.Fatalf("meta store: %v", err)
	}
	defer ms.Close()
	bootstrapAPIKey(context.Background(), ms, cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	vec := migrateVectorDB(ctx, cfg, ms)
	registry := embedder.NewRegistryFromConfig(ctx, cfg)
	llmRegistry := llm.NewRegistryFromConfig(cfg)
	ing := startIngester(ctx, cfg, ms, vec, registry)

	publicApp := createApp(cfg.MaxUploadSizeMB)
	api.RegisterPublic(publicApp, ms, vec, registry, llmRegistry, ing, cfg.DefaultPromptSlug, cfg.MaxUploadSizeMB)

	adminApp := createApp(cfg.MaxUploadSizeMB)
	api.RegisterAdmin(adminApp, ms, vec, registry, llmRegistry, ing, cfg.DefaultPromptSlug, cfg.MaxUploadSizeMB)

	go handleShutdown(cancel, ing, publicApp, adminApp)

	go mustListen(adminApp, cfg.AdminPort, "admin (internal)")
	mustListen(publicApp, cfg.Port, "public API")
}

func migrateVectorDB(ctx context.Context, cfg config.Config, ms meta.MetaStore) db.VectorDb {
	vec := lancedbpkg.New()
	if err := vec.Connect(ctx, cfg.LanceDBPath); err != nil {
		log.Fatalf("vector store: %v", err)
	}
	collections, err := ms.ListAllCollections(ctx)
	if err != nil {
		log.Fatalf("list collections for migration: %v", err)
	}
	if err := vec.MigrateAll(ctx, collections); err != nil {
		log.Fatalf("lancedb migrations: %v", err)
	}
	return vec
}

func startIngester(ctx context.Context, cfg config.Config, ms meta.MetaStore, vec db.VectorDb, registry *embedder.Registry) ingester.Ingester {
	chunkCfg := chunker.Config{
		ChunkSize: cfg.Ingester.ChunkSize,
		Overlap:   cfg.Ingester.ChunkOverlap,
		Strategy:  cfg.Ingester.ChunkStrategy,
	}
	ing := ingester.New(ms, vec, registry, cfg.Ingester.WorkerCount, cfg.Ingester.EmbedRateLimit, chunkCfg)
	ing.Start(ctx, cfg.Ingester.WorkerCount)
	return ing
}

func createApp(maxUploadSize int) *fiber.App {
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

func handleShutdown(cancel context.CancelFunc, ing ingester.Ingester, apps ...*fiber.App) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")
	cancel()
	ing.Stop()
	for _, app := range apps {
		app.Shutdown()
	}
}

func mustListen(app *fiber.App, port, name string) {
	log.Printf("RagPack %s server starting on :%s", name, port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("%s server: %v", name, err)
	}
}
