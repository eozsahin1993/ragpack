package keys

import (
	"context"

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

// keyResponse fetches k's grants and assembles the shared List/Create response shape.
func (h *Handler) keyResponse(ctx context.Context, k meta.APIKey) (KeyResponse, error) {
	grants, err := h.meta.ListGrants(ctx, k.ID)
	if err != nil {
		return KeyResponse{}, err
	}
	adminGrants, err := h.meta.ListAdminGrants(ctx, k.ID)
	if err != nil {
		return KeyResponse{}, err
	}
	return KeyResponse{APIKey: k, Grants: grants, AdminGrants: adminGrants}, nil
}

func (h *Handler) List(c *fiber.Ctx) error {
	keys, err := h.meta.ListAPIKeys(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	items := make([]KeyResponse, len(keys))
	for i, k := range keys {
		item, err := h.keyResponse(c.Context(), k)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		items[i] = item
	}

	return c.JSON(fiber.Map{"keys": items})
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreateRequest
	if err := validate.Body(c, &req); err != nil {
		return err
	}
	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
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

	item, err := h.keyResponse(c.Context(), key)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(CreateResponse{KeyResponse: item, Key: plaintext})
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	if err := h.meta.DeleteAPIKey(c.Context(), c.Params("id")); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
