package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"ragpack/pkg/meta"
)

func Auth(ms meta.MetaStore) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing API key"})
		}
		key := strings.TrimPrefix(header, "Bearer ")
		if err := ms.ValidateAPIKey(c.Context(), key); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid API key"})
		}
		return c.Next()
	}
}
