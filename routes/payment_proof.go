package routes

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"mengonten-api/models"
	"mengonten-api/utils"
	"mengonten-api/worker"
)

// @Summary Upload bukti transfer (user)
// @Description Upload foto bukti transfer (max 5) + info transfer
// @Tags Payment Proof
// @Accept multipart/form-data
// @Produce json
// @Security Bearer
// @Param transaction_id path string true "Transaction ID"
// @Param account_name formData string true "Atas nama rekening"
// @Param source_bank formData string true "Transfer dari bank"
// @Param source_account_number formData string false "Nomor rekening pengirim"
// @Param photos formData []file true "Foto bukti transfer (max 5)"
// @Param descriptions formData []string true "Deskripsi per foto"
// @Success 201 {object} utils.Response "Payment proof uploaded"
// @Failure 400 {object} utils.Response "Invalid request"
// @Router /api/transactions/{transaction_id}/payment-proof [post]
func UploadPaymentProof(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
			return
		}

		transactionID := c.Param("transaction_id")
		parsedTransactionID, err := uuid.Parse(transactionID)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid transaction ID")
			return
		}

		var transaction models.Transaction
		if err := db.First(&transaction, parsedTransactionID).Error; err != nil {
			utils.ErrorResponse(c, http.StatusNotFound, "Transaction not found")
			return
		}

		if transaction.UserID != userID.(uuid.UUID) {
			utils.ErrorResponse(c, http.StatusForbidden, "Access denied")
			return
		}

		if transaction.Status != "pending" {
			utils.ErrorResponse(c, http.StatusBadRequest, "Transaction is not in pending status")
			return
		}

		var existingProof models.PaymentProof
		if err := db.Where("transaction_id = ?", parsedTransactionID).First(&existingProof).Error; err == nil {
			utils.ErrorResponse(c, http.StatusConflict, "Payment proof already submitted for this transaction")
			return
		}

		accountName := c.PostForm("account_name")
		sourceBank := c.PostForm("source_bank")
		sourceAccountNumber := c.PostForm("source_account_number")

		if accountName == "" || sourceBank == "" {
			utils.ErrorResponse(c, http.StatusBadRequest, "Account name and source bank are required")
			return
		}

		form, _ := c.MultipartForm()
		files := form.File["photos"]
		descriptions := form.Value["descriptions"]

		if len(files) == 0 {
			utils.ErrorResponse(c, http.StatusBadRequest, "At least 1 photo is required")
			return
		}

		if len(files) > 5 {
			utils.ErrorResponse(c, http.StatusBadRequest, "Maximum 5 photos allowed")
			return
		}

		if len(descriptions) != len(files) {
			utils.ErrorResponse(c, http.StatusBadRequest, "Number of descriptions must match number of photos")
			return
		}

		proof := models.PaymentProof{
			TransactionID:       parsedTransactionID,
			UserID:              userID.(uuid.UUID),
			AccountName:         accountName,
			SourceBank:          sourceBank,
			SourceAccountNumber: sourceAccountNumber,
			Status:              "pending",
		}

		if err := db.Create(&proof).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create payment proof")
			return
		}

		uploadDir := "uploads/payment-proofs"
		for i, file := range files {
			filename := fmt.Sprintf("%s_%d_%s", proof.ID.String(), i+1, filepath.Ext(file.Filename))
			filepath := filepath.Join(uploadDir, filename)

			if err := c.SaveUploadedFile(file, filepath); err != nil {
				utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to save photo")
				return
			}

			clipURL, err := worker.UploadToCloudinaryStatic(filepath, fmt.Sprintf("payment_proofs/%s", proof.ID.String()))
			if err != nil {
				clipURL = ""
			}

			photo := models.PaymentProofPhoto{
				PaymentProofID: proof.ID,
				PhotoURL:       clipURL,
				Description:    descriptions[i],
			}
			db.Create(&photo)
		}

		db.Preload("Photos").First(&proof, proof.ID)
		utils.SuccessResponse(c, http.StatusCreated, "Payment proof uploaded successfully", proof)
	}
}

