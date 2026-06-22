package collections

import "github.com/gofiber/fiber/v2"

func Register(r fiber.Router, h *Handler) {
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/id/:id", h.GetByID)
	r.Delete("/id/:id", h.DeleteByID)
	r.Get("/:name", h.GetByName)
	r.Delete("/:name", h.Delete)
}
