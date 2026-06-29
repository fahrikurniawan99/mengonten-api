package main

import (
	"log"

	"github.com/joho/godotenv"
	"mengonten-api/config"
	"mengonten-api/worker"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	db := config.InitDB()
	if db == nil {
		log.Fatal("Failed to connect to database")
	}

	resendConfig := config.InitResend()
	emailSender := worker.NewEmailSender(resendConfig)

	worker.RunExpiryCheck(db, emailSender)
}
