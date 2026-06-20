package config

import (
	"os"
)

type ResendConfig struct {
	APIKey    string
	FromEmail string
	FrontendURL string
}

func InitResend() *ResendConfig {
	return &ResendConfig{
		APIKey:      os.Getenv("RESEND_API_KEY"),
		FromEmail:   os.Getenv("RESEND_FROM_EMAIL"),
		FrontendURL: os.Getenv("FRONTEND_URL"),
	}
}
