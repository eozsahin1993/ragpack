package query

import "github.com/gofiber/fiber/v2"

// Register mounts routes on the /:name group (e.g. /collections/:name).
// requireRead gates both routes on read access to the resolved collection —
// /query and /rag are POST (they carry a query body) but neither mutates
// anything, so they only need read, not write.
func Register(r fiber.Router, h *Handler, requireRead fiber.Handler) {
	r.Post("/query", requireRead, h.Query)
	r.Post("/rag", requireRead, h.Rag)
}
