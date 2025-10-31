package delivery

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// DeliveryService defines the interface for delivery operations
type DeliveryService interface {
	RequestQuote(ctx context.Context, params DeliveryQuoteParams) (*CreateQuoteResponse, error)
}

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

// generateJWT creates a JWT token for DoorDash API authentication
func (s *DoorDashService) generateJWT() (string, error) {
	// Decode the base64-encoded signing secret
	decodedSecret, err := base64.RawURLEncoding.DecodeString(s.config.SigningSecret)
	if err != nil {
		return "", fmt.Errorf("failed to decode signing secret: %w", err)
	}

	// Set token expiration to 5 minutes from now
	now := time.Now()
	expirationTime := now.Add(5 * time.Minute)

	// Create the JWT claims with DoorDash required fields
	claims := jwt.MapClaims{
		"aud": "doordash",
		"iss": s.config.DeveloperID,
		"kid": s.config.KeyID,
		"exp": expirationTime.Unix(),
		"iat": now.Unix(),
	}

	// Create token with custom header
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Add custom DoorDash header
	token.Header["dd-ver"] = "DD-JWT-V1"
	token.Header["algorithm"] = "HS256"

	// Sign the token with the decoded secret
	tokenString, err := token.SignedString(decodedSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}

	return tokenString, nil
}

// RequestQuote creates a delivery quote with DoorDash Drive API.
// This is a synchronous method that can be called from a goroutine.
func (s *DoorDashService) RequestQuote(ctx context.Context, params DeliveryQuoteParams) (*CreateQuoteResponse, error) {
	log.Printf("entered goroutine call")
	// Check if context is already cancelled before starting
	select {
	case <-ctx.Done():
		log.Printf("triggered done case")
		return nil, ctx.Err()
	default:
		log.Printf("triggered default case")
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

	// Generate JWT for authentication
	jwtToken, err := s.generateJWT()
	if err != nil {
		log.Printf("error generating JWT: %s", err.Error())
		return nil, err
	}

	// Set required headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwtToken))

	// Make the HTTP request
	res, err := s.client.Do(req)
	if err != nil {
		log.Printf("error getting response from doordash: %s", err.Error())
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		bodyReader, _ := io.ReadAll(res.Body)
		log.Printf("DoorDash API error: status=%d, body=%s", res.StatusCode,
			string(bodyReader))
		return nil, fmt.Errorf("doordash API returned status %d: %s", res.StatusCode,
			string(bodyReader))
	}

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

	log.Printf("created quote with id: %v", createQuoteReq.ExternalDeliveryID)
	return createQuoteRes, nil
}
