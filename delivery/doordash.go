package delivery

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/google/uuid"
)

// CreateQuoteRequest represents a bare minimum request to create a delivery quote with DoorDash Drive API.
// All fields are required by the DoorDash Drive API.
type CreateQuoteRequest struct {
	// ExternalDeliveryID is a unique identifier for the delivery from your system
	ExternalDeliveryID string `json:"external_delivery_id" binding:"required"`

	// PickupAddress is the full street address of the pickup location (must include city, state, ZIP)
	PickupAddress string `json:"pickup_address" binding:"required"`

	// PickupPhoneNumber is the phone number at pickup location (E.164 format recommended, e.g., +14155552671)
	PickupPhoneNumber string `json:"pickup_phone_number" binding:"required"`

	// DropoffAddress is the full street address of the dropoff location (must include city, state, ZIP)
	DropoffAddress string `json:"dropoff_address" binding:"required"`

	// DropoffPhoneNumber is the phone number at dropoff location (E.164 format recommended, e.g., +14155552671)
	DropoffPhoneNumber string `json:"dropoff_phone_number" binding:"required"`

	// OrderValue is the order value in cents (e.g., $20.00 = 2000)
	OrderValue int64 `json:"order_value" binding:"required,min=0"`
}

// CreateQuoteResponse represents the response from DoorDash Drive API after creating a delivery quote.
type CreateQuoteResponse struct {
	// ExternalDeliveryID is the unique identifier that was sent in the request
	ExternalDeliveryID string `json:"external_delivery_id"`

	// Currency is the currency code (e.g., "USD")
	Currency string `json:"currency"`

	// Fee is the delivery fee in cents
	Fee int64 `json:"fee"`

	// ID is DoorDash's unique identifier for this quote
	ID string `json:"id"`

	// ExpiresAt is the ISO 8601 timestamp when this quote expires
	ExpiresAt string `json:"expires_at"`
}

// DoorDashConfig holds the authentication credentials for DoorDash Drive API.
// These credentials are obtained from the DoorDash Developer Portal.
type DoorDashConfig struct {
	// DeveloperID is your DoorDash developer ID (sent as DD-DEVELOPER-ID header)
	DeveloperID string `json:"developer_id"`

	// KeyID is your DoorDash key ID (sent as DD-KEY-ID header)
	KeyID string `json:"key_id"`

	// SigningSecret is your DoorDash signing secret used to generate HMAC-SHA256 signatures
	SigningSecret string `json:"signing_secret"`
}

// QuoteResult wraps the result of a delivery quote request for channel communication.
// It contains either a successful response or an error.
type QuoteResult struct {
	// Response contains the successful quote response from DoorDash (nil if error occurred)
	Response *CreateQuoteResponse

	// Err contains any error that occurred during the quote request (nil if successful)
	Err error
}

// DeliveryQuoteParams contains all the parameters needed to request a delivery quote.
type DeliveryQuoteParams struct {
	// PickupAddress is the restaurant/store address where the order will be picked up
	PickupAddress string

	// PickupPhoneNumber is the restaurant/store phone number (E.164 format recommended)
	PickupPhoneNumber string

	// DropoffAddress is the customer's delivery address
	DropoffAddress string

	// DropoffPhoneNumber is the customer's phone number (E.164 format recommended)
	DropoffPhoneNumber string

	// OrderValue is the order total in cents (e.g., $20.00 = 2000)
	OrderValue int64
}

// CreateDeliveryQuote creates a delivery quote with DoorDash Drive API.
// This function is designed to be called in a goroutine and communicates results via a channel.
// It respects context cancellation and sends results to resultChan when complete.
func CreateDeliveryQuote(ctx context.Context, params DeliveryQuoteParams, resultChan chan<- QuoteResult) {
	defer close(resultChan)

	// Check if context is already cancelled before starting
	select {
	case <-ctx.Done():
		resultChan <- QuoteResult{Err: ctx.Err()}
		return
	default:
	}

	// Prepare the request payload
	createQuoteReq := CreateQuoteRequest{
		ExternalDeliveryID: uuid.New().String(),
		PickupAddress:      params.PickupAddress,
		PickupPhoneNumber:  params.PickupPhoneNumber,
		DropoffAddress:     params.DropoffAddress,
		DropoffPhoneNumber: params.DropoffPhoneNumber,
		OrderValue:         params.OrderValue,
	}

	// Marshal request to JSON
	jsonData, err := json.Marshal(createQuoteReq)
	if err != nil {
		log.Printf("error serializing delivery data: %s", err.Error())
		resultChan <- QuoteResult{Err: err}
		return
	}

	// Create HTTP request with context for proper cancellation support
	reqUrl := "https://openapi.doordash.com/drive/v2/quotes"
	req, err := http.NewRequestWithContext(ctx, "POST", reqUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("error creating request: %s", err.Error())
		resultChan <- QuoteResult{Err: err}
		return
	}

	// Set required headers
	req.Header.Set("Content-Type", "application/json")
	// TODO: Add authentication headers (DD-DEVELOPER-ID, DD-KEY-ID, Authorization)

	// Make the HTTP request
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Printf("error getting response from doordash: %s", err.Error())
		resultChan <- QuoteResult{Err: err}
		return
	}
	defer res.Body.Close()

	// Read response body
	bodyReader, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("error reading response body: %s", err.Error())
		resultChan <- QuoteResult{Err: err}
		return
	}

	// Unmarshal the response
	createQuoteRes := new(CreateQuoteResponse)
	if err := json.Unmarshal(bodyReader, createQuoteRes); err != nil {
		log.Printf("error unmarshaling response: %s", err.Error())
		resultChan <- QuoteResult{Err: err}
		return
	}

	// Send successful result
	resultChan <- QuoteResult{Response: createQuoteRes, Err: nil}
}
