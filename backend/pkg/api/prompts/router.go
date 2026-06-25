package prompts

import "github.com/gofiber/fiber/v2"

func Register(r fiber.Router, h *Handler) {
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/:slug", h.Get)
	r.Patch("/:slug", h.Update)
	r.Delete("/:slug", h.Delete)
}
