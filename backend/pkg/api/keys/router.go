package keys

import "github.com/gofiber/fiber/v2"

// Register mounts API key management routes, gated on the "keys"
// instance-administration resource type (see meta.ResourceType) — decoupled
// from any collection grant a key might have.
func Register(r fiber.Router, h *Handler, requireAdminRead, requireAdminWrite fiber.Handler) {
	r.Get("/", requireAdminRead, h.List)
	r.Post("/", requireAdminWrite, h.Create)
	r.Delete("/:id", requireAdminWrite, h.Delete)
}
