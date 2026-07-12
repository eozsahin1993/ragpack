package metadata_fields

import "github.com/gofiber/fiber/v2"

func Register(r fiber.Router, h *Handler, requireRead, requireWrite fiber.Handler) {
	r.Post("/", requireWrite, h.Register)
	r.Get("/", requireRead, h.List)
	r.Delete("/:name", requireWrite, h.Delete)
}
