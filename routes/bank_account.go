package routes

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"mengonten-api/config"
	"mengonten-api/models"
	"mengonten-api/utils"
	"mengonten-api/worker"
)

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
// @Description Tambah rekening bank baru + upload icon - admin only
// @Tags Bank Account
// @Accept multipart/form-data
// @Produce json
// @Security Bearer
// @Param bank_name formData string true "Nama bank"
// @Param account_number formData string true "Nomor rekening"
// @Param account_name formData string true "Atas nama"
// @Param icon formData file false "Icon bank"
// @Success 201 {object} utils.Response "Bank account created"
// @Router /api/admin/bank-accounts [post]
func CreateBankAccount(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		bankName := c.PostForm("bank_name")
		accountNumber := c.PostForm("account_number")
		accountName := c.PostForm("account_name")

		if bankName == "" || accountNumber == "" || accountName == "" {
			utils.ErrorResponse(c, http.StatusBadRequest, "bank_name, account_number, and account_name are required")
			return
		}

		var iconURL string
		file, _ := c.FormFile("icon")
		if file != nil {
			uploadDir := "uploads/bank-icons"
			os.MkdirAll(uploadDir, os.ModePerm)
			ext := filepath.Ext(file.Filename)
			fileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
			filePath := filepath.Join(uploadDir, fileName)

			if err := c.SaveUploadedFile(file, filePath); err != nil {
				utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to save icon")
				return
			}

			url, err := worker.UploadToR2Static(filePath, fmt.Sprintf("bank-icons/%s", fileName))
			if err != nil {
				log.Printf("Failed to upload bank icon to R2: %v", err)
				os.Remove(filePath)
			} else {
				os.Remove(filePath)
				iconURL = url
			}
		}

		account := models.BankAccount{
			BankName:      bankName,
			AccountNumber: accountNumber,
			AccountName:   accountName,
			Icon:          iconURL,
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
// @Description Update rekening bank + ganti icon - admin only
// @Tags Bank Account
// @Accept multipart/form-data
// @Produce json
// @Security Bearer
// @Param account_id path string true "Account ID"
// @Param bank_name formData string false "Nama bank"
// @Param account_number formData string false "Nomor rekening"
// @Param account_name formData string false "Atas nama"
// @Param icon formData file false "Icon bank (ganti)"
// @Param is_active formData string false "Aktif/nonaktif (true/false)"
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

		updates := map[string]interface{}{}

		if bankName := c.PostForm("bank_name"); bankName != "" {
			updates["bank_name"] = bankName
		}
		if accountNumber := c.PostForm("account_number"); accountNumber != "" {
			updates["account_number"] = accountNumber
		}
		if accountName := c.PostForm("account_name"); accountName != "" {
			updates["account_name"] = accountName
		}
		if isActiveStr := c.PostForm("is_active"); isActiveStr != "" {
			if isActiveStr == "true" {
				updates["is_active"] = true
			} else if isActiveStr == "false" {
				updates["is_active"] = false
			}
		}

		file, _ := c.FormFile("icon")
		if file != nil {
			uploadDir := "uploads/bank-icons"
			os.MkdirAll(uploadDir, os.ModePerm)
			ext := filepath.Ext(file.Filename)
			fileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
			filePath := filepath.Join(uploadDir, fileName)

			if err := c.SaveUploadedFile(file, filePath); err != nil {
				utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to save icon")
				return
			}

			url, err := worker.UploadToR2Static(filePath, fmt.Sprintf("bank-icons/%s", fileName))
			if err != nil {
				log.Printf("Failed to upload bank icon to R2: %v", err)
				os.Remove(filePath)
			} else {
				os.Remove(filePath)
				if account.Icon != "" {
					deleteIconFromR2(account.Icon)
				}
				updates["icon"] = url
			}
		}

		if len(updates) > 0 {
			db.Model(&account).Updates(updates)
		}

		db.First(&account, parsedID)
		utils.SuccessResponse(c, http.StatusOK, "Bank account updated", account)
	}
}

// @Summary Delete bank account (admin)
// @Description Hapus rekening bank + icon - admin only
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

		if account.Icon != "" {
			deleteIconFromR2(account.Icon)
		}

		if err := db.Delete(&account).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete bank account")
			return
		}

		utils.SuccessResponse(c, http.StatusOK, "Bank account deleted successfully", nil)
	}
}

func deleteIconFromR2(iconURL string) {
	publicURL := config.InitExternalAPIs().R2PublicURL
	publicURL = strings.TrimRight(publicURL, "/")
	key := strings.TrimPrefix(iconURL, publicURL+"/")
	if key != iconURL && key != "" {
		if err := worker.DeleteFromR2(key); err != nil {
			log.Printf("Failed to delete old icon from R2: %v", err)
		}
	}
}
