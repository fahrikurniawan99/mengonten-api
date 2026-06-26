package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	ginSwagger "github.com/swaggo/gin-swagger"
	swaggerFiles "github.com/swaggo/files"
	_ "mengonten-api/docs"
	"mengonten-api/config"
	"mengonten-api/middleware"
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

	db.AutoMigrate(&models.User{}, &models.YouTubeVideo{}, &models.VideoTranscript{}, &models.VideoSegment{}, &models.ProcessingJob{}, &models.BankAccount{}, &models.SubscriptionPlan{}, &models.SubscriptionRule{}, &models.Transaction{}, &models.TransactionPreview{}, &models.UserSubscription{}, &models.PaymentProof{}, &models.PaymentProofPhoto{})

	db.Exec("ALTER TABLE transactions DROP COLUMN IF EXISTS subscription_name")
	db.Exec("ALTER TABLE transactions DROP COLUMN IF EXISTS subscription_type")
	db.Exec("ALTER TABLE transactions DROP COLUMN IF EXISTS subscription_price")
	db.Exec("ALTER TABLE transactions DROP COLUMN IF EXISTS subscription_duration")
	db.Exec("ALTER TABLE transactions DROP COLUMN IF EXISTS subscription_benefits")
	db.Exec("ALTER TABLE youtube_videos ADD COLUMN IF NOT EXISTS duration DOUBLE PRECISION DEFAULT 0")
	db.Exec("ALTER TABLE youtube_videos ADD COLUMN IF NOT EXISTS file_size BIGINT DEFAULT 0")
	db.Exec("ALTER TABLE users DROP COLUMN IF EXISTS username")
	db.Exec("ALTER TABLE users DROP COLUMN IF EXISTS password")
	db.Exec("ALTER TABLE users ADD COLUMN IF NOT EXISTS otp_code VARCHAR(255) DEFAULT ''")
	db.Exec("ALTER TABLE users ADD COLUMN IF NOT EXISTS otp_expires_at TIMESTAMP")
	db.Exec("ALTER TABLE users ADD COLUMN IF NOT EXISTS otp_requested_at TIMESTAMP")

	externalAPIs := config.InitExternalAPIs()
	youtubeProcessor := worker.NewYouTubeProcessor(externalAPIs)

	resendConfig := config.InitResend()
	emailSender := worker.NewEmailSender(resendConfig)

	ginMode := os.Getenv("GIN_MODE")
	gin.SetMode(ginMode)
	r := gin.Default()

	corsOrigins := strings.Split(os.Getenv("CORS_ORIGINS"), ",")
	if len(corsOrigins) == 0 || corsOrigins[0] == "" {
		if ginMode == "release" {
			log.Fatal("CORS_ORIGINS must be set in production mode")
		}
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

	r.Use(middleware.RateLimit(100, time.Minute))

	if ginMode != "release" {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	routes.RegisterRoutes(r, db, youtubeProcessor, emailSender)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		fmt.Printf("Server running on port %s (mode: %s)\n", port, ginMode)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
