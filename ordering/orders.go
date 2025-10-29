package ordering

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v3"
)

type Order struct {
	// ID          uint
	CreatedAt   time.Time
	OrderStatus string
	IsDelivery  bool
	Basket      Basket
}

type DeliveryOrder struct {
	DeliveryData   DeliveryData
	DeliveryStatus string
	Order          Order
}

func (do *DeliveryOrder) updateStatus(deliveryStatus string) {
	do.DeliveryStatus = deliveryStatus
}

type DeliveryData struct {
	Address     string
	PhoneNumber string
}

type OrderReq struct {
	BasketId     uint
	PaymentType  string
	DeliveryData *DeliveryData
}

func (or OrderReq) isDelivery() bool {
	return false
}

func RegisterOrderRoutes(router fiber.Router) {
	orders := router.Group("/orders")
	orders.Post("/submit", SubmitOrderBasket)
}

func SubmitOrderBasket(c fiber.Ctx) error {
	or := new(OrderReq)
	if err := c.Bind().Body(or); err != nil {
		return err
	}
	log.Println("order is a delivery: ", or.isDelivery())
	return c.JSON("basket submitted")
}
