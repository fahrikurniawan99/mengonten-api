package routes

import (
	"fmt"
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

func GetMySubscriptionRules(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
			return
		}

		var subscription models.UserSubscription
		err := db.Where("user_id = ? AND status = ?", userID, "active").
			Order("end_date DESC").First(&subscription).Error

		if err != nil {
			utils.SuccessResponse(c, http.StatusOK, "No active subscription", map[string]interface{}{
				"has_subscription": false,
				"rules":            map[string]interface{}{},
				"usage":            map[string]interface{}{},
			})
			return
		}

		var plan models.SubscriptionPlan
		var rules []models.SubscriptionRule
		if err := db.Where("name = ?", subscription.PlanName).First(&plan).Error; err == nil {
			db.Where("plan_id = ?", plan.ID).Find(&rules)
		}

		rulesMap := make(map[string]interface{})
		for _, r := range rules {
			rulesMap[r.RuleKey] = r.RuleValue
		}

		storageLimitMB := 0
		if v, ok := rulesMap["max_storage_mb"]; ok {
			storageLimitMB, _ = strconv.Atoi(fmt.Sprintf("%v", v))
		}
		storageUsedMB := float64(subscription.StorageUsedBytes) / (1024 * 1024)

		utils.SuccessResponse(c, http.StatusOK, "Subscription rules retrieved", map[string]interface{}{
			"has_subscription": true,
			"plan_name":        subscription.PlanName,
			"plan_type":        subscription.PlanType,
			"end_date":         subscription.EndDate,
			"rules":            rulesMap,
			"usage": map[string]interface{}{
				"storage_used_bytes": subscription.StorageUsedBytes,
				"storage_used_mb":    fmt.Sprintf("%.2f", storageUsedMB),
				"storage_limit_mb":   storageLimitMB,
			},
		})
	}
}

func GetUserSubscriptionRules(db *gorm.DB, userID uuid.UUID) map[string]string {
	var subscription models.UserSubscription
	if err := db.Where("user_id = ? AND status = ? AND end_date > ?",
		userID, "active", time.Now()).
		Order("end_date DESC").First(&subscription).Error; err != nil {
		return nil
	}

	var plan models.SubscriptionPlan
	if err := db.Where("name = ?", subscription.PlanName).First(&plan).Error; err != nil {
		return nil
	}

	var rules []models.SubscriptionRule
	db.Where("plan_id = ?", plan.ID).Find(&rules)

	rulesMap := make(map[string]string)
	for _, r := range rules {
		rulesMap[r.RuleKey] = r.RuleValue
	}
	return rulesMap
}
