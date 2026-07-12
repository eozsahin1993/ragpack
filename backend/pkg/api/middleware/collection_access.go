package middleware

import (
	"github.com/gofiber/fiber/v2"

	"ragpack/pkg/api/validate"
	"ragpack/pkg/meta"
)

// RequireAccess returns middleware enforcing that the authenticated API key
// (see Auth) has `required` permission on the collection resolved by a prior
// Collection middleware call. Must be mounted after both Auth and Collection.
func RequireAccess(ms meta.MetaStore, required meta.Permission) fiber.Handler {
	return func(c *fiber.Ctx) error {
		col, ok := c.Locals(LocalCollection).(meta.Collection)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "collection not resolved"})
		}
		if err := CheckAccess(c, ms, col.ID, required); err != nil {
			return err
		}
		return c.Next()
	}
}

// CheckAccess checks whether the authenticated API key (see Auth) has
// `required` permission on collectionID. Use this directly from a handler
// when the collection isn't known until after resolving the target resource
// (e.g. a document's CollectionID) rather than from the URL — RequireAccess
// can't cover that case since it only runs before the handler.
//
// On failure it writes the error response and returns
// validate.ErrResponseWritten; callers must return that error immediately.
func CheckAccess(c *fiber.Ctx, ms meta.MetaStore, collectionID string, required meta.Permission) error {
	apiKey, ok := c.Locals(LocalAPIKey).(meta.APIKey)
	if !ok {
		_ = c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing API key"})
		return validate.ErrResponseWritten
	}
	grants, err := ms.ListGrants(c.Context(), apiKey.ID)
	if err != nil {
		_ = c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		return validate.ErrResponseWritten
	}
	if !hasAccess(grants, collectionID, required) {
		_ = c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "key does not have access to this collection"})
		return validate.ErrResponseWritten
	}
	return nil
}

// RequireWildcard enforces that the authenticated API key has a wildcard
// (all-collections) grant covering `required` — for reading/listing content
// across every collection at once (no single collection ID to check a scoped
// grant against). Not for instance-administration actions — see
// admin_access.go's RequireAdminAccess; a key scoped to "read every
// collection's documents" has no bearing on whether it can create other
// API keys.
func RequireWildcard(ms meta.MetaStore, required meta.Permission) fiber.Handler {
	return func(c *fiber.Ctx) error {
		apiKey, ok := c.Locals(LocalAPIKey).(meta.APIKey)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing API key"})
		}
		grants, err := ms.ListGrants(c.Context(), apiKey.ID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		for _, g := range grants {
			if g.CollectionID == nil && covers(g.Permission, required) {
				return c.Next()
			}
		}
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "requires an unrestricted API key"})
	}
}

// hasAccess reports whether grants include one on collectionID — or a
// wildcard grant (CollectionID == nil) — covering the required permission.
func hasAccess(grants []meta.CollectionGrant, collectionID string, required meta.Permission) bool {
	for _, g := range grants {
		if g.CollectionID != nil && *g.CollectionID != collectionID {
			continue
		}
		if covers(g.Permission, required) {
			return true
		}
	}
	return false
}

// covers reports whether a granted permission satisfies a required one. Shared by access.go and admin_access.go.
func covers(granted, required meta.Permission) bool {
	if granted == meta.PermissionBoth {
		return true
	}
	return granted == required
}
