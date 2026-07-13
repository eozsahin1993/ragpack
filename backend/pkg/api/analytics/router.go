package analytics

import "github.com/gofiber/fiber/v2"

// Register mounts admin-only analytics endpoints. No ACL middleware here —
// the admin surface (RegisterAdmin) is trusted by network placement alone,
// same as every other route mounted there directly rather than through
// mountRoutes; see pkg/api/router.go.
func Register(r fiber.Router, h *Handler) {
	r.Get("/volume", h.Volume)
	r.Get("/cost-by-collection", h.CostByCollection)
	r.Get("/latency", h.Latency)
	r.Get("/ingestion-failure-rate", h.IngestionFailureRate)
	r.Get("/token-usage", h.TokenUsage)
}
