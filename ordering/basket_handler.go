package ordering

import (
	"log"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

type BasketHandler struct {
	basketRepo BasketRepository
}

func NewBasketHandler(basketRepo BasketRepository) *BasketHandler {
	return &BasketHandler{
		basketRepo: basketRepo,
	}
}

func RegisterBasketsRoutes(router fiber.Router, handler *BasketHandler) {
	baskets := router.Group("/baskets")

	baskets.Get("/", handler.GetBaskets)
	baskets.Get("/:id", handler.GetBasket)
	baskets.Post("/", handler.CreateBasketWithItems)
	baskets.Put("/:id", handler.UpdateBasket)
	baskets.Delete("/:id", handler.DeleteBasket)
}

func (h *BasketHandler) GetBaskets(c fiber.Ctx) error {
	baskets, err := h.basketRepo.FindAll(10)
	if err != nil {
		log.Printf("error retrieving all baskets: %s", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "failed to retrieve baskets",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    baskets,
	})
}

// GetBasket returns a single basket by ID
func (h *BasketHandler) GetBasket(c fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "invalid basket ID",
		})
	}

	basket, err := h.basketRepo.FindByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "basket not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    basket,
		"id":      id,
	})
}

func (h *BasketHandler) CreateBasketWithItems(c fiber.Ctx) error {
	basket := new(Basket)

	if err := c.Bind().Body(basket); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	if err := h.basketRepo.Create(basket); err != nil {
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

func (h *BasketHandler) UpdateBasket(c fiber.Ctx) error {
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

func (h *BasketHandler) DeleteBasket(c fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "invalid basket ID",
		})
	}

	if err := h.basketRepo.Delete(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "failed to delete basket",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Basket deleted successfully",
		"id":      id,
	})
}
