package collections

import "github.com/gofiber/fiber/v2"

func Register(r fiber.Router, h *Handler) {
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/id/:id", h.GetByID)
	r.Patch("/id/:id", h.PatchCollection)
	r.Delete("/id/:id", h.DeleteByID)
	r.Get("/:slug", h.GetBySlug)
	r.Delete("/:slug", h.DeleteBySlug)
}
