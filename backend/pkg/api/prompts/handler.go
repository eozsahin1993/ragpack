package prompts

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"ragpack/pkg/api/validate"
	"ragpack/pkg/meta"
)

type Handler struct {
	meta meta.MetaStore
}

func NewHandler(ms meta.MetaStore) *Handler {
	return &Handler{meta: ms}
}

func (h *Handler) List(c *fiber.Ctx) error {
	limit, offset := validate.Pagination(c)

	system, err := h.meta.ListSystemPrompts(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	user, err := h.meta.ListPrompts(c.Context(), limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	total, err := h.meta.CountPrompts(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"system": system,
		"user":   user,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreateRequest
	if err := validate.Body(c, &req); err != nil {
		return err
	}

	p, err := h.meta.CreatePrompt(c.Context(), meta.CreatePromptInput{
		Name:    req.Name,
		Content: req.Content,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(p)
}

func (h *Handler) Get(c *fiber.Ctx) error {
	p, err := h.meta.GetPromptBySlug(c.Context(), c.Params("slug"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "prompt not found"})
	}
	return c.JSON(p)
}

func (h *Handler) Update(c *fiber.Ctx) error {
	var req UpdateRequest
	if err := validate.Body(c, &req); err != nil {
		return err
	}

	p, err := h.meta.UpdatePrompt(c.Context(), c.Params("slug"), meta.UpdatePromptInput{
		Name:    req.Name,
		Content: req.Content,
	})
	if err != nil {
		if errors.Is(err, meta.ErrSystemReadOnly) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(p)
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	if err := h.meta.DeletePrompt(c.Context(), c.Params("slug")); err != nil {
		if errors.Is(err, meta.ErrSystemReadOnly) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
