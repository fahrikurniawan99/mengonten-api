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

	db.AutoMigrate(&models.User{}, &models.YouTubeVideo{}, &models.VideoTranscript{}, &models.VideoSegment{}, &models.ProcessingJob{}, &models.BankAccount{}, &models.SubscriptionPlan{}, &models.SubscriptionRule{}, &models.Transaction{}, &models.Order{}, &models.OrderRule{})

	db.Exec(`CREATE TABLE IF NOT EXISTS orders (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id UUID NOT NULL,
		transaction_id UUID NOT NULL,
		plan_id UUID NOT NULL,
		product_name VARCHAR(255) NOT NULL,
		product_price DOUBLE PRECISION NOT NULL,
		status VARCHAR(50) NOT NULL DEFAULT 'pending',
		expired_at TIMESTAMP,
		created_at TIMESTAMP,
		updated_at TIMESTAMP
	)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS order_rules (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		order_id UUID NOT NULL,
		subscription_rule_id UUID,
		rule_key VARCHAR(255) NOT NULL,
		rule_value VARCHAR(255) NOT NULL,
		created_at TIMESTAMP
	)`)

	db.Migrator().DropTable("user_subscriptions")
	db.Migrator().DropTable("transaction_previews")
	db.Migrator().DropTable("payment_proof_photos")
	db.Migrator().DropTable("payment_proofs")

	db.Exec("ALTER TABLE transactions DROP COLUMN IF EXISTS amount")
	db.Exec("ALTER TABLE transactions DROP COLUMN IF EXISTS unique_code")
	db.Exec("ALTER TABLE transactions DROP COLUMN IF EXISTS total_amount")
	db.Exec("ALTER TABLE transactions DROP COLUMN IF EXISTS bank_name")
	db.Exec("ALTER TABLE transactions DROP COLUMN IF EXISTS bank_account_number")
	db.Exec("ALTER TABLE transactions DROP COLUMN IF EXISTS bank_account_name")
	db.Exec("ALTER TABLE transactions DROP COLUMN IF EXISTS user_subscription_id")
	db.Exec("ALTER TABLE transactions DROP COLUMN IF EXISTS expired_at")
	db.Exec("ALTER TABLE transactions ADD COLUMN IF NOT EXISTS subscription_plan_id UUID")
	db.Exec("ALTER TABLE transactions ADD COLUMN IF NOT EXISTS payment_method VARCHAR(50) DEFAULT ''")
	db.Exec("ALTER TABLE transactions ADD COLUMN IF NOT EXISTS payment_number VARCHAR(255) DEFAULT ''")
	db.Exec("ALTER TABLE transactions DROP COLUMN IF EXISTS payment_url")
	db.Exec("ALTER TABLE transactions DROP COLUMN IF EXISTS payment_code")
	db.Exec("ALTER TABLE transactions DROP COLUMN IF EXISTS duitku_ref")
	db.Exec("ALTER TABLE transactions ADD COLUMN IF NOT EXISTS order_id UUID")
	db.Exec("ALTER TABLE transactions ADD COLUMN IF NOT EXISTS payment_at TIMESTAMP")
	db.Exec("ALTER TABLE transactions ADD COLUMN IF NOT EXISTS payment_total DOUBLE PRECISION DEFAULT 0")
	db.Exec("ALTER TABLE transactions ADD COLUMN IF NOT EXISTS product_name VARCHAR(255) DEFAULT ''")
	db.Exec("ALTER TABLE transactions ADD COLUMN IF NOT EXISTS expired_at TIMESTAMP")
	db.Exec("ALTER TABLE orders ADD COLUMN IF NOT EXISTS last_reminder_sent_at TIMESTAMP")

	db.Exec("ALTER TABLE youtube_videos ADD COLUMN IF NOT EXISTS duration DOUBLE PRECISION DEFAULT 0")
	db.Exec("ALTER TABLE youtube_videos ADD COLUMN IF NOT EXISTS file_size BIGINT DEFAULT 0")

	externalAPIs := config.InitExternalAPIs()
	youtubeProcessor := worker.NewYouTubeProcessor(externalAPIs)

	resendConfig := config.InitResend()
	emailSender := worker.NewEmailSender(resendConfig)

	pakasirConfig := config.InitPakasir()
	pakasirClient := worker.NewPakasirClient(pakasirConfig)

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

	routes.RegisterRoutes(r, db, youtubeProcessor, emailSender, pakasirClient)

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
