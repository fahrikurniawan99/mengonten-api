package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"mengonten-api/middleware"
)

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
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
	}
}
