package payment

import "time"

type PaymentData struct {
	CardNumber string
	Cvv        string
	ExpireDate time.Time
}

// PaymentType represents the payment method used
type PaymentType string

const (
	Cash   PaymentType = "Cash"
	Credit PaymentType = "Credit"
	Gift   PaymentType = "Gift"
	Crypto PaymentType = "Crypto"
)

type PaymentGateway interface {
	ProcessPayment(p *PaymentData) bool
}

type paymentGateway struct{}

func NewPaymentGateway() PaymentGateway {
	return &paymentGateway{}
}

func (pg *paymentGateway) ProcessPayment(p *PaymentData) bool {
	time.Sleep(3 * time.Second)
	return true
}
