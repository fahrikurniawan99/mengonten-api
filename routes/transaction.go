package routes

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"mengonten-api/models"
	"mengonten-api/utils"
)

type CreateTransactionRequest struct {
	SubscriptionPlanID uuid.UUID `json:"subscription_plan_id" binding:"required"`
	BankAccountID      uuid.UUID `json:"bank_account_id" binding:"required"`
}

type UpdateTransactionStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=paid expired cancelled"`
}

func generateUniqueCode() int {
	return 100 + rand.Intn(900)
}

func generateReferenceID() string {
	return fmt.Sprintf("TXN-%d-%s", time.Now().UnixNano(), uuid.New().String()[:8])
}

func mapStatus(status string) string {
	switch status {
	case "pending":
		return "Menunggu Pembayaran"
	case "paid":
		return "Dibayar"
	case "expired":
		return "Kadaluarsa"
	case "cancelled":
		return "Dibatalkan"
	default:
		return status
	}
}

// @Summary Create transaction (user)
// @Description Buat transaksi baru untuk langganan
// @Tags Transaction
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body CreateTransactionRequest true "Transaction data"
// @Success 201 {object} utils.Response "Transaction created"
// @Router /api/transactions [post]
func CreateTransaction(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
			return
		}

		var req CreateTransactionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
			return
		}

		var plan models.SubscriptionPlan
		if err := db.First(&plan, req.SubscriptionPlanID).Error; err != nil {
			utils.ErrorResponse(c, http.StatusNotFound, "Subscription plan not found")
			return
		}

		var bank models.BankAccount
		if err := db.First(&bank, req.BankAccountID).Error; err != nil {
			utils.ErrorResponse(c, http.StatusNotFound, "Bank account not found")
			return
		}

		if !plan.IsActive || !bank.IsActive {
			utils.ErrorResponse(c, http.StatusBadRequest, "Plan or bank account is not active")
			return
		}

		price := plan.Price
		if plan.DiscountPercent > 0 {
			price = price - (price * plan.DiscountPercent / 100)
		}

		uniqueCode := generateUniqueCode()
		totalAmount := price + float64(uniqueCode)

		expireAt := time.Now().Add(24 * time.Hour)

		transaction := models.Transaction{
			UserID:               userID.(uuid.UUID),
			ReferenceID:          generateReferenceID(),
			Status:               "pending",
			Amount:               price,
			UniqueCode:           uniqueCode,
			TotalAmount:          totalAmount,
			BankName:             bank.BankName,
			BankAccountNumber:    bank.AccountNumber,
			BankAccountName:      bank.AccountName,
			SubscriptionName:     plan.Name,
			SubscriptionType:     plan.Type,
			SubscriptionPrice:    price,
			SubscriptionDuration: plan.DurationDays,
			SubscriptionBenefits: plan.Benefits,
			ExpiredAt:            &expireAt,
		}

		if err := db.Create(&transaction).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create transaction")
			return
		}

		transaction.PrepareResponse()
		utils.SuccessResponse(c, http.StatusCreated, "Transaction created", transaction)
	}
}

// @Summary Get my transactions (user)
// @Description List transaksi milik user sendiri
// @Tags Transaction
// @Produce json
// @Security Bearer
// @Success 200 {object} utils.Response "My transactions"
// @Router /api/transactions [get]
func GetMyTransactions(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
			return
		}

		var transactions []models.Transaction
		if err := db.Where("user_id = ?", userID).Order("created_at DESC").Find(&transactions).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch transactions")
			return
		}

		for i := range transactions {
			transactions[i].PrepareResponse()
		}

		utils.SuccessResponse(c, http.StatusOK, "Transactions retrieved", transactions)
	}
}

// @Summary Get all transactions (admin)
// @Description List semua transaksi - admin only
// @Tags Admin Transaction
// @Produce json
// @Security Bearer
// @Success 200 {object} utils.Response "All transactions"
// @Router /api/admin/transactions [get]
func GetAllTransactions(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var transactions []models.Transaction
		if err := db.Order("created_at DESC").Find(&transactions).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch transactions")
			return
		}

		for i := range transactions {
			transactions[i].PrepareResponse()
		}

		utils.SuccessResponse(c, http.StatusOK, "Transactions retrieved", transactions)
	}
}

