package usecase

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
)

// EmailService defines the interface for sending emails
type EmailService interface {
	SendPasswordReset(ctx context.Context, userEmail, resetLink string) error
}

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

// SendPasswordReset sends a password reset email
func (s *emailService) SendPasswordReset(ctx context.Context, userEmail, resetLink string) error {
	subject := "Password Reset Request"
	body := fmt.Sprintf(
		"You have requested to reset your password.\n\n"+
			"Please click on the following link to reset your password:\n"+
			"%s\n\n"+
			"If you did not request this, please ignore this email.\n"+
			"This link will expire in 1 hour.",
		resetLink,
	)

	message := fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"\r\n"+
			"%s",
		s.from, userEmail, subject, body,
	)

	auth := smtp.PlainAuth("", "", "", s.smtpHost)

	isLocalDev := s.smtpHost == "mailhog" || s.smtpHost == "localhost" || s.smtpHost == "127.0.0.1"

	addr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)
	client, err := smtp.Dial(addr)
	if err != nil {
		log.Printf("Failed to connect to SMTP server: %v", err)
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer client.Close()

	if !isLocalDev {
		tlsConfig := &tls.Config{
			ServerName: s.smtpHost,
		}
		if err = client.StartTLS(tlsConfig); err != nil {
			log.Printf("Failed to start TLS: %v", err)
			return fmt.Errorf("failed to start TLS: %w", err)
		}

		if err = client.Auth(auth); err != nil {
			log.Printf("Failed to authenticate: %v", err)
			return fmt.Errorf("failed to authenticate: %w", err)
		}
	}

	if err = client.Mail(s.from); err != nil {
		log.Printf("Failed to set sender: %v", err)
		return fmt.Errorf("failed to set sender: %w", err)
	}

	if err = client.Rcpt(userEmail); err != nil {
		log.Printf("Failed to set recipient: %v", err)
		return fmt.Errorf("failed to set recipient: %w", err)
	}

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

	log.Printf("Successfully sent password reset email to %s", userEmail)
	return nil
}
