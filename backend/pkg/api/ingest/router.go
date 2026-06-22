package ingest

import "github.com/gofiber/fiber/v2"

// Register mounts routes on the /:name group (e.g. /collections/:name).
func Register(r fiber.Router, h *Handler) {
	r.Post("/ingest", h.Ingest)
}
