package config

import (
	"os"

	"github.com/sashabaranov/go-openai"
)

type OpenAIConfig struct {
	Client *openai.Client
	Model  string
}

func InitOpenAI() *OpenAIConfig {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil
	}

	client := openai.NewClient(apiKey)

	return &OpenAIConfig{
		Client: client,
		Model:  "gpt-4-vision-preview",
	}
}
