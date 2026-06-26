package api

import (
	"github.com/gofiber/fiber/v2"

	"ragpack/pkg/api/collections"
	"ragpack/pkg/api/documents"
	"ragpack/pkg/api/embedders"
	"ragpack/pkg/api/ingest"
	"ragpack/pkg/api/jobs"
	"ragpack/pkg/api/keys"
	"ragpack/pkg/api/llms"
	"ragpack/pkg/api/middleware"
	"ragpack/pkg/api/prompts"
	"ragpack/pkg/api/query"
	"ragpack/pkg/db"
	"ragpack/pkg/embedder"
	"ragpack/pkg/ingester"
	"ragpack/pkg/llm"
	"ragpack/pkg/meta"
)

// RegisterPublic mounts the external API (requires auth) on the given app.
// Intended for the public-facing port exposed to the internet.
func RegisterPublic(app *fiber.App, ms meta.MetaStore, vec db.VectorDb, registry *embedder.Registry, llmRegistry *llm.Registry, ing ingester.Ingester, defaultPromptSlug string) {
	app.Get("/api/v1/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "healthy", "engine": "Go + Fiber"})
	})

	v1 := app.Group("/api/v1")
	v1.Use(middleware.Auth(ms))
	mountRoutes(v1, ms, vec, registry, llmRegistry, ing, defaultPromptSlug)
}

// RegisterAdmin mounts the admin API (no auth) on the given app.
// Intended for an internal-only port never published outside the Docker network.
func RegisterAdmin(app *fiber.App, ms meta.MetaStore, vec db.VectorDb, registry *embedder.Registry, llmRegistry *llm.Registry, ing ingester.Ingester, defaultPromptSlug string) {
	adminGroup := app.Group("/admin")
	mountRoutes(adminGroup, ms, vec, registry, llmRegistry, ing, defaultPromptSlug)
}

func mountRoutes(r fiber.Router, ms meta.MetaStore, vec db.VectorDb, registry *embedder.Registry, llmRegistry *llm.Registry, ing ingester.Ingester, defaultPromptSlug string) {
	r.Get("/jobs", jobs.NewHandler(ms).GetAllJobs)
	r.Get("/jobs/:id", jobs.NewHandler(ms).GetJob)

	embedders.Register(r.Group("/embeddings"), embedders.NewHandler(registry))
	llms.Register(r.Group("/llms"), llms.NewHandler(llmRegistry))

	keys.Register(r.Group("/keys"), keys.NewHandler(ms))
	prompts.Register(r.Group("/prompts"), prompts.NewHandler(ms))

	collGroup := r.Group("/collections")
	collections.Register(collGroup, collections.NewHandler(ms, vec, registry))

	nameGroup := collGroup.Group("/:slug")
	jobs.Register(nameGroup, jobs.NewHandler(ms))
	ingest.Register(nameGroup, ingest.NewHandler(ms, ing))
	query.Register(nameGroup, query.NewHandler(ms, vec, registry, llmRegistry, defaultPromptSlug))
	documents.Register(nameGroup, documents.NewHandler(ms, vec))
}
