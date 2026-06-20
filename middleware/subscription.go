package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"mengonten-api/models"
	"mengonten-api/utils"
)

func RequireActiveSubscription(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
			c.Abort()
			return
		}

		var subscription models.UserSubscription
		err := db.Where("user_id = ? AND status = ? AND end_date > ?",
			userID, "active", time.Now()).
			Order("end_date DESC").
			First(&subscription).Error

		if err != nil {
			utils.ErrorResponse(c, http.StatusPaymentRequired, "Active subscription required. Please subscribe to use this feature.")
			c.Abort()
			return
		}

		c.Set("subscription", subscription)
		c.Next()
	}
}
