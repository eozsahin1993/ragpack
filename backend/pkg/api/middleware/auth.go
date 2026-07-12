package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"ragpack/pkg/meta"
)

// LocalAPIKey is the c.Locals key the authenticated meta.APIKey is stored
// under, set by Auth and read by RequireAccess/CheckAccess.
const LocalAPIKey = "api_key"

func Auth(ms meta.MetaStore) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing API key"})
		}
		key := strings.TrimPrefix(header, "Bearer ")
		apiKey, err := ms.ValidateAPIKey(c.Context(), key)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid API key"})
		}
		c.Locals(LocalAPIKey, apiKey)
		return c.Next()
	}
}
