package worker

import (
	"fmt"
	"log"

	"github.com/resend/resend-go/v2"
	"mengonten-api/config"
)

type EmailSender struct {
	Client    *resend.Client
	FromEmail string
	FrontendURL string
}

func NewEmailSender(cfg *config.ResendConfig) *EmailSender {
	if cfg.APIKey == "" {
		log.Println("RESEND_API_KEY not set, email sending disabled")
		return nil
	}

	client := resend.NewClient(cfg.APIKey)

	return &EmailSender{
		Client:      client,
		FromEmail:   cfg.FromEmail,
		FrontendURL: cfg.FrontendURL,
	}
}

func (es *EmailSender) SendVerificationEmail(toEmail, username, token string) error {
	if es == nil {
		return fmt.Errorf("email sender not configured")
	}

	verificationURL := fmt.Sprintf("%s/verify-email?token=%s", es.FrontendURL, token)

	subject := "Verify Your Email - Mengonten"

	htmlBody := buildVerificationTemplate(username, verificationURL)

	params := &resend.SendEmailRequest{
		From:    es.FromEmail,
		To:      []string{toEmail},
		Subject: subject,
		Html:    htmlBody,
	}

	sent, err := es.Client.Emails.Send(params)
	if err != nil {
		log.Printf("Failed to send verification email to %s: %v", toEmail, err)
		return err
	}

	log.Printf("Verification email sent to %s, ID: %s", toEmail, sent.Id)
	return nil
}
