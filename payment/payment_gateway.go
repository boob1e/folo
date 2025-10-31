package payment

import "time"

type PaymentGateway interface {
	ProcessPayment(cc string) bool
}

type paymentGateway struct{}

func NewPaymentGateway() PaymentGateway {
	return &paymentGateway{}
}

func (pg *paymentGateway) ProcessPayment(cc string) bool {
	time.Sleep(3 * time.Second)
	return true
}
