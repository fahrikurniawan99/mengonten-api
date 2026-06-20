package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"mengonten-api/models"
	"mengonten-api/utils"
)

type CreateBankAccountRequest struct {
	BankName      string `json:"bank_name" binding:"required"`
	AccountNumber string `json:"account_number" binding:"required"`
	AccountName   string `json:"account_name" binding:"required"`
	Icon          string `json:"icon"`
}

type UpdateBankAccountRequest struct {
	BankName      string `json:"bank_name" binding:"omitempty"`
	AccountNumber string `json:"account_number" binding:"omitempty"`
	AccountName   string `json:"account_name" binding:"omitempty"`
	Icon          string `json:"icon" binding:"omitempty"`
	IsActive      *bool  `json:"is_active" binding:"omitempty"`
}

// @Summary Get all bank accounts (admin)
// @Description List semua bank accounts - admin only
// @Tags Bank Account
// @Produce json
// @Security Bearer
// @Success 200 {object} utils.Response "Bank accounts list"
// @Router /api/admin/bank-accounts [get]
func GetBankAccounts(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var accounts []models.BankAccount
		if err := db.Order("created_at DESC").Find(&accounts).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch bank accounts")
			return
		}
		utils.SuccessResponse(c, http.StatusOK, "Bank accounts retrieved", accounts)
	}
}

// @Summary Get active bank accounts (public)
// @Description List bank accounts yang aktif untuk user
// @Tags Bank Account
// @Produce json
// @Success 200 {object} utils.Response "Active bank accounts"
// @Router /api/bank-accounts [get]
func GetActiveBankAccounts(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var accounts []models.BankAccount
		if err := db.Where("is_active = ?", true).Order("created_at DESC").Find(&accounts).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch bank accounts")
			return
		}
		utils.SuccessResponse(c, http.StatusOK, "Active bank accounts retrieved", accounts)
	}
}

// @Summary Create bank account (admin)
// @Description Tambah rekening bank baru - admin only
// @Tags Bank Account
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body CreateBankAccountRequest true "Bank account data"
// @Success 201 {object} utils.Response "Bank account created"
// @Router /api/admin/bank-accounts [post]
func CreateBankAccount(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateBankAccountRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
			return
		}

		account := models.BankAccount{
			BankName:      req.BankName,
			AccountNumber: req.AccountNumber,
			AccountName:   req.AccountName,
			Icon:          req.Icon,
			IsActive:      true,
		}

		if err := db.Create(&account).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create bank account")
			return
		}

		utils.SuccessResponse(c, http.StatusCreated, "Bank account created", account)
	}
}

// @Summary Update bank account (admin)
// @Description Update rekening bank - admin only
// @Tags Bank Account
// @Accept json
// @Produce json
// @Security Bearer
// @Param account_id path string true "Account ID"
// @Param request body UpdateBankAccountRequest true "Update data"
// @Success 200 {object} utils.Response "Bank account updated"
// @Router /api/admin/bank-accounts/{account_id} [put]
func UpdateBankAccount(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		accountID := c.Param("account_id")
		parsedID, err := uuid.Parse(accountID)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid account ID")
			return
		}

		var account models.BankAccount
		if err := db.First(&account, parsedID).Error; err != nil {
			utils.ErrorResponse(c, http.StatusNotFound, "Bank account not found")
			return
		}

		var req UpdateBankAccountRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
			return
		}

		updates := map[string]interface{}{}
		if req.BankName != "" {
			updates["bank_name"] = req.BankName
		}
		if req.AccountNumber != "" {
			updates["account_number"] = req.AccountNumber
		}
		if req.AccountName != "" {
			updates["account_name"] = req.AccountName
		}
		if req.Icon != "" {
			updates["icon"] = req.Icon
		}
		if req.IsActive != nil {
			updates["is_active"] = *req.IsActive
		}

		db.Model(&account).Updates(updates)

		utils.SuccessResponse(c, http.StatusOK, "Bank account updated", account)
	}
}

// @Summary Delete bank account (admin)
// @Description Hapus rekening bank - admin only
// @Tags Bank Account
// @Produce json
// @Security Bearer
// @Param account_id path string true "Account ID"
// @Success 200 {object} utils.Response "Bank account deleted"
// @Router /api/admin/bank-accounts/{account_id} [delete]
func DeleteBankAccount(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		accountID := c.Param("account_id")
		parsedID, err := uuid.Parse(accountID)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid account ID")
			return
		}

		var account models.BankAccount
		if err := db.First(&account, parsedID).Error; err != nil {
			utils.ErrorResponse(c, http.StatusNotFound, "Bank account not found")
			return
		}

		if err := db.Delete(&account).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete bank account")
			return
		}

		utils.SuccessResponse(c, http.StatusOK, "Bank account deleted successfully", nil)
	}
}
