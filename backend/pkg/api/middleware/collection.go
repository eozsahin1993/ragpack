package middleware

import (
	"github.com/gofiber/fiber/v2"

	"ragpack/pkg/meta"
)

const LocalCollection = "collection"

func Collection(ms meta.MetaStore) fiber.Handler {
	return func(c *fiber.Ctx) error {
		col, err := ms.GetCollectionBySlug(c.Context(), c.Params("slug"))
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "collection not found"})
		}
		c.Locals(LocalCollection, col)
		return c.Next()
	}
}
