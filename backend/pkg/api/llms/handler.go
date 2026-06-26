package llms

import (
	"github.com/gofiber/fiber/v2"

	"ragpack/pkg/llm"
)

type Handler struct {
	registry *llm.Registry
}

func NewHandler(registry *llm.Registry) *Handler {
	return &Handler{registry: registry}
}

func (h *Handler) List(c *fiber.Ctx) error {
	defaultModel, _, _ := h.registry.Default()
	return c.JSON(fiber.Map{
		"models":  h.registry.Models(),
		"default": defaultModel,
	})
}
