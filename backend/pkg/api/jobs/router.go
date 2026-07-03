package jobs

import "github.com/gofiber/fiber/v2"

// Register mounts the top-level job routes.
func Register(r fiber.Router, h *Handler) {
	r.Get("", h.ListJobs)
	r.Get("/:id", h.GetJob)
	r.Delete("/:id", h.DeleteJob)
}