// @Summary Get my payment proof (user)
// @Description Lihat bukti transfer milik user sendiri
// @Tags Payment Proof
// @Produce json
// @Security Bearer
// @Param transaction_id path string true "Transaction ID"
// @Success 200 {object} utils.Response "Payment proof"
// @Router /api/transactions/{transaction_id}/payment-proof [get]
func GetMyPaymentProof(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
			return
		}

		transactionID := c.Param("transaction_id")
		parsedTransactionID, err := uuid.Parse(transactionID)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid transaction ID")
			return
		}

		var proof models.PaymentProof
		if err := db.Preload("Photos").Where("transaction_id = ? AND user_id = ?",
			parsedTransactionID, userID).First(&proof).Error; err != nil {
			utils.ErrorResponse(c, http.StatusNotFound, "Payment proof not found")
			return
		}

		utils.SuccessResponse(c, http.StatusOK, "Payment proof retrieved", proof)
	}
}

type AdminReviewRequest struct {
	Status      string  `json:"status" binding:"required,oneof=confirmed rejected"`
	AdminNotes  string  `json:"admin_notes"`
	RefundAmount float64 `json:"refund_amount"`
}

type OverpaidRequest struct {
	AdminNotes   string  `json:"admin_notes" binding:"required"`
	RefundAmount float64 `json:"refund_amount" binding:"required,gt=0"`
	RefundProof  *string `json:"refund_proof_url"`
}

// @Summary Get all payment proofs (admin)
// @Description List semua bukti transfer - admin only
// @Tags Admin Payment Proof
// @Produce json
// @Security Bearer
// @Success 200 {object} utils.Response "Payment proofs list"
// @Router /api/admin/payment-proofs [get]
func GetAllPaymentProofs(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var proofs []models.PaymentProof
		if err := db.Preload("Photos").Order("created_at DESC").Find(&proofs).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch payment proofs")
			return
		}
		utils.SuccessResponse(c, http.StatusOK, "Payment proofs retrieved", proofs)
	}
}

// @Summary Get payment proof detail (admin)
// @Description Detail bukti transfer - admin only
// @Tags Admin Payment Proof
// @Produce json
// @Security Bearer
// @Param proof_id path string true "Payment Proof ID"
// @Success 200 {object} utils.Response "Payment proof detail"
// @Router /api/admin/payment-proofs/{proof_id} [get]
func GetPaymentProofDetail(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		proofID := c.Param("proof_id")
		parsedID, err := uuid.Parse(proofID)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid proof ID")
			return
		}

		var proof models.PaymentProof
		if err := db.Preload("Photos").First(&proof, parsedID).Error; err != nil {
			utils.ErrorResponse(c, http.StatusNotFound, "Payment proof not found")
			return
		}

		var transaction models.Transaction
		db.First(&transaction, proof.TransactionID)

		utils.SuccessResponse(c, http.StatusOK, "Payment proof retrieved", map[string]interface{}{
			"proof":       proof,
			"transaction": transaction,
		})
	}
}

