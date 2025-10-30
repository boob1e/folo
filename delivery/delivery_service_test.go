package delivery

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

func TestRequestQuote_RespectsContextCancellation(t *testing.T) {
	config := DoorDashConfig{
		DeveloperID:   "test-dev-id",
		KeyID:         "test-key-id",
		SigningSecret: "dGVzdC1zaWduaW5nLXNlY3JldA==",
	}

	service := NewDoorDashService(config)

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	params := DeliveryQuoteParams{
		PickupAddress:      "123 Test St",
		PickupPhoneNumber:  "+14155551234",
		DropoffAddress:     "456 Test Ave",
		DropoffPhoneNumber: "+14155555678",
		OrderValue:         2000,
	}

	// Call your function
	_, err := service.RequestQuote(ctx, params)

	// Assert that it returned context.Canceled error
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestRequestQuote_RespectsTimeout(t *testing.T) {
	godotenv.Load(".env")
	doorDashConfig := DoorDashConfig{
		DeveloperID:   os.Getenv("DOORDASH_DEVELOPER_ID"),
		KeyID:         os.Getenv("DOORDASH_KEY_ID"),
		SigningSecret: os.Getenv("DOORDASH_SIGNING_SECRET"),
	}
	service := NewDoorDashService(doorDashConfig)

	// Context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(10 * time.Millisecond) // Ensure timeout happens

	params := DeliveryQuoteParams{
		PickupAddress:      "123 Test St",
		PickupPhoneNumber:  "+14155551234",
		DropoffAddress:     "456 Test Ave",
		DropoffPhoneNumber: "+14155555678",
		OrderValue:         2000,
	}

	_, err := service.RequestQuote(ctx, params)

	// Assert timeout error
	if err != context.DeadlineExceeded {
		t.Errorf("expected context.DeadlineExceeded, got %v", err)
	}
}

func TestRequestQuote_Succeeds(t *testing.T) {
	godotenv.Load("../.env")
	doorDashConfig := DoorDashConfig{
		DeveloperID:   os.Getenv("DOORDASH_DEVELOPER_ID"),
		KeyID:         os.Getenv("DOORDASH_KEY_ID"),
		SigningSecret: os.Getenv("DOORDASH_SIGNING_SECRET"),
	}
	service := NewDoorDashService(doorDashConfig)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	params := DeliveryQuoteParams{
		PickupAddress:      "303 2nd Street, San Francisco, CA 94105",
		PickupPhoneNumber:  "+14155551234",
		DropoffAddress:     "5 Embarcadero Ctr, San Francisco, CA 94111",
		DropoffPhoneNumber: "+14155555678",
		OrderValue:         2000,
	}

	deliveryQuote, err := service.RequestQuote(ctx, params)

	t.Logf("response id from doordash %v", deliveryQuote.ID)

	// Assert timeout error
	if err != nil {
		t.Errorf("Expected happy path, got %v", err)
	}
}
