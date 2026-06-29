package routes

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"mengonten-api/models"
	"mengonten-api/utils"
)

// @Summary Get my orders (user)
// @Description List semua order/langganan milik user
// @Tags Order
// @Produce json
// @Security Bearer
// @Success 200 {object} utils.Response "My orders"
// @Router /api/orders [get]
func GetMyOrders(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
			return
		}

		var orders []models.Order
		if err := db.Where("user_id = ?", userID).Order("created_at DESC").Find(&orders).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch orders")
			return
		}

		for i := range orders {
			if orders[i].Status == "active" && time.Now().After(orders[i].ExpiredAt) {
				orders[i].Status = "expired"
				db.Model(&orders[i]).Update("status", "expired")
			}
		}

		utils.SuccessResponse(c, http.StatusOK, "Orders retrieved", orders)
	}
}

// @Summary Check active subscription
// @Description Cek langganan aktif + rules + usage
// @Tags Order
// @Produce json
// @Security Bearer
// @Success 200 {object} utils.Response "Subscription status"
// @Router /api/orders/subscription [get]
func CheckActiveSubscription(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
			return
		}

		var order models.Order
		err := db.Where("user_id = ? AND status = ? AND expired_at > ?",
			userID, "active", time.Now()).
			Order("expired_at DESC").
			First(&order).Error

		if err != nil {
			utils.SuccessResponse(c, http.StatusOK, "No active subscription", map[string]interface{}{
				"has_active":   false,
				"subscription": nil,
				"rules":        map[string]string{},
				"usage":        map[string]interface{}{},
			})
			return
		}

		var orderRules []models.OrderRule
		db.Where("order_id = ?", order.ID).Find(&orderRules)

		rules := map[string]string{}
		for _, r := range orderRules {
			rules[r.RuleKey] = r.RuleValue
		}

		storageLimitMB := 0
		if v, ok := rules["max_storage_mb"]; ok {
			storageLimitMB, _ = strconv.Atoi(v)
		}

		var totalBytes int64
		db.Model(&models.YouTubeVideo{}).Where("user_id = ? AND status != ?", userID, "failed").Select("COALESCE(SUM(file_size), 0)").Scan(&totalBytes)
		usedMB := float64(totalBytes) / (1024 * 1024)

		utils.SuccessResponse(c, http.StatusOK, "Active subscription found", map[string]interface{}{
			"has_active":   true,
			"subscription": order,
			"rules":        rules,
			"usage": map[string]interface{}{
				"storage_limit_mb":   storageLimitMB,
				"storage_used_bytes": totalBytes,
				"storage_used_mb":    strconv.FormatFloat(usedMB, 'f', 2, 64),
			},
		})
	}
}
