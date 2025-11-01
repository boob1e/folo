package main

import (
	"log"
	"os"

	"folo/database"
	"folo/delivery"
	"folo/ordering"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/joho/godotenv"
)

func main() {
	// Initialize database
	if err := database.InitDatabase(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// Run auto migrations
	if err := database.AutoMigrate(
		&ordering.Basket{},
		&ordering.BasketItem{},
		&ordering.MenuItem{},
		&ordering.Order{},
		&delivery.DeliveryData{}); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Initialize repositories
	orderRepo := ordering.NewOrderRepository(database.DB)
	basketRepo := ordering.NewBasketRepository(database.DB)
	deliveryDataRepo := ordering.NewDeliveryDataRepository(database.DB)

	// Initialize delivery service
	godotenv.Load()
	doorDashConfig := delivery.DoorDashConfig{
		DeveloperID:   os.Getenv("DOORDASH_DEVELOPER_ID"),
		KeyID:         os.Getenv("DOORDASH_KEY_ID"),
		SigningSecret: os.Getenv("DOORDASH_SIGNING_SECRET"),
	}
	doorDashService := delivery.NewDoorDashService(doorDashConfig)
	orderService := ordering.NewOrderService(orderRepo, basketRepo, deliveryDataRepo, doorDashService)

	// Initialize handlers
	orderHandler := ordering.NewOrderHandler(orderService)
	basketHandler := ordering.NewBasketHandler(basketRepo)

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		AppName: "Folo API v1.0.0",
	})

	app.Use(logger.New())
	app.Use(recover.New())

	api := app.Group("/api")

	// Register routes with handlers
	ordering.RegisterBasketsRoutes(api, basketHandler)
	ordering.RegisterOrderRoutes(api, orderHandler)

	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})

	log.Fatal(app.Listen(":3000"))
}
