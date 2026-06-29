package config

import (
	"log"
	"os"
)

type PakasirConfig struct {
	APIKey  string
	Project string
	BaseURL string
}

func InitPakasir() *PakasirConfig {
	apiKey := os.Getenv("PAKASIR_API_KEY")
	project := os.Getenv("PAKASIR_PROJECT")
	env := os.Getenv("PAKASIR_ENV")

	if apiKey == "" || project == "" {
		log.Println("WARNING: Pakasir credentials not set - payment gateway will not work")
	}

	baseURL := "https://app.pakasir.com"
	if env == "production" {
		baseURL = "https://app.pakasir.com"
	}

	return &PakasirConfig{
		APIKey:  apiKey,
		Project: project,
		BaseURL: baseURL,
	}
}