// @Summary Update transaction status (admin)
// @Description Update status transaksi - admin only. Jika paid, otomatis buat subscription
// @Tags Admin Transaction
// @Accept json
// @Produce json
// @Security Bearer
// @Param transaction_id path string true "Transaction ID"
// @Param request body UpdateTransactionStatusRequest true "New status"
// @Success 200 {object} utils.Response "Transaction updated"
// @Router /api/admin/transactions/{transaction_id}/status [put]
func UpdateTransactionStatus(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		transactionID := c.Param("transaction_id")
		parsedID, err := uuid.Parse(transactionID)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid transaction ID")
			return
		}

		var req UpdateTransactionStatusRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
			return
		}

		var transaction models.Transaction
		if err := db.First(&transaction, parsedID).Error; err != nil {
			utils.ErrorResponse(c, http.StatusNotFound, "Transaction not found")
			return
		}

		if transaction.Status != "pending" {
			utils.ErrorResponse(c, http.StatusBadRequest, "Transaction is not in pending status")
			return
		}

		updates := map[string]interface{}{
			"status": req.Status,
		}

		if req.Status == "paid" {
			now := time.Now()
			updates["paid_at"] = &now
		}

		db.Model(&transaction).Updates(updates)

		if req.Status == "paid" {
			createUserSubscription(db, &transaction)
		}

		transaction.Status = req.Status
		transaction.PrepareResponse()
		utils.SuccessResponse(c, http.StatusOK, "Transaction updated", transaction)
	}
}

func createUserSubscription(db *gorm.DB, transaction *models.Transaction) {
	startDate := time.Now()
	endDate := startDate.AddDate(0, 0, transaction.SubscriptionDuration)

	subscription := models.UserSubscription{
		UserID:        transaction.UserID,
		TransactionID: transaction.ID,
		PlanName:      transaction.SubscriptionName,
		PlanType:      transaction.SubscriptionType,
		PlanBenefits:  transaction.SubscriptionBenefits,
		PlanPrice:     transaction.SubscriptionPrice,
		PlanDuration:  transaction.SubscriptionDuration,
		Status:        "active",
		StartDate:     startDate,
		EndDate:       endDate,
	}

	if err := db.Create(&subscription).Error; err != nil {
		fmt.Printf("Failed to create subscription for transaction %s: %v\n", transaction.ID, err)
	}
}

func parseBenefitsString(benefits string) []string {
	if benefits == "" {
		return []string{}
	}
	parts := strings.Split(benefits, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

type PreviewTransactionRequest struct {
	SubscriptionPlanID uuid.UUID `json:"subscription_plan_id" binding:"required"`
	BankAccountID      uuid.UUID `json:"bank_account_id" binding:"required"`
}

type ConfirmTransactionRequest struct {
	ReferenceID string `json:"reference_id" binding:"required"`
}

func generatePreviewReferenceID() string {
	return fmt.Sprintf("PREV-%d-%s", time.Now().UnixNano(), uuid.New().String()[:8])
}

// @Summary Generate payment preview (user)
// @Description Generate preview pembayaran dengan unique code (berlaku 15 menit)
// @Tags Transaction
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body PreviewTransactionRequest true "Preview data"
// @Success 200 {object} utils.Response "Payment preview"
// @Router /api/transactions/preview [post]
func PreviewTransaction(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
			return
		}

		var req PreviewTransactionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
			return
		}

		var plan models.SubscriptionPlan
		if err := db.First(&plan, req.SubscriptionPlanID).Error; err != nil {
			utils.ErrorResponse(c, http.StatusNotFound, "Subscription plan not found")
			return
		}

		var bank models.BankAccount
		if err := db.First(&bank, req.BankAccountID).Error; err != nil {
			utils.ErrorResponse(c, http.StatusNotFound, "Bank account not found")
			return
		}

		if !plan.IsActive || !bank.IsActive {
			utils.ErrorResponse(c, http.StatusBadRequest, "Plan or bank account is not active")
			return
		}

		db.Where("user_id = ?", userID).Delete(&models.TransactionPreview{})

		price := plan.Price
		if plan.DiscountPercent > 0 {
			price = price - (price * plan.DiscountPercent / 100)
		}

		uniqueCode := generateUniqueCode()
		totalAmount := price + float64(uniqueCode)
		expiresAt := time.Now().Add(15 * time.Minute)

		preview := models.TransactionPreview{
			ReferenceID:          generatePreviewReferenceID(),
			UserID:               userID.(uuid.UUID),
			SubscriptionPlanID:   req.SubscriptionPlanID,
			BankAccountID:        req.BankAccountID,
			Amount:               price,
			UniqueCode:           uniqueCode,
			TotalAmount:          totalAmount,
			BankName:             bank.BankName,
			BankAccountNumber:    bank.AccountNumber,
			BankAccountName:      bank.AccountName,
			SubscriptionName:     plan.Name,
			SubscriptionType:     plan.Type,
			SubscriptionDuration: plan.DurationDays,
			SubscriptionBenefits: plan.Benefits,
			ExpiresAt:            expiresAt,
		}

		if err := db.Create(&preview).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create preview")
			return
		}

		preview.PrepareResponse()
		utils.SuccessResponse(c, http.StatusOK, "Payment preview generated", map[string]interface{}{
			"reference_id":         preview.ReferenceID,
			"amount":               preview.Amount,
			"unique_code":          preview.UniqueCode,
			"total_amount":         preview.TotalAmount,
			"bank_name":            preview.BankName,
			"bank_account_number":  preview.BankAccountNumber,
			"bank_account_name":    preview.BankAccountName,
			"subscription_name":    preview.SubscriptionName,
			"expires_at":           preview.ExpiresAt,
			"expires_in_seconds":   int(time.Until(preview.ExpiresAt).Seconds()),
		})
	}
}

