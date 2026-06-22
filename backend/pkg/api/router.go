package api

import (
	"github.com/gofiber/fiber/v2"

	"ragpack/pkg/api/collections"
	"ragpack/pkg/api/ingest"
	"ragpack/pkg/api/jobs"
	"ragpack/pkg/api/query"
	"ragpack/pkg/db"
	"ragpack/pkg/meta"
)

func Register(app *fiber.App, ms meta.MetaStore, vec db.VectorDb) {
	v1 := app.Group("/api/v1")

	v1.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "healthy", "engine": "Go + Fiber"})
	})

	collGroup := v1.Group("/collections")
	collections.Register(collGroup, collections.NewHandler(ms, vec))

	nameGroup := collGroup.Group("/:name")
	jobs.Register(nameGroup, jobs.NewHandler(ms))
	ingest.Register(nameGroup, ingest.NewHandler(ms, vec))
	query.Register(nameGroup, query.NewHandler(ms, vec))
}
