package api

import (
	"fmt"
	"math"
	"time"

	"github.com/gofiber/fiber/v2"
)

func healthHandler(c *fiber.Ctx) error {
	uptime := time.Since(startedAt)
	hours := int(math.Floor(uptime.Hours()))
	minutes := int(math.Floor(uptime.Minutes())) % 60
	seconds := int(uptime.Seconds()) % 60
	var uptimeStr string
	if hours > 0 {
		uptimeStr = fmt.Sprintf("%dh %dm", hours, minutes)
	} else if minutes > 0 {
		uptimeStr = fmt.Sprintf("%dm %ds", minutes, seconds)
	} else {
		uptimeStr = fmt.Sprintf("%ds", seconds)
	}
	return c.JSON(fiber.Map{
		"status":  "healthy",
		"version": Version,
		"uptime":  uptimeStr,
	})
}
