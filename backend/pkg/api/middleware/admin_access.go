package middleware

import (
	"github.com/gofiber/fiber/v2"

	"ragpack/pkg/api/validate"
	"ragpack/pkg/meta"
)

// RequireAdminAccess returns middleware enforcing that the authenticated API
// key has `required` permission on the given instance-administration
// resource type (see meta.ResourceType) — for actions that aren't about
// collection content at all (managing keys, prompts, or collection
// lifecycle). Fully decoupled from CollectionGrant/RequireAccess in
// access.go: a key's collection access says nothing about its admin
// capability, by design.
func RequireAdminAccess(ms meta.MetaStore, resourceType meta.ResourceType, required meta.Permission) fiber.Handler {
	return func(c *fiber.Ctx) error {
		apiKey, ok := c.Locals(LocalAPIKey).(meta.APIKey)
		if !ok {
			_ = c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing API key"})
			return validate.ErrResponseWritten
		}
		grants, err := ms.ListAdminGrants(c.Context(), apiKey.ID)
		if err != nil {
			_ = c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
			return validate.ErrResponseWritten
		}
		if !hasAdminAccess(grants, resourceType, required) {
			_ = c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "key does not have this administrative capability"})
			return validate.ErrResponseWritten
		}
		return c.Next()
	}
}

// hasAdminAccess reports whether grants include one on resourceType — or a
// "*" grant (matches every resource type, current and future) — covering
// the required permission.
func hasAdminAccess(grants []meta.AdminGrant, resourceType meta.ResourceType, required meta.Permission) bool {
	for _, g := range grants {
		if g.ResourceType != resourceType && g.ResourceType != meta.ResourceAll {
			continue
		}
		if covers(g.Permission, required) {
			return true
		}
	}
	return false
}
