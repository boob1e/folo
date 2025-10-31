package ordering

import (
	"context"
	"log"
	"time"

	"folo/delivery"
	"folo/payment"
)

// OrderService handles order business logic
type OrderService interface {
	CreateOrder(req OrderReq) (*Order, error)
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
func (s *orderService) CreateOrder(req OrderReq) (*Order, error) {
	basket, err := s.basketRepo.FindByIDWithItems(req.BasketId)
	if err != nil {
		return nil, err
	}

	orderTotal := basket.CalculateTotal()

	quoteChan := make(chan *delivery.QuoteResult, 1)
	// If delivery order, launch async goroutine to get quote from DoorDash
	if req.IsDelivery() {
		go s.handleDeliveryQuote(req, orderTotal, quoteChan)
	}

	// Create the order - not waiting for routine to finish
	order := &Order{
		OrderStatus: Processing,
		IsDelivery:  req.IsDelivery(),
		BasketID:    req.BasketId,
		Subtotal:    orderTotal,
	}
	if err := s.orderRepo.Create(order); err != nil {
		return nil, err
	}
	log.Printf("order created with orderID: %v", order.ID)

	// If delivery order, wait for quote
	if order.IsDelivery {
		select {
		case result := <-quoteChan:
			s.addDeliveryToOrder(result, order, req)
		case <-time.After(5 * time.Second):
			log.Printf("Timeout waiting for delivery quote for order %d", order.ID)
			// Could update delivery status to "quote_timeout" here
		}
	}

	// Process payment for all orders (pickup and delivery)
	err = processOrderWithPayment(order)
	if err != nil {
		log.Printf("error paying for order: %v", err)
	}

	return order, nil
}

func (s *orderService) addDeliveryToOrder(result *delivery.QuoteResult, order *Order, req OrderReq) {
	if result.Error != nil {
		log.Printf("DoorDash quote error for order %d: %s", order.ID, result.Error.Error())
	} else {
		log.Printf("DoorDash quote received for order %d: Fee=%d, ID=%s",
			order.ID, result.Response.Fee, result.Response.ID)

		deliveryData := &DeliveryData{
			Address:     req.DeliveryData.Address,
			PhoneNumber: req.DeliveryData.PhoneNumber,
			Order:       *order,
		}

		if err := s.deliveryDataRepo.Create(deliveryData); err != nil {
			log.Printf("failed to create delivery data: %s", err.Error())
			// Order already created, just log the error
		}
		order.Subtotal = order.Subtotal + int(result.Response.Fee)
	}
}

// handleDeliveryQuote handles the async delivery quote request
func (s *orderService) handleDeliveryQuote(req OrderReq, orderTotal int, resultChan chan<- *delivery.QuoteResult) {
	ctx, cancel := context.WithTimeout(context.Background(), 3000*time.Millisecond)
	defer cancel()

	// TODO: Replace with actual restaurant pickup details from config
	params := delivery.DeliveryQuoteParams{
		PickupAddress:      "303 2nd St, San Francisco, CA 94107",
		PickupPhoneNumber:  "+11234567890",
		DropoffAddress:     req.DeliveryData.Address,
		DropoffPhoneNumber: req.DeliveryData.PhoneNumber,
		OrderValue:         orderTotal,
	}

	result, err := s.deliveryService.RequestQuote(ctx, params)
	if err != nil {
		log.Printf("DoorDash quote error: %s", err.Error())
		resultChan <- &delivery.QuoteResult{
			Response: result,
			Error:    err,
		}
		return
	}

	// Update DeliveryData with quote information
	log.Printf("DoorDash quote received: Fee=%d, ID=%s", result.Fee, result.ID)
	// TODO: Update DeliveryData record with quote details (Fee, DoorDash ID, etc.)
}

func processOrderWithPayment(o *Order) error {
	pg := payment.NewPaymentGateway()
	payResult := pg.ProcessPayment("123")

	if payResult {
		o.OrderStatus = Processing
	} else {
		o.OrderStatus = Failed
	}

	return nil
}
