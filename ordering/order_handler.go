package ordering

import (
	"github.com/gofiber/fiber/v3"
)

type OrderHandler struct {
	orderService OrderService
}

func NewOrderHandler(orderService OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

func RegisterOrderRoutes(router fiber.Router, handler *OrderHandler) {
	orders := router.Group("/orders")
	orders.Post("/submit", handler.CreateOrder)
}

func (h *OrderHandler) CreateOrder(c fiber.Ctx) error {
	or := new(OrderReq)
	if err := c.Bind().Body(or); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	order, err := h.orderService.CreateOrder(*or)
	if err != nil {
		// Check specific error types
		if err.Error() == "record not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "basket not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create order",
		})
	}

	return c.JSON(fiber.Map{
		"message":     "order created successfully",
		"order_id":    order.ID,
		"total":       order.Subtotal,
		"is_delivery": order.IsDelivery,
		"status":      order.OrderStatus,
	})
}
