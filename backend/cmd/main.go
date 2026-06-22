package main

import (
	"log"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	app := fiber.New(fiber.Config{
		AppName: "RagPack Engine v1.0",
	})

	app.Use(logger.New())

	app.Get("/api/v1/health", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{
			"status": "healthy",
			"engine": "Go + Fiber",
		})
	})

	log.Println("🚀 RagPack server starting on port :9000")

	app.Listen(":9000")
}