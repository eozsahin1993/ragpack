package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"

	"ragpack/pkg/app"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
	registry := embedder.NewRegistryFromConfig(ctx, cfg)
	llmRegistry := llm.NewRegistryFromConfig(cfg)

	a := app.New(ctx, app.Deps{
		Meta:      ms,
		Vector:    vec,
		Embedders: registry,
		LLMs:      llmRegistry,
		Config:    cfg,
	})

	go handleShutdown(cancel, a.Ingester, a.Public, a.Admin)

	go mustListen(a.Admin, cfg.AdminPort, "admin (internal)")
	mustListen(a.Public, cfg.Port, "public API")
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
