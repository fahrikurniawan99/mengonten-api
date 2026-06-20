package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"mengonten-api/middleware"
	"mengonten-api/worker"
)

func RegisterRoutes(r *gin.Engine, db *gorm.DB, youtubeProcessor *worker.YouTubeProcessor) {
	auth := r.Group("/api/auth")
	{
		auth.POST("/register", Register(db))
		auth.POST("/login", Login(db))
		auth.POST("/logout", Logout)
	}

	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/profile", GetProfile(db))
		protected.POST("/youtube/submit", SubmitYouTubeVideo(db, youtubeProcessor))
		protected.GET("/youtube/:video_id", GetYouTubeVideo(db))
		protected.GET("/youtube/jobs/:job_id", GetProcessingJobStatus(db))
		protected.DELETE("/youtube/segments/:segment_id", DeleteVideoSegment(db))
	}
}
