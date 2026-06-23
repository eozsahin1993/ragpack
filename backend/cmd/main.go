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
	"ragpack/pkg/config"
	lancedbpkg "ragpack/pkg/db/lancedb"
	"ragpack/pkg/embedder"
	"ragpack/pkg/ingester"
	sqlitemeta "ragpack/pkg/meta/sqlite"
)

func main() {
	cfg := config.Load()

	ms, err := sqlitemeta.New(cfg.SqlitePath)
	if err != nil {
		log.Fatalf("meta store: %v", err)
	}
	defer ms.Close()

	vec := lancedbpkg.New()
	if err := vec.Connect(context.Background(), cfg.LanceDBPath); err != nil {
		log.Fatalf("vector store: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	registry := embedder.NewRegistryFromConfig(ctx, cfg)

	ing := ingester.New(ms, vec, registry, cfg.Ingester.WorkerCount, cfg.Ingester.EmbedRateLimit)
	ing.Start(ctx, cfg.Ingester.WorkerCount)

	app := fiber.New(fiber.Config{
		AppName: "RagPack Engine v1.0",
	})
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
		AllowMethods: "GET, POST, PATCH, DELETE, OPTIONS",
	}))

	api.Register(app, ms, vec, registry, ing)

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-quit
		log.Println("shutting down...")
		cancel()
		ing.Stop()
		app.Shutdown()
	}()

	log.Printf("RagPack server starting on :%s", cfg.Port)
	if err := app.Listen(fmt.Sprintf(":%s", cfg.Port)); err != nil {
		log.Fatalf("server: %v", err)
	}
}

