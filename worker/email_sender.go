package worker

import (
	"fmt"
	"log"

	"github.com/google/uuid"
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

func (es *EmailSender) SendVerificationEmail(toEmail, token string) error {
	if es == nil {
		return fmt.Errorf("email sender not configured")
	}

	verificationURL := fmt.Sprintf("%s/verify-email?token=%s", es.FrontendURL, token)

	subject := "Verifikasi Email - Mengonten"

	htmlBody := buildVerificationTemplate(verificationURL)

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

func (es *EmailSender) SendCheckoutEmail(toEmail, username, referenceID, planName, description string, amount, totalAmount float64, uniqueCode int, planID uuid.UUID) error {
	if es == nil {
		return fmt.Errorf("email sender not configured")
	}

	checkoutURL := fmt.Sprintf("%s/checkout?plan_id=%s", es.FrontendURL, planID.String())
	subject := fmt.Sprintf("Checkout #%s - Mengonten", referenceID)
	htmlBody := buildCheckoutTemplate(username, referenceID, planName, description, amount, totalAmount, uniqueCode, checkoutURL)

	params := &resend.SendEmailRequest{
		From:    es.FromEmail,
		To:      []string{toEmail},
		Subject: subject,
		Html:    htmlBody,
	}

	sent, err := es.Client.Emails.Send(params)
	if err != nil {
		log.Printf("Failed to send checkout email to %s: %v", toEmail, err)
		return err
	}

	log.Printf("Checkout email sent to %s, ID: %s", toEmail, sent.Id)
	return nil
}

func (es *EmailSender) SendOTPEmail(toEmail, otp string) error {
	if es == nil {
		return fmt.Errorf("email sender not configured")
	}

	subject := "Kode OTP Login - Mengonten"
	htmlBody := buildOTPTemplate(otp)

	params := &resend.SendEmailRequest{
		From:    es.FromEmail,
		To:      []string{toEmail},
		Subject: subject,
		Html:    htmlBody,
	}

	sent, err := es.Client.Emails.Send(params)
	if err != nil {
		log.Printf("Failed to send OTP email to %s: %v", toEmail, err)
		return err
	}

	log.Printf("OTP email sent to %s, ID: %s", toEmail, sent.Id)
	return nil
}
