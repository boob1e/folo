package ordering

import (
	"gorm.io/gorm"
)

// OrderRepository handles database operations for orders
type OrderRepository interface {
	Create(order *Order) error
	FindByID(id uint) (*Order, error)
	Update(order *Order) error
}

type orderRepository struct {
	db *gorm.DB
}

// NewOrderRepository creates a new order repository
func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

// Create creates a new order in the database
func (r *orderRepository) Create(order *Order) error {
	return r.db.Create(order).Error
}

// FindByID finds an order by ID
func (r *orderRepository) FindByID(id uint) (*Order, error) {
	var order Order
	err := r.db.First(&order, id).Error
	return &order, err
}

// Update updates an existing order
func (r *orderRepository) Update(order *Order) error {
	return r.db.Save(order).Error
}

// DeliveryDataRepository handles database operations for delivery data
type DeliveryDataRepository interface {
	Create(deliveryData *DeliveryData) error
	FindByOrderID(orderID uint) (*DeliveryData, error)
	Update(deliveryData *DeliveryData) error
}

type deliveryDataRepository struct {
	db *gorm.DB
}

// NewDeliveryDataRepository creates a new delivery data repository
func NewDeliveryDataRepository(db *gorm.DB) DeliveryDataRepository {
	return &deliveryDataRepository{db: db}
}

// Create creates delivery data in the database
func (r *deliveryDataRepository) Create(deliveryData *DeliveryData) error {
	return r.db.Create(deliveryData).Error
}

// FindByOrderID finds delivery data by order ID
func (r *deliveryDataRepository) FindByOrderID(orderID uint) (*DeliveryData, error) {
	var deliveryData DeliveryData
	err := r.db.Where("order_id = ?", orderID).First(&deliveryData).Error
	return &deliveryData, err
}

// Update updates existing delivery data
func (r *deliveryDataRepository) Update(deliveryData *DeliveryData) error {
	return r.db.Save(deliveryData).Error
}
