package routes

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"mengonten-api/models"
	"mengonten-api/utils"
)

type CreateRuleRequest struct {
	RuleKey     string `json:"rule_key" binding:"required"`
	RuleValue   string `json:"rule_value" binding:"required"`
	Description string `json:"description"`
}

type UpdateRuleRequest struct {
	RuleValue   string `json:"rule_value" binding:"omitempty"`
	Description string `json:"description" binding:"omitempty"`
}

// @Summary Get rules for a plan (admin)
// @Description List semua rules untuk subscription plan tertentu
// @Tags Subscription Rule
// @Produce json
// @Security Bearer
// @Param plan_id path string true "Plan ID"
// @Success 200 {object} utils.Response "Rules list"
// @Router /api/admin/subscription-plans/{plan_id}/rules [get]
func GetPlanRules(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		planID := c.Param("plan_id")
		parsedPlanID, err := uuid.Parse(planID)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid plan ID")
			return
		}

		var rules []models.SubscriptionRule
		if err := db.Where("plan_id = ?", parsedPlanID).Order("rule_key ASC").Find(&rules).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch rules")
			return
		}

		utils.SuccessResponse(c, http.StatusOK, "Rules retrieved", rules)
	}
}

// @Summary Create rule for a plan (admin)
// @Description Tambah rule baru ke subscription plan
// @Tags Subscription Rule
// @Accept json
// @Produce json
// @Security Bearer
// @Param plan_id path string true "Plan ID"
// @Param request body CreateRuleRequest true "Rule data"
// @Success 201 {object} utils.Response "Rule created"
// @Failure 409 {object} utils.Response "Rule key already exists"
// @Router /api/admin/subscription-plans/{plan_id}/rules [post]
func CreatePlanRule(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		planID := c.Param("plan_id")
		parsedPlanID, err := uuid.Parse(planID)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid plan ID")
			return
		}

		var plan models.SubscriptionPlan
		if err := db.First(&plan, parsedPlanID).Error; err != nil {
			utils.ErrorResponse(c, http.StatusNotFound, "Plan not found")
			return
		}

		var req CreateRuleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
			return
		}

		var existing models.SubscriptionRule
		if err := db.Where("plan_id = ? AND rule_key = ?", parsedPlanID, req.RuleKey).First(&existing).Error; err == nil {
			utils.ErrorResponse(c, http.StatusConflict, "Rule key already exists for this plan")
			return
		}

		rule := models.SubscriptionRule{
			PlanID:      parsedPlanID,
			RuleKey:     req.RuleKey,
			RuleValue:   req.RuleValue,
			Description: req.Description,
		}

		if err := db.Create(&rule).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create rule")
			return
		}

		utils.SuccessResponse(c, http.StatusCreated, "Rule created", rule)
	}
}

// @Summary Update rule (admin)
// @Description Update rule value dan description
// @Tags Subscription Rule
// @Accept json
// @Produce json
// @Security Bearer
// @Param rule_id path string true "Rule ID"
// @Param request body UpdateRuleRequest true "Update data"
// @Success 200 {object} utils.Response "Rule updated"
// @Router /api/admin/rules/{rule_id} [put]
func UpdatePlanRule(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		ruleID := c.Param("rule_id")
		parsedID, err := uuid.Parse(ruleID)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid rule ID")
			return
		}

		var rule models.SubscriptionRule
		if err := db.First(&rule, parsedID).Error; err != nil {
			utils.ErrorResponse(c, http.StatusNotFound, "Rule not found")
			return
		}

		var req UpdateRuleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
			return
		}

		updates := map[string]interface{}{}
		if req.RuleValue != "" {
			updates["rule_value"] = req.RuleValue
		}
		if req.Description != "" {
			updates["description"] = req.Description
		}

		db.Model(&rule).Updates(updates)
		db.First(&rule, parsedID)

		utils.SuccessResponse(c, http.StatusOK, "Rule updated", rule)
	}
}

// @Summary Delete rule (admin)
// @Description Hapus rule dari subscription plan
// @Tags Subscription Rule
// @Produce json
// @Security Bearer
// @Param rule_id path string true "Rule ID"
// @Success 200 {object} utils.Response "Rule deleted"
// @Router /api/admin/rules/{rule_id} [delete]
func DeletePlanRule(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		ruleID := c.Param("rule_id")
		parsedID, err := uuid.Parse(ruleID)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid rule ID")
			return
		}

		var rule models.SubscriptionRule
		if err := db.First(&rule, parsedID).Error; err != nil {
			utils.ErrorResponse(c, http.StatusNotFound, "Rule not found")
			return
		}

		if err := db.Delete(&rule).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete rule")
			return
		}

		utils.SuccessResponse(c, http.StatusOK, "Rule deleted successfully", nil)
	}
}

// @Summary Get my active subscription rules
// @Description Cek rules dan usage dari langganan aktif user
// @Tags Subscription Rule
// @Produce json
// @Security Bearer
// @Success 200 {object} utils.Response "Active subscription rules + usage"
// @Router /api/subscription/rules [get]
func GetMySubscriptionRules(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
			return
		}

		var order models.Order
		err := db.Where("user_id = ? AND status = ? AND expired_at > ?",
			userID, "active", time.Now()).
			Order("expired_at DESC").First(&order).Error

		if err != nil {
			utils.SuccessResponse(c, http.StatusOK, "No active subscription", map[string]interface{}{
				"has_subscription": false,
				"rules":            map[string]interface{}{},
				"usage":            map[string]interface{}{},
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

		utils.SuccessResponse(c, http.StatusOK, "Subscription rules retrieved", map[string]interface{}{
			"has_subscription": true,
			"plan_name":        order.ProductName,
			"end_date":         order.ExpiredAt,
			"rules":            rules,
			"usage": map[string]interface{}{
				"storage_limit_mb": storageLimitMB,
			},
		})
	}
}

func GetUserOrder(db *gorm.DB, userID uuid.UUID) (*models.Order, map[string]string) {
	var order models.Order
	if err := db.Where("user_id = ? AND status = ? AND expired_at > ?",
		userID, "active", time.Now()).
		Order("expired_at DESC").First(&order).Error; err != nil {
		return nil, nil
	}

	var orderRules []models.OrderRule
	db.Where("order_id = ?", order.ID).Find(&orderRules)

	rules := map[string]string{}
	for _, r := range orderRules {
		rules[r.RuleKey] = r.RuleValue
	}

	return &order, rules
}
