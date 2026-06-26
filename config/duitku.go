package config

import (
	"log"
	"os"
)

type DuitkuConfig struct {
	MerchantCode string
	APIKey       string
	MerchantKey  string
	BaseURL      string
	CallbackURL  string
	ReturnURL    string
}

func InitDuitku() *DuitkuConfig {
	merchantCode := os.Getenv("DUITKU_MERCHANT_CODE")
	apiKey := os.Getenv("DUITKU_API_KEY")
	merchantKey := os.Getenv("DUITKU_MERCHANT_KEY")
	env := os.Getenv("DUITKU_ENV")

	if merchantCode == "" || apiKey == "" || merchantKey == "" {
		log.Println("WARNING: Duitku credentials not set - payment gateway will not work")
	}

	baseURL := "https://api-sandbox.duitku.com"
	if env == "production" {
		baseURL = "https://api.duitku.com"
	}

	return &DuitkuConfig{
		MerchantCode: merchantCode,
		APIKey:       apiKey,
		MerchantKey:  merchantKey,
		BaseURL:      baseURL,
		CallbackURL:  "https://api-mengonten.tiroe.io/callback/duitku",
		ReturnURL:    "https://mengonten.tiroe.io/duitku/redirect",
	}
}
