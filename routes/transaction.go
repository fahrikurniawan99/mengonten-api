package routes

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"mengonten-api/models"
	"mengonten-api/utils"
	"mengonten-api/worker"
)

type CreateTransactionRequest struct {
	SubscriptionPlanID uuid.UUID `json:"subscription_plan_id" binding:"required"`
	PaymentMethod      string    `json:"payment_method" binding:"required"`
}

func generateTransactionReferenceID() string {
	return fmt.Sprintf("TXN-%d-%s", time.Now().UnixNano(), uuid.New().String()[:8])
}

// @Summary Create transaction (user)
// @Description Buat transaksi baru + redirect ke Duitku payment page
// @Tags Transaction
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body CreateTransactionRequest true "Plan ID"
// @Success 201 {object} utils.Response "Transaction created"
// @Router /api/transactions [post]
func CreateTransaction(db *gorm.DB, pakasirClient *worker.PakasirClient, emailSender *worker.EmailSender) gin.HandlerFunc {
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

		if !worker.IsValidPakasirMethod(req.PaymentMethod) {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid payment method")
			return
		}

		var plan models.SubscriptionPlan
		if err := db.First(&plan, req.SubscriptionPlanID).Error; err != nil {
			utils.ErrorResponse(c, http.StatusNotFound, "Subscription plan not found")
			return
		}

		if !plan.IsActive {
			utils.ErrorResponse(c, http.StatusBadRequest, "Plan is not active")
			return
		}

		if plan.FinalPrice > 0 {
			var activeOrder models.Order
			if err := db.Where("user_id = ? AND status = ? AND expired_at > ?",
				userID, "active", time.Now()).
				Order("expired_at DESC").
				First(&activeOrder).Error; err == nil {
				utils.ErrorResponse(c, http.StatusBadRequest, "Anda masih memiliki langganan aktif. Selesaikan langganan saat ini sebelum berlangganan baru.")
				return
			}
		}

		referenceID := generateTransactionReferenceID()
		amount := int64(plan.FinalPrice)

		transaction := models.Transaction{
			UserID:             userID.(uuid.UUID),
			SubscriptionPlanID: plan.ID,
			ReferenceID:        referenceID,
			ProductName:        fmt.Sprintf("%s (%d hari)", plan.Name, plan.DurationDays),
			PaymentTotal:       plan.FinalPrice,
			PaymentMethod:      req.PaymentMethod,
			Status:             "pending",
		}

		if err := db.Create(&transaction).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create transaction")
			return
		}

		if pakasirClient != nil && amount > 0 {
			result, err := pakasirClient.CreateTransaction(referenceID, amount, req.PaymentMethod)
			if err != nil {
				db.Delete(&transaction)
				log.Printf("[Transaction] Pakasir create failed: %v", err)
				utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create payment: "+err.Error())
				return
			}

			log.Printf("[Transaction] Pakasir success: ref=%s method=%s number=%s",
				referenceID, result.PaymentMethod, result.PaymentNumber)

			var expiredAtTime *time.Time
			if result.ExpiredAt != "" {
				t, err := time.Parse(time.RFC3339Nano, result.ExpiredAt)
				if err == nil {
					expiredAtTime = &t
				}
			}

			db.Model(&transaction).Updates(map[string]interface{}{
				"payment_number": result.PaymentNumber,
				"payment_method": result.PaymentMethod,
				"expired_at":     expiredAtTime,
			})
			transaction.PaymentNumber = result.PaymentNumber
			transaction.PaymentMethod = result.PaymentMethod
			transaction.ExpiredAt = expiredAtTime

			if emailSender != nil {
				var user models.User
				if err := db.First(&user, userID).Error; err == nil {
					expiredAtDisplay := "24 jam"
					if expiredAtTime != nil {
						expiredAtDisplay = expiredAtTime.Format("02 Jan 2006, 15:04 WIB")
					}
					go emailSender.SendTransactionEmail(
						user.Email, "",
						transaction.ReferenceID, plan.Name,
						transaction.PaymentTotal, transaction.PaymentMethod,
						transaction.PaymentNumber, expiredAtDisplay,
					)
				}
			}
		}

		db.First(&transaction, transaction.ID)
		utils.SuccessResponse(c, http.StatusCreated, "Transaction created", transaction)
	}
}

