package validate

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

const (
	DefaultLimit = 50
	MaxLimit     = 500
)

// Pagination parses ?limit= and ?offset= from the request query string.
// limit is clamped to [1, MaxLimit]; offset defaults to 0.
func Pagination(c *fiber.Ctx) (limit, offset int) {
	limit = DefaultLimit
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			if v > MaxLimit {
				v = MaxLimit
			}
			limit = v
		}
	}
	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}
	return
}
