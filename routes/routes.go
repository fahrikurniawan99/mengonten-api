package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"mengonten-api/middleware"
	"mengonten-api/worker"
)

func RegisterRoutes(r *gin.Engine, db *gorm.DB, youtubeProcessor *worker.YouTubeProcessor, emailSender *worker.EmailSender) {
	auth := r.Group("/api/auth")
	{
		auth.POST("/register", Register(db, emailSender))
		auth.POST("/login", Login(db))
		auth.POST("/logout", Logout)
		auth.POST("/verify-email", VerifyEmail(db, emailSender))
		auth.POST("/resend-verification", ResendVerification(db, emailSender))
	}

	user := r.Group("/api")
	user.Use(middleware.AuthMiddleware())
	{
		user.GET("/profile", GetProfile(db))
		user.POST("/youtube/submit", middleware.RequireActiveSubscription(db), SubmitYouTubeVideo(db, youtubeProcessor))
		user.GET("/youtube/:video_id", GetYouTubeVideo(db))
		user.GET("/youtube/jobs/:job_id", GetProcessingJobStatus(db))
		user.DELETE("/youtube/segments/:segment_id", DeleteVideoSegment(db))

		user.POST("/transactions", CreateTransaction(db))
		user.GET("/transactions", GetMyTransactions(db))
		user.GET("/subscription/check", CheckActiveSubscription(db))
		user.GET("/subscription/history", GetSubscriptionHistory(db))
	}

	r.GET("/api/bank-accounts", GetActiveBankAccounts(db))
	r.GET("/api/subscription-plans", GetActivePlans(db))

	admin := r.Group("/api/admin")
	admin.Use(middleware.AuthMiddleware())
	admin.Use(middleware.RequireRole(db, "admin"))
	{
		admin.GET("/users", GetUsers(db))
		admin.GET("/users/:user_id", GetUserByID(db))
		admin.PUT("/users/:user_id/role", UpdateUserRole(db))
		admin.DELETE("/users/:user_id", DeleteUser(db))

		admin.GET("/bank-accounts", GetBankAccounts(db))
		admin.POST("/bank-accounts", CreateBankAccount(db))
		admin.PUT("/bank-accounts/:account_id", UpdateBankAccount(db))
		admin.DELETE("/bank-accounts/:account_id", DeleteBankAccount(db))

		admin.GET("/subscription-plans", GetPlansAdmin(db))
		admin.POST("/subscription-plans", CreatePlan(db))
		admin.PUT("/subscription-plans/:plan_id", UpdatePlan(db))
		admin.DELETE("/subscription-plans/:plan_id", DeletePlan(db))

		admin.GET("/transactions", GetAllTransactions(db))
		admin.PUT("/transactions/:transaction_id/status", UpdateTransactionStatus(db))
	}
}