// @Summary Pakasir payment callback
// @Description Webhook dari Pakasir untuk update status transaksi
// @Tags Transaction
// @Accept json
// @Produce json
// @Param request body object false "Callback params"
// @Success 200 {object} utils.Response "Callback processed"
// @Router /callback/pakasir [post]
func CallbackPakasir(db *gorm.DB, pakasirClient *worker.PakasirClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var params worker.WebhookParams
		if err := c.ShouldBindJSON(&params); err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid callback data")
			return
		}

		if pakasirClient != nil {
			if !pakasirClient.VerifyWebhook(&params) {
				utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid webhook")
				return
			}
		}

		var transaction models.Transaction
		if err := db.Where("reference_id = ?", params.OrderID).First(&transaction).Error; err != nil {
			log.Printf("[Callback] Transaction not found: order=%s", params.OrderID)
			utils.ErrorResponse(c, http.StatusNotFound, "Transaction not found")
			return
		}

		log.Printf("[Callback] Pakasir webhook: order=%s status=%s amount=%.0f method=%s",
			params.OrderID, params.Status, params.Amount, params.PaymentMethod)

		if transaction.Status != "pending" {
			c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "already processed"})
			return
		}

		updates := map[string]interface{}{}

		if params.Status == "completed" || params.Status == "success" {
			now := time.Now()
			updates["status"] = "success"
			updates["payment_at"] = &now

			db.Model(&transaction).Updates(updates)
			db.First(&transaction, transaction.ID)

			var plan models.SubscriptionPlan
			if err := db.First(&plan, transaction.SubscriptionPlanID).Error; err != nil {
				c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "transaction updated but plan not found"})
				return
			}

			var activeOrder models.Order
			if err := db.Where("user_id = ? AND status = ? AND expired_at > ?",
				transaction.UserID, "active", time.Now()).
				Order("expired_at DESC").
				First(&activeOrder).Error; err == nil {
				if activeOrder.ProductPrice > 0 {
					db.Model(&activeOrder).Update("status", "cancelled")
				}
			}

			order := models.Order{
				UserID:        transaction.UserID,
				TransactionID: transaction.ID,
				PlanID:        plan.ID,
				ProductName:   plan.Name,
				ProductPrice:  plan.FinalPrice,
				Status:        "active",
				ExpiredAt:     time.Now().AddDate(0, 0, plan.DurationDays),
			}

			if err := db.Create(&order).Error; err != nil {
				c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "transaction updated but order creation failed"})
				return
			}

			var rules []models.SubscriptionRule
			db.Where("plan_id = ?", plan.ID).Find(&rules)
			for _, rule := range rules {
				ruleID := rule.ID
				orderRule := models.OrderRule{
					OrderID:            order.ID,
					SubscriptionRuleID: &ruleID,
					RuleKey:            rule.RuleKey,
					RuleValue:          rule.RuleValue,
				}
				db.Create(&orderRule)
			}

			db.Model(&transaction).Update("order_id", order.ID)

		} else {
			updates["status"] = "failed"
			db.Model(&transaction).Updates(updates)
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "Callback processed"})
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

		utils.SuccessResponse(c, http.StatusOK, "Transactions retrieved", transactions)
	}
}

// @Summary Get transaction detail (user)
// @Description Lihat detail transaksi milik user sendiri
// @Tags Transaction
// @Produce json
// @Security Bearer
// @Param transaction_id path string true "Transaction ID"
// @Success 200 {object} utils.Response "Transaction detail"
// @Router /api/transactions/{transaction_id} [get]
func GetTransactionDetail(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
			return
		}

		transactionID := c.Param("transaction_id")
		parsedID, err := uuid.Parse(transactionID)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid transaction ID")
			return
		}

		var transaction models.Transaction
		if err := db.Where("id = ? AND user_id = ?", parsedID, userID).First(&transaction).Error; err != nil {
			utils.ErrorResponse(c, http.StatusNotFound, "Transaction not found")
			return
		}

		data := map[string]interface{}{
			"id":                  transaction.ID,
			"user_id":             transaction.UserID,
			"subscription_plan_id": transaction.SubscriptionPlanID,
			"reference_id":        transaction.ReferenceID,
			"product_name":        transaction.ProductName,
			"payment_total":       transaction.PaymentTotal,
			"status":              transaction.Status,
			"payment_method":      transaction.PaymentMethod,
			"payment_number":      transaction.PaymentNumber,
			"expired_at":          transaction.ExpiredAt,
			"order_id":            transaction.OrderID,
			"payment_at":          transaction.PaymentAt,
			"created_at":          transaction.CreatedAt,
			"updated_at":          transaction.UpdatedAt,
		}

		if transaction.OrderID != nil {
			var order models.Order
			if err := db.Preload("Rules").First(&order, *transaction.OrderID).Error; err == nil {
				data["order"] = order
			}
		}

		utils.SuccessResponse(c, http.StatusOK, "Transaction retrieved", data)
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

		utils.SuccessResponse(c, http.StatusOK, "Transactions retrieved", transactions)
	}
}

type UpdateTransactionStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=success failed cancel"`
}

// @Summary Update transaction status (admin)
// @Description Update status transaksi - admin only
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

		if req.Status == "success" {
			now := time.Now()
			updates["payment_at"] = &now

			db.Model(&transaction).Updates(updates)
			db.First(&transaction, parsedID)

			var plan models.SubscriptionPlan
			if err := db.First(&plan, transaction.SubscriptionPlanID).Error; err != nil {
				utils.ErrorResponse(c, http.StatusOK, "Transaction updated but plan not found")
				return
			}

			var activeOrder models.Order
			if err := db.Where("user_id = ? AND status = ? AND expired_at > ?",
				transaction.UserID, "active", time.Now()).
				Order("expired_at DESC").
				First(&activeOrder).Error; err == nil {
				if activeOrder.ProductPrice > 0 {
					db.Model(&activeOrder).Update("status", "cancelled")
				}
			}

			order := models.Order{
				UserID:        transaction.UserID,
				TransactionID: transaction.ID,
				PlanID:        plan.ID,
				ProductName:   plan.Name,
				ProductPrice:  plan.FinalPrice,
				Status:        "active",
				ExpiredAt:     time.Now().AddDate(0, 0, plan.DurationDays),
			}

			if err := db.Create(&order).Error; err != nil {
				utils.ErrorResponse(c, http.StatusOK, "Transaction updated but order creation failed")
				return
			}

			var rules []models.SubscriptionRule
			db.Where("plan_id = ?", plan.ID).Find(&rules)
			for _, rule := range rules {
				ruleID := rule.ID
				orderRule := models.OrderRule{
					OrderID:            order.ID,
					SubscriptionRuleID: &ruleID,
					RuleKey:            rule.RuleKey,
					RuleValue:          rule.RuleValue,
				}
				db.Create(&orderRule)
			}

			db.Model(&transaction).Update("order_id", order.ID)
			db.First(&transaction, parsedID)
		} else {
			db.Model(&transaction).Updates(updates)
			db.First(&transaction, parsedID)
		}

		utils.SuccessResponse(c, http.StatusOK, "Transaction updated", transaction)
	}
}
