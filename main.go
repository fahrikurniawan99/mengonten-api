package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
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

	db.AutoMigrate(&models.User{}, &models.YouTubeVideo{}, &models.VideoTranscript{}, &models.VideoSegment{}, &models.ProcessingJob{}, &models.BankAccount{}, &models.SubscriptionPlan{}, &models.SubscriptionRule{}, &models.Transaction{}, &models.UserSubscription{}, &models.PaymentProof{}, &models.PaymentProofPhoto{})

	externalAPIs := config.InitExternalAPIs()
	youtubeProcessor := worker.NewYouTubeProcessor(externalAPIs)

	resendConfig := config.InitResend()
	emailSender := worker.NewEmailSender(resendConfig)

	gin.SetMode(os.Getenv("GIN_MODE"))
	r := gin.Default()

	corsOrigins := strings.Split(os.Getenv("CORS_ORIGINS"), ",")
	if len(corsOrigins) == 0 || corsOrigins[0] == "" {
		corsOrigins = []string{"*"}
	}

	r.Use(cors.New(cors.Config{
		AllowOrigins:     corsOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

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
