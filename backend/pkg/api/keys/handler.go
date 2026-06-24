package keys

import (
	"github.com/gofiber/fiber/v2"

	"ragpack/pkg/auth"
	"ragpack/pkg/meta"
)

type Handler struct {
	meta meta.MetaStore
}

func NewHandler(ms meta.MetaStore) *Handler {
	return &Handler{meta: ms}
}

func (h *Handler) List(c *fiber.Ctx) error {
	keys, err := h.meta.ListAPIKeys(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"keys": keys})
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreateRequest
	if err := c.BodyParser(&req); err != nil || req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name is required"})
	}

	plaintext, _, _, err := auth.Generate()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate key"})
	}

	key, err := h.meta.CreateAPIKey(c.Context(), req.Name, plaintext)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(CreateResponse{
		Key:     plaintext,
		ID:      key.ID,
		Name:    key.Name,
		KeyHint: key.KeyHint,
	})
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	if err := h.meta.DeleteAPIKey(c.Context(), c.Params("id")); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
