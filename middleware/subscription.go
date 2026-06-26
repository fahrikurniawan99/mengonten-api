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

		var order models.Order
		err := db.Where("user_id = ? AND status = ? AND expired_at > ?",
			userID, "active", time.Now()).
			Order("expired_at DESC").
			First(&order).Error

		if err != nil {
			utils.ErrorResponse(c, http.StatusPaymentRequired, "Active subscription required. Please subscribe to use this feature.")
			c.Abort()
			return
		}

		var rules []models.OrderRule
		db.Where("order_id = ?", order.ID).Find(&rules)

		c.Set("order", order)
		c.Set("order_rules", rules)
		c.Next()
	}
}
