package metadata_fields

import "github.com/gofiber/fiber/v2"

func Register(r fiber.Router, h *Handler) {
	r.Post("/", h.Register)
	r.Get("/", h.List)
	r.Delete("/:name", h.Delete)
}
