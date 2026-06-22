package config

import (
	"log"
	"os"
	"time"
)

type JWTConfig struct {
	Secret     string
	Expiration time.Duration
}

func GetJWTConfig() JWTConfig {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	return JWTConfig{
		Secret:     secret,
		Expiration: 24 * time.Hour,
	}
}
