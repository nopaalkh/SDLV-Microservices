package usecase

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"

	"order-service/internal/repository"
)

// emailService implements EmailService interface
type emailService struct {
	smtpHost string
	smtpPort string
	from     string
}

// NewEmailService creates a new EmailService
func NewEmailService(smtpHost, smtpPort, from string) EmailService {
	return &emailService{
		smtpHost: smtpHost,
		smtpPort: smtpPort,
		from:     from,
	}
}

// SendOrderConfirmation sends an order confirmation email
func (s *emailService) SendOrderConfirmation(ctx context.Context, userEmail string, order *repository.Order) error {
	subject := "Order Confirmation"
	body := fmt.Sprintf(
		"Your order #%s has been processed successfully!\n\n"+
			"Asset: %s\n"+
			"Quantity: %d\n"+
			"Total Price: $%.2f\n\n"+
			"Thank you for your purchase!",
		order.ID, order.AssetID, order.Quantity, order.TotalPrice,
	)

	// Format email message
	message := fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"\r\n"+
			"%s",
		s.from, userEmail, subject, body,
	)

	// Set up authentication
	auth := smtp.PlainAuth("", "", "", s.smtpHost)

	// TLS config — only skip verification for local dev (MailHog)
	isLocalDev := s.smtpHost == "mailhog" || s.smtpHost == "localhost" || s.smtpHost == "127.0.0.1"
	tlsConfig := &tls.Config{
		InsecureSkipVerify: isLocalDev,
		ServerName:         s.smtpHost,
	}

	// Connect to the server
	addr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)
	client, err := smtp.Dial(addr)
	if err != nil {
		log.Printf("Failed to connect to SMTP server: %v", err)
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer client.Close()

	// Start TLS
	if err = client.StartTLS(tlsConfig); err != nil {
		log.Printf("Failed to start TLS: %v", err)
		return fmt.Errorf("failed to start TLS: %w", err)
	}

	// Authenticate
	if err = client.Auth(auth); err != nil {
		log.Printf("Failed to authenticate: %v", err)
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	// Set sender
	if err = client.Mail(s.from); err != nil {
		log.Printf("Failed to set sender: %v", err)
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipient
	if err = client.Rcpt(userEmail); err != nil {
		log.Printf("Failed to set recipient: %v", err)
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	// Send email body
	wc, err := client.Data()
	if err != nil {
		log.Printf("Failed to create data writer: %v", err)
		return fmt.Errorf("failed to create data writer: %w", err)
	}

	_, err = wc.Write([]byte(message))
	if err != nil {
		log.Printf("Failed to write email data: %v", err)
		return fmt.Errorf("failed to write email data: %w", err)
	}

	err = wc.Close()
	if err != nil {
		log.Printf("Failed to close email writer: %v", err)
		return fmt.Errorf("failed to close email writer: %w", err)
	}

	// Log success
	log.Printf("Successfully sent email to %s", userEmail)
	return nil
}
