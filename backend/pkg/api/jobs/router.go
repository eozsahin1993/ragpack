package jobs

import "github.com/gofiber/fiber/v2"

// Register mounts routes on the /:name group (e.g. /collections/:name).
func Register(r fiber.Router, h *Handler) {
	r.Get("/jobs", h.ListByCollection)
	r.Get("/jobs/status/:status", h.ListByStatus)
	r.Get("/jobs/:id", h.GetJob)
}
