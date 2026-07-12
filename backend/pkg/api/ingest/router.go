package ingest

import "github.com/gofiber/fiber/v2"

// Register mounts routes on the /:name group (e.g. /collections/:name).
// requireWrite gates the route on write access to the resolved collection.
func Register(r fiber.Router, h *Handler, requireWrite fiber.Handler) {
	r.Post("/ingest", requireWrite, h.Ingest)
}
