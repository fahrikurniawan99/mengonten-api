package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"mengonten-api/models"
	"mengonten-api/utils"
)

type CreatePlanRequest struct {
	Name            string  `json:"name" binding:"required"`
	Description     string  `json:"description"`
	Benefits        string  `json:"benefits" binding:"required"`
	FinalPrice      float64 `json:"final_price" binding:"required,gt=0"`
	DiscountPercent float64 `json:"discount_percent" binding:"omitempty,min=0,max=100"`
	Type            string  `json:"type" binding:"required"`
	DurationDays    int     `json:"duration_days" binding:"required,gt=0"`
	SortOrder       int     `json:"sort_order"`
}

type UpdatePlanRequest struct {
	Name            string  `json:"name" binding:"omitempty"`
	Description     string  `json:"description" binding:"omitempty"`
	Benefits        string  `json:"benefits" binding:"omitempty"`
	FinalPrice      float64 `json:"final_price" binding:"omitempty,gt=0"`
	DiscountPercent float64 `json:"discount_percent" binding:"omitempty,min=0,max=100"`
	Type            string  `json:"type" binding:"omitempty"`
	DurationDays    int     `json:"duration_days" binding:"omitempty,gt=0"`
	IsActive        *bool   `json:"is_active" binding:"omitempty"`
	SortOrder       int     `json:"sort_order" binding:"omitempty"`
}

// @Summary Get all subscription plans (admin)
// @Description List semua subscription plans - admin only
// @Tags Subscription Plan
// @Produce json
// @Security Bearer
// @Success 200 {object} utils.Response "Plans list"
// @Router /api/admin/subscription-plans [get]
func GetPlansAdmin(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var plans []models.SubscriptionPlan
		if err := db.Order("sort_order ASC, created_at DESC").Find(&plans).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch plans")
			return
		}
		for i := range plans {
			plans[i].PrepareResponse()
		}
		utils.SuccessResponse(c, http.StatusOK, "Plans retrieved", plans)
	}
}

// @Summary Get active subscription plans (public)
// @Description List plans yang aktif untuk user
// @Tags Subscription Plan
// @Produce json
// @Success 200 {object} utils.Response "Active plans"
// @Router /api/subscription-plans [get]
func GetActivePlans(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var plans []models.SubscriptionPlan
		if err := db.Where("is_active = ?", true).Order("sort_order ASC").Find(&plans).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch plans")
			return
		}
		for i := range plans {
			plans[i].PrepareResponse()
		}
		utils.SuccessResponse(c, http.StatusOK, "Active plans retrieved", plans)
	}
}

// @Summary Create subscription plan (admin)
// @Description Tambah plan baru - admin only
// @Tags Subscription Plan
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body CreatePlanRequest true "Plan data"
// @Success 201 {object} utils.Response "Plan created"
// @Router /api/admin/subscription-plans [post]
func CreatePlan(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreatePlanRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
			return
		}

		plan := models.SubscriptionPlan{
			Name:            req.Name,
			Description:     req.Description,
			Benefits:        req.Benefits,
			Price:           req.FinalPrice,
			DiscountPercent: req.DiscountPercent,
			Type:            req.Type,
			DurationDays:    req.DurationDays,
			SortOrder:       req.SortOrder,
			IsActive:        true,
		}

		if req.DiscountPercent > 0 {
			plan.FinalPrice = req.FinalPrice / (1 - req.DiscountPercent/100)
		} else {
			plan.FinalPrice = req.FinalPrice
		}

		if err := db.Create(&plan).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create plan")
			return
		}

		plan.PrepareResponse()
		utils.SuccessResponse(c, http.StatusCreated, "Plan created", plan)
	}
}

// @Summary Update subscription plan (admin)
// @Description Update plan - admin only
// @Tags Subscription Plan
// @Accept json
// @Produce json
// @Security Bearer
// @Param plan_id path string true "Plan ID"
// @Param request body UpdatePlanRequest true "Update data"
// @Success 200 {object} utils.Response "Plan updated"
// @Router /api/admin/subscription-plans/{plan_id} [put]
func UpdatePlan(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		planID := c.Param("plan_id")
		parsedID, err := uuid.Parse(planID)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid plan ID")
			return
		}

		var plan models.SubscriptionPlan
		if err := db.First(&plan, parsedID).Error; err != nil {
			utils.ErrorResponse(c, http.StatusNotFound, "Plan not found")
			return
		}

		var req UpdatePlanRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
			return
		}

		updates := map[string]interface{}{}
		if req.Name != "" {
			updates["name"] = req.Name
		}
		if req.Description != "" {
			updates["description"] = req.Description
		}
		if req.Benefits != "" {
			updates["benefits"] = req.Benefits
		}
		if req.FinalPrice > 0 {
			updates["price"] = req.FinalPrice
			discount := req.DiscountPercent
			if discount == 0 {
				discount = plan.DiscountPercent
			}
			if discount > 0 {
				updates["final_price"] = req.FinalPrice / (1 - discount/100)
			} else {
				updates["final_price"] = req.FinalPrice
			}
		}
		if req.DiscountPercent > 0 && req.FinalPrice == 0 {
			updates["discount_percent"] = req.DiscountPercent
			sellingPrice := plan.Price
			updates["final_price"] = sellingPrice / (1 - req.DiscountPercent/100)
		} else if req.DiscountPercent > 0 {
			updates["discount_percent"] = req.DiscountPercent
		}
		if req.Type != "" {
			updates["type"] = req.Type
		}
		if req.DurationDays > 0 {
			updates["duration_days"] = req.DurationDays
		}
		if req.IsActive != nil {
			updates["is_active"] = *req.IsActive
		}
		if req.SortOrder > 0 {
			updates["sort_order"] = req.SortOrder
		}

		db.Model(&plan).Updates(updates)
		db.First(&plan, parsedID)
		plan.PrepareResponse()

		utils.SuccessResponse(c, http.StatusOK, "Plan updated", plan)
	}
}

// @Summary Delete subscription plan (admin)
// @Description Hapus plan - admin only
// @Tags Subscription Plan
// @Produce json
// @Security Bearer
// @Param plan_id path string true "Plan ID"
// @Success 200 {object} utils.Response "Plan deleted"
// @Router /api/admin/subscription-plans/{plan_id} [delete]
func DeletePlan(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		planID := c.Param("plan_id")
		parsedID, err := uuid.Parse(planID)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid plan ID")
			return
		}

		var plan models.SubscriptionPlan
		if err := db.First(&plan, parsedID).Error; err != nil {
			utils.ErrorResponse(c, http.StatusNotFound, "Plan not found")
			return
		}

		if err := db.Delete(&plan).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete plan")
			return
		}

		utils.SuccessResponse(c, http.StatusOK, "Plan deleted successfully", nil)
	}
}
