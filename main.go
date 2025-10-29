package main

import (
	"log"

	"folo/database"
	"folo/ordering"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

func main() {
	// Initialize database
	if err := database.InitDatabase(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// Run auto migrations
	if err := database.AutoMigrate(&ordering.Basket{}); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	app := fiber.New(fiber.Config{
		AppName: "Folo API v1.0.0",
	})

	app.Use(logger.New())
	app.Use(recover.New())

	api := app.Group("/api")

	ordering.RegisterBasketsRoutes(api)
	ordering.RegisterOrderRoutes(api)

	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})

	log.Fatal(app.Listen(":3000"))
}
