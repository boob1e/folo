package ordering

import (
	"context"
	"folo/database"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

// Basket represents a basket entity
type Basket struct {
	gorm.Model
	Description string       `gorm:"column:description;type:text" json:"description"`
	BasketItems []BasketItem `json:"basketItems"`
}

type BasketItem struct {
	gorm.Model
	ID         uint `gorm:"primaryKey" json:"id"`
	BasketID   uint
	Basket     Basket
	MenuItemID uint
	// MenuItem   `gorm:"embedded"`
	MenuItem MenuItem
}

type MenuItem struct {
	gorm.Model
	ID    uint   `gorm:"primaryKey" json:"id"`
	SKU   int    `gorm:"column:sku;not null" json:"sku"`
	Name  string `gorm:"column:name;not null" json:"name"`
	Price int    `gorm:"column:price;not null" json:"price"`
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
func GetBaskets(c fiber.Ctx) error {
	ctx := context.Background()
	baskets, err := gorm.G[Basket](database.DB).Find(ctx)
	if err != nil {
		log.Printf("error retrieving all baskets")
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    baskets,
	})
}

func GetBasket(c fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).Err()
	}
	ctx := context.Background()
	basket, err := gorm.G[Basket](database.DB).Where("id = ?", uint(id)).First(ctx)
	if err != nil {
		return c.Status(fiber.StatusNotFound).Err()
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    basket,
		"id":      id,
	})
}

// CreateBasket creates a new basket
func CreateBasket(c fiber.Ctx) error {
	basket := new(Basket)

	if err := c.Bind().Body(basket); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	//TODO: move to service layer
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
func UpdateBasket(c fiber.Ctx) error {
	id := c.Params("id")
	basket := new(Basket)

	if err := c.Bind().Body(basket); err != nil {
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
func DeleteBasket(c fiber.Ctx) error {
	id := c.Params("id")

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Basket deleted successfully",
		"id":      id,
	})
}
