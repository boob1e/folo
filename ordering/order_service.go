package ordering

import (
	"context"
	"log"
	"time"

	"folo/delivery"
)

// OrderService handles order business logic
type OrderService interface {
	CreateOrder(req OrderReq) (*Order, int64, error)
}

type orderService struct {
	orderRepo        OrderRepository
	basketRepo       BasketRepository
	deliveryDataRepo DeliveryDataRepository
	deliveryService  delivery.DeliveryService
}

// NewOrderService creates a new order service
func NewOrderService(
	orderRepo OrderRepository,
	basketRepo BasketRepository,
	deliveryDataRepo DeliveryDataRepository,
	deliveryService delivery.DeliveryService,
) OrderService {
	return &orderService{
		orderRepo:        orderRepo,
		basketRepo:       basketRepo,
		deliveryDataRepo: deliveryDataRepo,
		deliveryService:  deliveryService,
	}
}

// CreateOrder creates a new order from a basket
func (s *orderService) CreateOrder(req OrderReq) (*Order, int64, error) {
	// Fetch basket with all items and menu details
	basket, err := s.basketRepo.FindByIDWithItems(req.BasketId)
	if err != nil {
		return nil, 0, err
	}

	// Calculate order total
	orderTotal := basket.CalculateTotal()

	// If delivery order, launch async goroutine to get quote from DoorDash
	if req.IsDelivery() {
		go s.handleDeliveryQuote(req, orderTotal)
	}

	// Create the order - not waiting for routine to finish
	order := &Order{
		OrderStatus: Processing,
		IsDelivery:  req.IsDelivery(),
		BasketID:    req.BasketId,
	}

	if err := s.orderRepo.Create(order); err != nil {
		return nil, 0, err
	}

	// If delivery order, create DeliveryData record
	if req.IsDelivery() {
		deliveryData := &DeliveryData{
			Address:     req.DeliveryData.Address,
			PhoneNumber: req.DeliveryData.PhoneNumber,
			Order:       *order,
		}

		if err := s.deliveryDataRepo.Create(deliveryData); err != nil {
			log.Printf("failed to create delivery data: %s", err.Error())
			// Order already created, just log the error
		}
	}

	return order, orderTotal, nil
}

// handleDeliveryQuote handles the async delivery quote request
func (s *orderService) handleDeliveryQuote(req OrderReq, orderTotal int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 3000*time.Millisecond)
	defer cancel()

	// TODO: Replace with actual restaurant pickup details from config
	params := delivery.DeliveryQuoteParams{
		PickupAddress:      "123 Restaurant St, City, State, ZIP",
		PickupPhoneNumber:  "+11234567890",
		DropoffAddress:     req.DeliveryData.Address,
		DropoffPhoneNumber: req.DeliveryData.PhoneNumber,
		OrderValue:         orderTotal,
	}

	result, err := s.deliveryService.RequestQuote(ctx, params)
	if err != nil {
		log.Printf("DoorDash quote error: %s", err.Error())
		// TODO: Update delivery status to indicate failure
		return
	}

	// Update DeliveryData with quote information
	log.Printf("DoorDash quote received: Fee=%d, ID=%s", result.Fee, result.ID)
	// TODO: Update DeliveryData record with quote details (Fee, DoorDash ID, etc.)
}
