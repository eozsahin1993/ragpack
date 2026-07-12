package middleware

import "github.com/gofiber/fiber/v2"

// NoOp is a passthrough handler, used in place of an access-control
// middleware on the admin surface (RegisterAdmin), which has no API key in
// context to check — admin routes are trusted by virtue of only being
// reachable on the internal Docker network.
func NoOp(c *fiber.Ctx) error {
	return c.Next()
}
