package ordering

import "gorm.io/gorm"

// OrderStatus represents the current status of an order
type OrderStatus string

const (
	Processing OrderStatus = "PROCESSING"
	Unpaid     OrderStatus = "UNPAID"
	Paid       OrderStatus = "PAID"
	Completed  OrderStatus = "COMPLETED"
	Failed     OrderStatus = "FAILED"
	Canceled   OrderStatus = "CANCELED"
)

// Order represents a customer order
type Order struct {
	gorm.Model
	OrderStatus OrderStatus
	IsDelivery  bool
	BasketID    uint `json:"-"`
	Basket      Basket
}

// DeliveryStatus represents the current status of a delivery
type DeliveryStatus string

const (
	Pending      DeliveryStatus = "PENDING"
	Dispatched   DeliveryStatus = "DISPATCHED"
	Interacted   DeliveryStatus = "INTERACTED"
	Delivered    DeliveryStatus = "DELIVERED"
	NotDelivered DeliveryStatus = "NOT_DELIVERED"
)

// DeliveryOrder represents an order with delivery information
type DeliveryOrder struct {
	DeliveryData   DeliveryData
	DeliveryStatus DeliveryStatus
	Order          Order
}

// UpdateStatus updates the delivery status
func (do *DeliveryOrder) UpdateStatus(deliveryStatus DeliveryStatus) {
	do.DeliveryStatus = deliveryStatus
}

// DeliveryData contains delivery address and contact information
type DeliveryData struct {
	gorm.Model
	Address     string
	PhoneNumber string
	Order       Order
}

// PaymentType represents the payment method used
type PaymentType string

const (
	Cash   PaymentType = "Cash"
	Credit PaymentType = "Credit"
	Gift   PaymentType = "Gift"
	Crypto PaymentType = "Crypto"
)

// OrderReq represents the request body for creating an order
type OrderReq struct {
	BasketId     uint
	PaymentType  PaymentType
	DeliveryData *DeliveryData
}

// IsDelivery checks if the order is a delivery order
func (or OrderReq) IsDelivery() bool {
	return or.DeliveryData != nil
}
