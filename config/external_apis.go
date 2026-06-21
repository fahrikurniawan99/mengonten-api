package config

import (
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
	return &ExternalAPIs{
		WhisperAPIKey: os.Getenv("OPENAI_API_KEY"),
		GPTAPIKey:     os.Getenv("OPENAI_API_KEY"),
		R2AccountID:   os.Getenv("R2_ACCOUNT_ID"),
		R2AccessKeyID: os.Getenv("R2_ACCESS_KEY_ID"),
		R2SecretKey:   os.Getenv("R2_SECRET_ACCESS_KEY"),
		R2Bucket:      os.Getenv("R2_BUCKET_NAME"),
		R2PublicURL:   os.Getenv("R2_PUBLIC_URL"),
	}
}