// @Summary Check preview status (user)
// @Description Cek apakah preview masih valid
// @Tags Transaction
// @Produce json
// @Security Bearer
// @Param reference_id path string true "Preview Reference ID"
// @Success 200 {object} utils.Response "Preview status"
// @Router /api/transactions/preview/{reference_id} [get]
func GetPreviewStatus(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		referenceID := c.Param("reference_id")

		var preview models.TransactionPreview
		if err := db.Where("reference_id = ?", referenceID).First(&preview).Error; err != nil {
			utils.SuccessResponse(c, http.StatusOK, "Preview not found", map[string]interface{}{
				"valid":   false,
				"message": "Preview not found",
			})
			return
		}

		if preview.IsExpired() {
			db.Delete(&preview)
			utils.SuccessResponse(c, http.StatusOK, "Preview expired", map[string]interface{}{
				"valid":   false,
				"message": "Preview has expired. Please create a new one.",
			})
			return
		}

		utils.SuccessResponse(c, http.StatusOK, "Preview is valid", map[string]interface{}{
			"valid":              true,
			"total_amount":       preview.TotalAmount,
			"expires_at":         preview.ExpiresAt,
			"expires_in_seconds": int(time.Until(preview.ExpiresAt).Seconds()),
		})
	}
}

// @Summary Confirm payment preview (user)
// @Description Konfirmasi preview → buat transaction
// @Tags Transaction
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body ConfirmTransactionRequest true "Preview reference_id"
// @Success 201 {object} utils.Response "Transaction created"
// @Router /api/transactions/confirm [post]
func ConfirmTransaction(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
			return
		}

		var req ConfirmTransactionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
			return
		}

		var preview models.TransactionPreview
		if err := db.Where("reference_id = ? AND user_id = ?", req.ReferenceID, userID).First(&preview).Error; err != nil {
			utils.ErrorResponse(c, http.StatusNotFound, "Preview not found")
			return
		}

		if preview.IsExpired() {
			db.Delete(&preview)
			utils.ErrorResponse(c, http.StatusBadRequest, "Preview has expired. Please create a new one.")
			return
		}

		expireAt := time.Now().Add(24 * time.Hour)

		transaction := models.Transaction{
			UserID:               userID.(uuid.UUID),
			ReferenceID:          generateReferenceID(),
			Status:               "pending",
			Amount:               preview.Amount,
			UniqueCode:           preview.UniqueCode,
			TotalAmount:          preview.TotalAmount,
			BankName:             preview.BankName,
			BankAccountNumber:    preview.BankAccountNumber,
			BankAccountName:      preview.BankAccountName,
			SubscriptionName:     preview.SubscriptionName,
			SubscriptionType:     preview.SubscriptionType,
			SubscriptionPrice:    preview.Amount,
			SubscriptionDuration: preview.SubscriptionDuration,
			SubscriptionBenefits: preview.SubscriptionBenefits,
			ExpiredAt:            &expireAt,
		}

		if err := db.Create(&transaction).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create transaction")
			return
		}

		db.Delete(&preview)

		transaction.PrepareResponse()
		utils.SuccessResponse(c, http.StatusCreated, "Transaction created. Please complete payment within 24 hours.", transaction)
	}
}
