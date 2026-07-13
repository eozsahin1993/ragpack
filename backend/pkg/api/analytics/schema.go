package analytics

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

const (
	defaultDays = 30
	maxDays     = 365
)

// cutoffFromDays parses ?days= (default 30, clamped to [1, 365]) into the
// UTC cutoff timestamp events must be >= to be included.
func cutoffFromDays(c *fiber.Ctx) time.Time {
	days := c.QueryInt("days", defaultDays)
	if days < 1 {
		days = 1
	}
	if days > maxDays {
		days = maxDays
	}
	return time.Now().UTC().AddDate(0, 0, -days)
}
