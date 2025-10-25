package main

import (
	"log"

	"folo/database"
	"folo/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// Initialize database
	if err := database.InitDatabase(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// Run auto migrations
	if err := database.AutoMigrate(&handlers.Basket{}); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	app := fiber.New(fiber.Config{
		AppName: "Folo API v1.0.0",
	})

	app.Use(logger.New())
	app.Use(recover.New())

	api := app.Group("/api")

	handlers.RegisterBasketsRoutes(api)

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})

	log.Fatal(app.Listen(":3000"))
}
