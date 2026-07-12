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
	return PaginationWithDefault(c, DefaultLimit)
}

// PaginationWithDefault is Pagination with a caller-supplied default limit,
// for endpoints whose natural page size differs from DefaultLimit.
func PaginationWithDefault(c *fiber.Ctx, defaultLimit int) (limit, offset int) {
	limit = defaultLimit
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
