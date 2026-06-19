package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"mengonten-api/middleware"
	"mengonten-api/worker"
)

func RegisterRoutes(r *gin.Engine, db *gorm.DB, clipWorker *worker.ClipWorker) {
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
		protected.POST("/videos/upload", UploadVideo(db))
		protected.POST("/videos/:video_id/clip", CreateClipJob(db, clipWorker))
		protected.GET("/videos/:video_id/clips", GetVideoClips(db))
		protected.GET("/videos/jobs/:job_id", GetClipJobStatus(db))
	}
}
