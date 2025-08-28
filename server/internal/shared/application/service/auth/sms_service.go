package auth

import (
	"fmt"
	"os"
)

// SMSService interface for sending SMS messages
type SMSService interface {
	SendSMS(to, message string) error
}

// ConsoleSMSService is a development implementation that prints SMS to console
type ConsoleSMSService struct{}

func NewConsoleSMSService() *ConsoleSMSService {
	return &ConsoleSMSService{}
}

func (s *ConsoleSMSService) SendSMS(to, message string) error {
	fmt.Printf("SMS to %s: %s\n", to, message)
	return nil
}

// TwilioSMSService is an implementation using Twilio (commented out for now)
type TwilioSMSService struct {
	accountSID string
	authToken  string
	fromNumber string
}

func NewTwilioSMSService() *TwilioSMSService {
	return &TwilioSMSService{
		accountSID: os.Getenv("TWILIO_ACCOUNT_SID"),
		authToken:  os.Getenv("TWILIO_AUTH_TOKEN"),
		fromNumber: os.Getenv("TWILIO_PHONE_NUMBER"),
	}
}

func (s *TwilioSMSService) SendSMS(to, message string) error {
	// TODO: Implement Twilio SMS sending
	// This would require adding the Twilio Go SDK to go.mod
	// Example implementation:
	// client := twilio.NewRestClientWithParams(twilio.ClientParams{
	//     Username: s.accountSID,
	//     Password: s.authToken,
	// })
	// params := &twilio.CreateMessageParams{}
	// params.SetTo(to)
	// params.SetFrom(s.fromNumber)
	// params.SetBody(message)
	// _, err := client.Api.CreateMessage(params)
	// return err

	// For now, just print to console like the development service
	fmt.Printf("SMS to %s: %s\n", to, message)
	return nil
}

// GetSMSService returns the appropriate SMS service based on environment
func GetSMSService() SMSService {
	// Check if Twilio credentials are available
	if os.Getenv("TWILIO_ACCOUNT_SID") != "" && os.Getenv("TWILIO_AUTH_TOKEN") != "" {
		return NewTwilioSMSService()
	}

	// Default to console service for development
	return NewConsoleSMSService()
}
