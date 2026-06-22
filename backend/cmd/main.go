package main

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"ragpack/backend/pkg/api"
	lancedbpkg "ragpack/backend/pkg/db/lancedb"
	sqlitemeta "ragpack/backend/pkg/meta/sqlite"
)

func main() {
	ms, err := sqlitemeta.New("./ragpack.db")
	if err != nil {
		log.Fatalf("meta store: %v", err)
	}
	defer ms.Close()

	vec := lancedbpkg.New()
	if err := vec.Connect(context.Background(), "./lancedb_data"); err != nil {
		log.Fatalf("vector store: %v", err)
	}

	app := fiber.New(fiber.Config{
		AppName: "RagPack Engine v1.0",
	})

	app.Use(logger.New())

	api.Register(app, ms, vec)

	log.Println("RagPack server starting on :9000")
	if err := app.Listen(":9000"); err != nil {
		log.Fatalf("server: %v", err)
	}
}
