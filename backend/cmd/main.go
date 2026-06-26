package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"ragpack/pkg/api"
	"ragpack/pkg/chunker"
	"ragpack/pkg/config"
	lancedbpkg "ragpack/pkg/db/lancedb"
	"ragpack/pkg/embedder"
	"ragpack/pkg/ingester"
	"ragpack/pkg/llm"
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

	vec := lancedbpkg.New()
	if err := vec.Connect(context.Background(), cfg.LanceDBPath); err != nil {
		log.Fatalf("vector store: %v", err)
	}

	collections, err := ms.ListAllCollections(context.Background())
	if err != nil {
		log.Fatalf("list collections for migration: %v", err)
	}
	if err := vec.MigrateAll(context.Background(), collections); err != nil {
		log.Fatalf("lancedb migrations: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	registry := embedder.NewRegistryFromConfig(ctx, cfg)
	llmRegistry := llm.NewRegistryFromConfig(cfg)

	chunkCfg := chunker.Config{
		ChunkSize: cfg.Ingester.ChunkSize,
		Overlap:   cfg.Ingester.ChunkOverlap,
		Strategy:  cfg.Ingester.ChunkStrategy,
	}
	ing := ingester.New(ms, vec, registry, cfg.Ingester.WorkerCount, cfg.Ingester.EmbedRateLimit, chunkCfg)
	ing.Start(ctx, cfg.Ingester.WorkerCount)

	newApp := func() *fiber.App {
		app := fiber.New(fiber.Config{AppName: "RagPack Engine v1.0"})
		app.Use(logger.New())
		app.Use(cors.New(cors.Config{
			AllowOrigins: "*",
			AllowHeaders: "Origin, Content-Type, Accept, Authorization",
			AllowMethods: "GET, POST, PATCH, DELETE, OPTIONS",
		}))
		return app
	}

	publicApp := newApp()
	api.RegisterPublic(publicApp, ms, vec, registry, llmRegistry, ing, cfg.DefaultPromptSlug)

	adminApp := newApp()
	api.RegisterAdmin(adminApp, ms, vec, registry, llmRegistry, ing, cfg.DefaultPromptSlug)

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-quit
		log.Println("shutting down...")
		cancel()
		ing.Stop()
		publicApp.Shutdown()
		adminApp.Shutdown()
	}()

	go func() {
		log.Printf("RagPack admin server starting on :%s (internal only)", cfg.AdminPort)
		if err := adminApp.Listen(fmt.Sprintf(":%s", cfg.AdminPort)); err != nil {
			log.Fatalf("admin server: %v", err)
		}
	}()

	log.Printf("RagPack public server starting on :%s", cfg.Port)
	if err := publicApp.Listen(fmt.Sprintf(":%s", cfg.Port)); err != nil {
		log.Fatalf("server: %v", err)
	}
}


