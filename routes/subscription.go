package routes

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"mengonten-api/models"
	"mengonten-api/utils"
)

// @Summary Check active subscription
// @Description Cek apakah user masih punya langganan aktif
// @Tags Subscription
// @Produce json
// @Security Bearer
// @Success 200 {object} utils.Response "Subscription status"
// @Router /api/subscription/check [get]
func CheckActiveSubscription(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
			return
		}

		var subscription models.UserSubscription
		err := db.Where("user_id = ? AND status = ? AND end_date > ?",
			userID, "active", time.Now()).
			Order("end_date DESC").
			First(&subscription).Error

		if err != nil {
			utils.SuccessResponse(c, http.StatusOK, "No active subscription", map[string]interface{}{
				"has_active": false,
				"status":     "none",
			})
			return
		}

		subscription.PrepareResponse()

		utils.SuccessResponse(c, http.StatusOK, "Active subscription found", map[string]interface{}{
			"has_active":   true,
			"status":       "active",
			"subscription": subscription,
		})
	}
}

// @Summary Get subscription history
// @Description History langganan yang sudah di-order user
// @Tags Subscription
// @Produce json
// @Security Bearer
// @Success 200 {object} utils.Response "Subscription history"
// @Router /api/subscription/history [get]
func GetSubscriptionHistory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
			return
		}

		var subscriptions []models.UserSubscription
		if err := db.Where("user_id = ?", userID).Order("created_at DESC").Find(&subscriptions).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch subscriptions")
			return
		}

		for i := range subscriptions {
			subscriptions[i].PrepareResponse()
			if subscriptions[i].Status == "active" && time.Now().After(subscriptions[i].EndDate) {
				subscriptions[i].Status = "expired"
				db.Model(&subscriptions[i]).Update("status", "expired")
			}
		}

		utils.SuccessResponse(c, http.StatusOK, "Subscription history retrieved", subscriptions)
	}
}
