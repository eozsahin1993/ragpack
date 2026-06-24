package api

import (
	"github.com/gofiber/fiber/v2"

	"ragpack/pkg/api/collections"
	"ragpack/pkg/api/documents"
	"ragpack/pkg/api/ingest"
	"ragpack/pkg/api/jobs"
	"ragpack/pkg/api/keys"
	"ragpack/pkg/api/middleware"
	"ragpack/pkg/api/query"
	"ragpack/pkg/db"
	"ragpack/pkg/embedder"
	"ragpack/pkg/ingester"
	"ragpack/pkg/meta"
)

func Register(app *fiber.App, ms meta.MetaStore, vec db.VectorDb, registry *embedder.Registry, ing ingester.Ingester) {
	app.Get("/api/v1/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "healthy", "engine": "Go + Fiber"})
	})

	// External API — requires authentication
	v1 := app.Group("/api/v1")
	v1.Use(middleware.Auth(ms))
	mountRoutes(v1, ms, vec, registry, ing)

	// Admin API — internal only, no auth (never published outside Docker network)
	admin := app.Group("/admin")
	mountRoutes(admin, ms, vec, registry, ing)
}

func mountRoutes(r fiber.Router, ms meta.MetaStore, vec db.VectorDb, registry *embedder.Registry, ing ingester.Ingester) {
	r.Get("/jobs", jobs.NewHandler(ms).GetAllJobs)
	r.Get("/jobs/:id", jobs.NewHandler(ms).GetJob)

	keys.Register(r.Group("/keys"), keys.NewHandler(ms))

	collGroup := r.Group("/collections")
	collections.Register(collGroup, collections.NewHandler(ms, vec, registry))

	nameGroup := collGroup.Group("/:slug")
	jobs.Register(nameGroup, jobs.NewHandler(ms))
	ingest.Register(nameGroup, ingest.NewHandler(ms, ing))
	query.Register(nameGroup, query.NewHandler(ms, vec, registry))
	documents.Register(nameGroup, documents.NewHandler(ms, vec))
}
