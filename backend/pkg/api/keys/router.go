package keys

import "github.com/gofiber/fiber/v2"

func Register(r fiber.Router, h *Handler) {
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Delete("/:id", h.Delete)
}
