package collections

import "github.com/gofiber/fiber/v2"

// Register mounts collection management routes, gated on the "collections"
// instance-administration resource type (see meta.ResourceType) — this is
// about managing collections as objects (create/rename/delete/list them),
// not the same thing as a CollectionGrant's per-collection content access.
func Register(r fiber.Router, h *Handler, requireAdminRead, requireAdminWrite fiber.Handler) {
	r.Post("/", requireAdminWrite, h.Create)
	r.Get("/", requireAdminRead, h.List)
	r.Get("/id/:id", requireAdminRead, h.GetByID)
	r.Patch("/id/:id", requireAdminWrite, h.PatchCollection)
	r.Delete("/id/:id", requireAdminWrite, h.DeleteByID)
	r.Get("/:slug", requireAdminRead, h.GetBySlug)
	r.Delete("/:slug", requireAdminWrite, h.DeleteBySlug)
}
