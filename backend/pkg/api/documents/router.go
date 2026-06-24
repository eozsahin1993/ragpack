package documents

import "github.com/gofiber/fiber/v2"

func Register(r fiber.Router, h *Handler) {
	r.Get("/documents", h.List)
	r.Get("/documents/:id", h.Get)
	r.Delete("/documents/:id", h.Delete)
}
