package prompts

import "github.com/gofiber/fiber/v2"

// Register mounts prompt management routes, gated on the "prompts"
// instance-administration resource type (see meta.ResourceType). Prompts
// aren't collection-scoped — any collection can reference any prompt by
// slug — so this can't use a CollectionGrant check.
func Register(r fiber.Router, h *Handler, requireAdminRead, requireAdminWrite fiber.Handler) {
	r.Get("/", requireAdminRead, h.List)
	r.Post("/", requireAdminWrite, h.Create)
	r.Get("/:slug", requireAdminRead, h.Get)
	r.Patch("/:slug", requireAdminWrite, h.Update)
	r.Delete("/:slug", requireAdminWrite, h.Delete)
}
