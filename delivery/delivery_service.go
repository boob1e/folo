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

// DoorDashService handles DoorDash API interactions
type DoorDashService struct {
	config DoorDashConfig
	client *http.Client
}

// NewDoorDashService creates a new DoorDash service
func NewDoorDashService(config DoorDashConfig) *DoorDashService {
	return &DoorDashService{
		config: config,
		client: &http.Client{},
	}
}

// RequestQuote creates a delivery quote with DoorDash Drive API.
// This is a synchronous method that can be called from a goroutine.
func (s *DoorDashService) RequestQuote(ctx context.Context, params DeliveryQuoteParams) (*CreateQuoteResponse, error) {
	// Check if context is already cancelled before starting
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
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
		return nil, err
	}

	// Create HTTP request with context for proper cancellation support
	reqUrl := "https://openapi.doordash.com/drive/v2/quotes"
	req, err := http.NewRequestWithContext(ctx, "POST", reqUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("error creating request: %s", err.Error())
		return nil, err
	}

	// Set required headers
	req.Header.Set("Content-Type", "application/json")
	// TODO: Add authentication headers (DD-DEVELOPER-ID, DD-KEY-ID, Authorization)
	// req.Header.Set("DD-DEVELOPER-ID", s.config.DeveloperID)
	// req.Header.Set("DD-KEY-ID", s.config.KeyID)
	// req.Header.Set("Authorization", s.generateSignature(req))

	// Make the HTTP request
	res, err := s.client.Do(req)
	if err != nil {
		log.Printf("error getting response from doordash: %s", err.Error())
		return nil, err
	}
	defer res.Body.Close()

	// Read response body
	bodyReader, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("error reading response body: %s", err.Error())
		return nil, err
	}

	// Unmarshal the response
	createQuoteRes := new(CreateQuoteResponse)
	if err := json.Unmarshal(bodyReader, createQuoteRes); err != nil {
		log.Printf("error unmarshaling response: %s", err.Error())
		return nil, err
	}

	return createQuoteRes, nil
}

// RequestQuoteAsync creates a delivery quote with DoorDash Drive API asynchronously.
// This function is designed to be called in a goroutine and communicates results via a channel.
// It respects context cancellation and sends results to resultChan when complete.
func (s *DoorDashService) RequestQuoteAsync(ctx context.Context, params DeliveryQuoteParams, resultChan chan<- QuoteResult) {
	defer close(resultChan)

	response, err := s.RequestQuote(ctx, params)

	resultChan <- QuoteResult{
		Response: response,
		Err:      err,
	}
}
