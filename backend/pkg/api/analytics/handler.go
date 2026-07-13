package analytics

import (
	"github.com/gofiber/fiber/v2"

	"ragpack/pkg/analytics"
)

type Handler struct {
	eng *analytics.Engine
}

func NewHandler(eng *analytics.Engine) *Handler {
	return &Handler{eng: eng}
}

func (h *Handler) Volume(c *fiber.Ctx) error {
	points, err := h.eng.VolumeOverTime(c.Context(), cutoffFromDays(c))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"points": points})
}

func (h *Handler) CostByCollection(c *fiber.Ctx) error {
	costs, err := h.eng.CostByCollection(c.Context(), cutoffFromDays(c))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"collections": costs})
}

func (h *Handler) Latency(c *fiber.Ctx) error {
	buckets, err := h.eng.Latency(c.Context(), cutoffFromDays(c))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"buckets": buckets})
}

func (h *Handler) IngestionSuccessRate(c *fiber.Ctx) error {
	rates, err := h.eng.IngestionSuccessRate(c.Context(), cutoffFromDays(c))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"mime_types": rates})
}

func (h *Handler) TokenUsage(c *fiber.Ctx) error {
	tokens, err := h.eng.TokenUsageByCollection(c.Context(), cutoffFromDays(c))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"collections": tokens})
}
