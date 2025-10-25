package handlers

import (
	"time"

	"folo/database"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// Basket represents a basket entity
type Basket struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	SKU         int            `gorm:"column:sku;not null" json:"sku"`
	Name        string         `gorm:"column:name;size:255;not null" json:"name"`
	Description string         `gorm:"column:description;type:text" json:"description"`
	Items       int            `gorm:"column:items;default:0" json:"items"`
	CreatedAt   time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// RegisterBasketsRoutes sets up all basket-related routes
func RegisterBasketsRoutes(router fiber.Router) {
	baskets := router.Group("/baskets")

	baskets.Get("/", GetBaskets)
	baskets.Get("/:id", GetBasket)
	baskets.Post("/", CreateBasket)
	baskets.Put("/:id", UpdateBasket)
	baskets.Delete("/:id", DeleteBasket)
}

// GetBaskets returns all baskets
func GetBaskets(c *fiber.Ctx) error {
	baskets := []Basket{
		{
			ID:          1,
			SKU:         12345,
			Name:        "Shopping Basket",
			Description: "My shopping basket",
			Items:       5,
		},
		{
			ID:          2,
			SKU:         67890,
			Name:        "Wishlist",
			Description: "Items I want to buy later",
			Items:       12,
		},
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    baskets,
	})
}

func GetBasket(c *fiber.Ctx) error {
	id := c.Params("id")

	basket := Basket{
		ID:          1,
		SKU:         12345,
		Name:        "Shopping Basket",
		Description: "My shopping basket",
		Items:       5,
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    basket,
		"id":      id,
	})
}

// CreateBasket creates a new basket
func CreateBasket(c *fiber.Ctx) error {
	basket := new(Basket)

	if err := c.BodyParser(basket); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	// Create basket in database
	if err := database.DB.Create(basket).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to create basket",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    basket,
	})
}

// UpdateBasket updates an existing basket
func UpdateBasket(c *fiber.Ctx) error {
	id := c.Params("id")
	basket := new(Basket)

	if err := c.BodyParser(basket); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	// Mock update - keeping for now
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    basket,
		"id":      id,
	})
}

// DeleteBasket deletes a basket
func DeleteBasket(c *fiber.Ctx) error {
	id := c.Params("id")

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Basket deleted successfully",
		"id":      id,
	})
}