// @Summary Confirm or reject payment proof (admin)
// @Description Konfirmasi atau tolak bukti transfer
// @Tags Admin Payment Proof
// @Accept json
// @Produce json
// @Security Bearer
// @Param proof_id path string true "Payment Proof ID"
// @Param request body AdminReviewRequest true "Review result"
// @Success 200 {object} utils.Response "Review result"
// @Router /api/admin/payment-proofs/{proof_id}/review [put]
func ReviewPaymentProof(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		proofID := c.Param("proof_id")
		parsedID, err := uuid.Parse(proofID)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid proof ID")
			return
		}

		var req AdminReviewRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
			return
		}

		var proof models.PaymentProof
		if err := db.Preload("Photos").First(&proof, parsedID).Error; err != nil {
			utils.ErrorResponse(c, http.StatusNotFound, "Payment proof not found")
			return
		}

		if proof.Status != "pending" {
			utils.ErrorResponse(c, http.StatusBadRequest, "Payment proof has already been reviewed")
			return
		}

		db.Model(&proof).Updates(map[string]interface{}{
			"status":      req.Status,
			"admin_notes": req.AdminNotes,
		})

		if req.Status == "confirmed" {
			db.Model(&models.Transaction{}).Where("id = ?", proof.TransactionID).
				Update("status", "paid")

			var transaction models.Transaction
			db.First(&transaction, proof.TransactionID)
			createUserSubscription(db, &transaction)
		} else if req.Status == "rejected" {
			db.Model(&models.Transaction{}).Where("id = ?", proof.TransactionID).
				Update("status", "cancelled")
		}

		db.Preload("Photos").First(&proof, parsedID)
		utils.SuccessResponse(c, http.StatusOK, "Payment proof reviewed", proof)
	}
}

// @Summary Confirm overpaid payment proof (admin)
// @Description Konfirmasi bukti transfer dengan nominal kelebihan + upload bukti refund
// @Tags Admin Payment Proof
// @Accept multipart/form-data
// @Produce json
// @Security Bearer
// @Param proof_id path string true "Payment Proof ID"
// @Param admin_notes formData string true "Catatan admin"
// @Param refund_amount formData number true "Jumlah refund"
// @Param refund_proof formData file false "Bukti pengembalian dana"
// @Success 200 {object} utils.Response "Overpaid confirmed"
// @Router /api/admin/payment-proofs/{proof_id}/confirm-overpaid [put]
func ConfirmOverpaidProof(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		proofID := c.Param("proof_id")
		parsedID, err := uuid.Parse(proofID)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid proof ID")
			return
		}

		var proof models.PaymentProof
		if err := db.Preload("Photos").First(&proof, parsedID).Error; err != nil {
			utils.ErrorResponse(c, http.StatusNotFound, "Payment proof not found")
			return
		}

		if proof.Status != "pending" {
			utils.ErrorResponse(c, http.StatusBadRequest, "Payment proof has already been reviewed")
			return
		}

		adminNotes := c.PostForm("admin_notes")
		refundAmountStr := c.PostForm("refund_amount")

		if adminNotes == "" || refundAmountStr == "" {
			utils.ErrorResponse(c, http.StatusBadRequest, "Admin notes and refund amount are required")
			return
		}

		var refundAmount float64
		fmt.Sscanf(refundAmountStr, "%f", &refundAmount)

		var refundProofURL string
		file, _ := c.FormFile("refund_proof")
		if file != nil {
			uploadDir := "uploads/refund-proofs"
			filename := fmt.Sprintf("refund_%s_%s", proof.ID.String(), filepath.Ext(file.Filename))
			filePath := filepath.Join(uploadDir, filename)

			if err := c.SaveUploadedFile(file, filePath); err == nil {
				url, _ := worker.UploadToCloudinaryStatic(filePath, fmt.Sprintf("refund_proofs/%s", proof.ID.String()))
				refundProofURL = url
			}
		}

		db.Model(&proof).Updates(map[string]interface{}{
			"status":          "confirmed",
			"admin_notes":     adminNotes,
			"refund_proof_url": refundProofURL,
			"refund_amount":   refundAmount,
		})

		db.Model(&models.Transaction{}).Where("id = ?", proof.TransactionID).
			Update("status", "paid")

		var transaction models.Transaction
		db.First(&transaction, proof.TransactionID)
		createUserSubscription(db, &transaction)

		db.Preload("Photos").First(&proof, parsedID)
		utils.SuccessResponse(c, http.StatusOK, "Overpaid proof confirmed", proof)
	}
}
