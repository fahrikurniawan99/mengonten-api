package config

import (
	"os"
)

type ExternalAPIs struct {
	WhisperAPIKey    string
	GPTAPIKey        string
	CloudinaryName   string
	CloudinaryKey    string
	CloudinarySecret string
}

func InitExternalAPIs() *ExternalAPIs {
	return &ExternalAPIs{
		WhisperAPIKey:    os.Getenv("OPENAI_API_KEY"),
		GPTAPIKey:        os.Getenv("OPENAI_API_KEY"),
		CloudinaryName:   os.Getenv("CLOUDINARY_NAME"),
		CloudinaryKey:    os.Getenv("CLOUDINARY_KEY"),
		CloudinarySecret: os.Getenv("CLOUDINARY_SECRET"),
	}
}
