package ordering

import (
	"context"
	"log"
	"time"

	"folo/database"
	"folo/delivery"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

type OrderStatus string

const (
	Processing OrderStatus = "PROCESSING"
	Unpaid     OrderStatus = "UNPAID"
	Paid       OrderStatus = "PAID"
	Completed  OrderStatus = "COMPLETED"
	Failed     OrderStatus = "FAILED"
	Canceled   OrderStatus = "CANCELED"
)

type Order struct {
	gorm.Model
	OrderStatus OrderStatus
	IsDelivery  bool
	BasketID    uint `json:"-"`
	Basket      Basket
}

type DeliveryStatus string

const (
	Pending      DeliveryStatus = "PENDING"
	Dispatched   DeliveryStatus = "DISPATCHED"
	Interacted   DeliveryStatus = "INTERACTED"
	Delivered    DeliveryStatus = "DELIVERED"
	NotDelivered DeliveryStatus = "NOT_DELIVERED"
)

type DeliveryOrder struct {
	DeliveryData   DeliveryData
	DeliveryStatus DeliveryStatus
	Order          Order
}

func (do *DeliveryOrder) updateStatus(deliveryStatus DeliveryStatus) {
	do.DeliveryStatus = deliveryStatus
}

type DeliveryData struct {
	gorm.Model
	Address     string
	PhoneNumber string
	Order       Order
}

type PaymentType string

const (
	Cash   PaymentType = "Cash"
	Credit PaymentType = "Credit"
	Gift   PaymentType = "Gift"
	Crypto PaymentType = "Crypto"
)

type OrderReq struct {
	BasketId     uint
	PaymentType  PaymentType
	DeliveryData *DeliveryData
}

func (or OrderReq) isDelivery() bool {
	return or.DeliveryData != nil
}

func RegisterOrderRoutes(router fiber.Router) {
	orders := router.Group("/orders")
	orders.Post("/submit", CreateOrder)
}

func CreateOrder(c fiber.Ctx) error {
	or := new(OrderReq)
	if err := c.Bind().Body(or); err != nil {
		return err
	}

	// Fetch basket with all items and menu item details
	var basket Basket
	if err := database.DB.Preload("BasketItems.MenuItem").First(&basket, or.BasketId).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "basket not found",
		})
	}

	// Calculate order total from basket items
	var orderTotal int64 = 0
	for _, item := range basket.BasketItems {
		orderTotal += int64(item.MenuItem.Price * item.Quantity)
	}

	// Channel to receive delivery quote results (if delivery order)
	var resultChan chan delivery.QuoteResult

	// If this is a delivery order, launch goroutine to get quote from DoorDash
	if or.isDelivery() {
		resultChan = make(chan delivery.QuoteResult, 1)
		ctx, cancel := context.WithTimeout(context.Background(), 3000*time.Millisecond)

		// TODO: Replace with actual restaurant pickup details
		params := delivery.DeliveryQuoteParams{
			PickupAddress:      "123 Restaurant St, City, State, ZIP",
			PickupPhoneNumber:  "+11234567890",
			DropoffAddress:     or.DeliveryData.Address,
			DropoffPhoneNumber: or.DeliveryData.PhoneNumber,
			OrderValue:         orderTotal,
		}

		go delivery.CreateDeliveryQuote(ctx, params, resultChan)

		// Launch background goroutine to handle DoorDash result and update order
		go func() {
			defer cancel()
			result := <-resultChan
			if result.Err != nil {
				log.Printf("DoorDash quote error: %s", result.Err.Error())
				// TODO: Update delivery status to indicate failure
				return
			}

			// Update DeliveryData with quote information
			log.Printf("DoorDash quote received: Fee=%d, ID=%s", result.Response.Fee, result.Response.ID)
			// TODO: Update DeliveryData record with quote details (Fee, DoorDash ID, etc.)
		}()
	}

	// Create the order - not waiting for routine to finish
	order := Order{
		OrderStatus: Processing,
		IsDelivery:  or.isDelivery(),
		BasketID:    or.BasketId,
	}

	if err := database.DB.Create(&order).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create order",
		})
	}

	// If delivery order, create DeliveryData record
	if or.isDelivery() {
		deliveryData := DeliveryData{
			Address:     or.DeliveryData.Address,
			PhoneNumber: or.DeliveryData.PhoneNumber,
			Order:       order,
		}

		if err := database.DB.Create(&deliveryData).Error; err != nil {
			log.Printf("failed to create delivery data: %s", err.Error())
			// Order already created, just log the error
		}
	}

	// TODO: Process payment based on or.PaymentType and update order status

	return c.JSON(fiber.Map{
		"message":     "order created successfully",
		"order_id":    order.ID,
		"total":       orderTotal,
		"is_delivery": order.IsDelivery,
		"status":      order.OrderStatus,
	})
}
