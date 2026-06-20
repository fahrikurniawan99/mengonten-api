package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	ginSwagger "github.com/swaggo/gin-swagger"
	swaggerFiles "github.com/swaggo/files"
	_ "mengonten-api/docs"
	"mengonten-api/config"
	"mengonten-api/models"
	"mengonten-api/routes"
	"mengonten-api/worker"
)

// @title Mengonten API
// @version 1.0
// @description API Documentation
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	db := config.InitDB()
	if db == nil {
		log.Fatal("Failed to connect to database")
	}

	db.AutoMigrate(&models.User{}, &models.YouTubeVideo{}, &models.VideoTranscript{}, &models.VideoSegment{}, &models.ProcessingJob{}, &models.BankAccount{})

	externalAPIs := config.InitExternalAPIs()
	youtubeProcessor := worker.NewYouTubeProcessor(externalAPIs)

	resendConfig := config.InitResend()
	emailSender := worker.NewEmailSender(resendConfig)

	gin.SetMode(os.Getenv("GIN_MODE"))
	r := gin.Default()

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	routes.RegisterRoutes(r, db, youtubeProcessor, emailSender)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server running on port %s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
