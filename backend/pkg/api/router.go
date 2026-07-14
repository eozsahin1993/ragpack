package api

import (
	"github.com/gofiber/fiber/v2"

	"ragpack/pkg/analytics"
	analyticsapi "ragpack/pkg/api/analytics"
	"ragpack/pkg/api/collections"
	"ragpack/pkg/api/collections/metadata_fields"
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
	"ragpack/pkg/telemetry"
)

// RegisterPublic mounts the external API (requires auth) on the given app.
// Intended for the public-facing port exposed to the internet.
func RegisterPublic(
	app *fiber.App,
	ms meta.MetaStore,
	vec db.VectorDb,
	registry *embedder.Registry,
	llmRegistry *llm.Registry,
	ing ingester.Ingester,
	defaultPromptSlug string,
	maxUploadSize int,
	minRefreshSeconds int,
	rec *telemetry.Recorder,
) {
	app.Get("/api/v1/health", healthHandler)

	v1 := app.Group("/api/v1")
	v1.Use(middleware.Auth(ms))
	mountRoutes(v1, ms, vec, registry, llmRegistry, ing, defaultPromptSlug, maxUploadSize, minRefreshSeconds, true, rec)
}

// RegisterAdmin mounts the admin API (no auth) on the given app.
// Intended for an internal-only port never published outside the Docker network.
func RegisterAdmin(
	app *fiber.App,
	ms meta.MetaStore,
	vec db.VectorDb,
	registry *embedder.Registry,
	llmRegistry *llm.Registry,
	ing ingester.Ingester,
	defaultPromptSlug string,
	maxUploadSize int,
	minRefreshSeconds int,
	rec *telemetry.Recorder,
	eng *analytics.Engine,
) {
	app.Get("/admin/health", healthHandler)
	adminGroup := app.Group("/admin")
	mountRoutes(adminGroup, ms, vec, registry, llmRegistry, ing, defaultPromptSlug, maxUploadSize, minRefreshSeconds, false, rec)

	// Deliberately not part of mountRoutes/RegisterPublic — analytics is
	// never meant to be public-facing (see TELEMETRY_ANALYTICS_PLAN.md's
	// trust boundary section), so it's admin-only by construction, not by
	// an ACL check. eng is nil when telemetry is disabled (see app.New) —
	// skipping registration means the routes just 404 rather than needing a
	// disabled-engine sentinel error in every handler.
	if eng != nil {
		analyticsapi.Register(adminGroup.Group("/analytics"), analyticsapi.NewHandler(eng))
	}
}

// mountRoutes wires up the shared route set for both the public (API-key
// gated) and admin (internal-network-only, no auth) surfaces. enforceACL is
// false on the admin surface — there's no API key in context there to check
// grants against, and admin routes are trusted by network placement instead.
func mountRoutes(
	r fiber.Router,
	ms meta.MetaStore,
	vec db.VectorDb,
	registry *embedder.Registry,
	llmRegistry *llm.Registry,
	ing ingester.Ingester,
	defaultPromptSlug string,
	maxUploadSize int,
	minRefreshSeconds int,
	enforceACL bool,
	rec *telemetry.Recorder,
) {
	requireRead := fiber.Handler(middleware.NoOp)
	requireWrite := fiber.Handler(middleware.NoOp)
	if enforceACL {
		requireRead = middleware.RequireAccess(ms, meta.PermissionRead)
		requireWrite = middleware.RequireAccess(ms, meta.PermissionWrite)
	}

	// adminMW gates instance-administration actions (managing keys, prompts,
	// or collection lifecycle) on their own resource type — fully decoupled
	// from CollectionGrant, since collection access says nothing about
	// whether a key should be able to do these.
	adminMW := func(resourceType meta.ResourceType, required meta.Permission) fiber.Handler {
		if !enforceACL {
			return middleware.NoOp
		}
		return middleware.RequireAdminAccess(ms, resourceType, required)
	}

	jobs.Register(r.Group("/jobs"), jobs.NewHandler(ms, enforceACL))

	documents.Register(r, documents.NewHandler(ms, vec, enforceACL))

	embedders.Register(r.Group("/embeddings"), embedders.NewHandler(registry))
	llms.Register(r.Group("/llms"), llms.NewHandler(llmRegistry))

	keys.Register(r.Group("/keys"), keys.NewHandler(ms),
		adminMW(meta.ResourceKeys, meta.PermissionRead), adminMW(meta.ResourceKeys, meta.PermissionWrite))
	prompts.Register(r.Group("/prompts"), prompts.NewHandler(ms),
		adminMW(meta.ResourcePrompts, meta.PermissionRead), adminMW(meta.ResourcePrompts, meta.PermissionWrite))

	collGroup := r.Group("/collections")
	collections.Register(collGroup, collections.NewHandler(ms, vec, registry, enforceACL, minRefreshSeconds),
		adminMW(meta.ResourceCollections, meta.PermissionRead), adminMW(meta.ResourceCollections, meta.PermissionWrite))

	nameGroup := collGroup.Group("/:slug")
	nameGroup.Use(middleware.Collection(ms))
	jobs.Register(nameGroup.Group("/jobs"), jobs.NewHandler(ms, enforceACL))
	ingest.Register(nameGroup, ingest.NewHandler(ms, ing, maxUploadSize), requireWrite)
	query.Register(nameGroup, query.NewHandler(ms, vec, registry, llmRegistry, defaultPromptSlug, rec), requireRead)
	documents.Register(nameGroup, documents.NewHandler(ms, vec, enforceACL))

	metadata_fields.Register(nameGroup.Group("/metadata-fields"), metadata_fields.NewHandler(ms, vec), requireRead, requireWrite)
}
