package keys

import (
	"github.com/gofiber/fiber/v2"

	"ragpack/pkg/api/validate"
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
	if err := validate.Body(c, &req); err != nil {
		return err
	}

	grants := make([]meta.GrantInput, len(req.Grants))
	for i, g := range req.Grants {
		input := meta.GrantInput{Permission: g.Permission}
		if g.CollectionSlug != "" {
			col, err := h.meta.GetCollectionBySlug(c.Context(), g.CollectionSlug)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "unknown collection: " + g.CollectionSlug})
			}
			input.CollectionID = &col.ID
		}
		grants[i] = input
	}

	adminGrants := make([]meta.AdminGrantInput, len(req.AdminGrants))
	for i, g := range req.AdminGrants {
		adminGrants[i] = meta.AdminGrantInput{ResourceType: g.ResourceType, Permission: g.Permission}
	}

	plaintext, _, _, err := auth.Generate()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate key"})
	}

	key, err := h.meta.CreateAPIKey(c.Context(), req.Name, plaintext, grants, adminGrants)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	stored, err := h.meta.ListGrants(c.Context(), key.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	storedAdmin, err := h.meta.ListAdminGrants(c.Context(), key.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(CreateResponse{
		Key:         plaintext,
		ID:          key.ID,
		Name:        key.Name,
		KeyHint:     key.KeyHint,
		Grants:      stored,
		AdminGrants: storedAdmin,
	})
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	if err := h.meta.DeleteAPIKey(c.Context(), c.Params("id")); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
