package config

import (
	"log"
	"os"
)

type ExternalAPIs struct {
	WhisperAPIKey string
	GPTAPIKey     string
	R2AccountID   string
	R2AccessKeyID string
	R2SecretKey   string
	R2Bucket      string
	R2PublicURL   string
}

func InitExternalAPIs() *ExternalAPIs {
	config := &ExternalAPIs{
		WhisperAPIKey: os.Getenv("OPENAI_API_KEY"),
		GPTAPIKey:     os.Getenv("OPENAI_API_KEY"),
		R2AccountID:   os.Getenv("R2_ACCOUNT_ID"),
		R2AccessKeyID: os.Getenv("R2_ACCESS_KEY_ID"),
		R2SecretKey:   os.Getenv("R2_SECRET_ACCESS_KEY"),
		R2Bucket:      os.Getenv("R2_BUCKET_NAME"),
		R2PublicURL:   os.Getenv("R2_PUBLIC_URL"),
	}

	if config.WhisperAPIKey == "" {
		log.Println("WARNING: OPENAI_API_KEY is not set - AI features will not work")
	}
	if config.R2AccountID == "" || config.R2AccessKeyID == "" || config.R2SecretKey == "" {
		log.Println("WARNING: R2 credentials are not set - file uploads will not work")
	}

	return config
}
