package api

import (
	"github.com/gofiber/fiber/v2"

	"ragpack/pkg/api/collections"
	"ragpack/pkg/api/documents"
	"ragpack/pkg/api/ingest"
	"ragpack/pkg/api/jobs"
	"ragpack/pkg/api/query"
	"ragpack/pkg/db"
	"ragpack/pkg/embedder"
	"ragpack/pkg/ingester"
	"ragpack/pkg/meta"
)

func Register(app *fiber.App, ms meta.MetaStore, vec db.VectorDb, registry *embedder.Registry, ing ingester.Ingester) {
	v1 := app.Group("/api/v1")

	v1.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "healthy", "engine": "Go + Fiber"})
	})

	jobHandler := jobs.NewHandler(ms)
	v1.Get("/jobs", jobHandler.GetAllJobs)
	v1.Get("/jobs/:id", jobHandler.GetJob)

	collGroup := v1.Group("/collections")
	collections.Register(collGroup, collections.NewHandler(ms, vec))

	nameGroup := collGroup.Group("/:slug")
	jobs.Register(nameGroup, jobHandler)
	ingest.Register(nameGroup, ingest.NewHandler(ms, ing))
	query.Register(nameGroup, query.NewHandler(ms, vec, registry))
	documents.Register(nameGroup, documents.NewHandler(ms))
}
