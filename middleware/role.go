package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"mengonten-api/models"
	"mengonten-api/utils"
)

func RequireRole(db *gorm.DB, roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
			c.Abort()
			return
		}

		var user models.User
		if err := db.First(&user, userID).Error; err != nil {
			utils.ErrorResponse(c, http.StatusNotFound, "User not found")
			c.Abort()
			return
		}

		for _, role := range roles {
			if user.Role == role {
				c.Set("user_role", user.Role)
				c.Next()
				return
			}
		}

		utils.ErrorResponse(c, http.StatusForbidden, "Access denied. Required role: "+roles[0])
		c.Abort()
	}
}
