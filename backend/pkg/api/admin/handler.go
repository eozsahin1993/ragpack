package admin

import (
	"github.com/gofiber/fiber/v2"

	"ragpack/pkg/embedder"
)

type Handler struct {
	registry *embedder.Registry
}

func NewHandler(registry *embedder.Registry) *Handler {
	return &Handler{registry: registry}
}

func (h *Handler) ListEmbedders(c *fiber.Ctx) error {
	defaultModel, _, _ := h.registry.Default()
	return c.JSON(fiber.Map{
		"models":  h.registry.Models(),
		"default": defaultModel,
	})
}
