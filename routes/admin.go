package routes

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"mengonten-api/models"
	"mengonten-api/utils"
)

type UpdateRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=admin user"`
}

type AdminUserResponse struct {
	ID               uuid.UUID `json:"id"`
	Email            string    `json:"email"`
	Role             string    `json:"role"`
	AccountStatus    string    `json:"account_status"`
	SuspendReason    string    `json:"suspend_reason"`
	SuspendExpiresAt *string   `json:"suspend_expires_at"`
	WarningMessage   string    `json:"warning_message"`
	WarningCount     int       `json:"warning_count"`
	IsVerified       bool      `json:"is_verified"`
	CreatedAt        string    `json:"created_at"`
}

func toAdminUserResponse(u models.User) AdminUserResponse {
	var suspendExpiry *string
	if u.SuspendExpiresAt != nil {
		s := u.SuspendExpiresAt.Format("2006-01-02 15:04:05")
		suspendExpiry = &s
	}
	return AdminUserResponse{
		ID:               u.ID,
		Email:            u.Email,
		Role:             u.Role,
		AccountStatus:    u.AccountStatus,
		SuspendReason:    u.SuspendReason,
		SuspendExpiresAt: suspendExpiry,
		WarningMessage:   u.WarningMessage,
		WarningCount:     u.WarningCount,
		IsVerified:       u.IsVerified,
		CreatedAt:        u.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// @Summary Get all users (admin)
// @Description List semua users - admin only
// @Tags Admin
// @Produce json
// @Security Bearer
// @Success 200 {object} utils.Response "Users list"
// @Failure 403 {object} utils.Response "Access denied"
// @Router /api/admin/users [get]
func GetUsers(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var users []models.User
		if err := db.Find(&users).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch users")
			return
		}

		response := make([]AdminUserResponse, len(users))
		for i, u := range users {
			response[i] = toAdminUserResponse(u)
		}

		utils.SuccessResponse(c, http.StatusOK, "Users retrieved", response)
	}
}

// @Summary Get user by ID (admin)
// @Description Get detail user - admin only
// @Tags Admin
// @Produce json
// @Security Bearer
// @Param user_id path string true "User ID"
// @Success 200 {object} utils.Response "User detail"
// @Failure 404 {object} utils.Response "User not found"
// @Router /api/admin/users/{user_id} [get]
func GetUserByID(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")
		parsedUserID, err := uuid.Parse(userID)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
			return
		}

		var user models.User
		if err := db.First(&user, parsedUserID).Error; err != nil {
			utils.ErrorResponse(c, http.StatusNotFound, "User not found")
			return
		}

		utils.SuccessResponse(c, http.StatusOK, "User retrieved", toAdminUserResponse(user))
	}
}

type SuspendRequest struct {
	Reason    string  `json:"reason" binding:"required"`
	ExpiresAt *string `json:"expires_at"`
}

type WarnRequest struct {
	Message string `json:"message" binding:"required"`
}

// @Summary Update user role (admin)
// @Description Ubah role user - admin only
// @Tags Admin
// @Accept json
// @Produce json
// @Security Bearer
// @Param user_id path string true "User ID"
// @Param request body UpdateRoleRequest true "New role"
// @Success 200 {object} utils.Response "Role updated"
// @Failure 400 {object} utils.Response "Invalid request"
// @Router /api/admin/users/{user_id}/role [put]
func UpdateUserRole(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")
		parsedUserID, err := uuid.Parse(userID)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
			return
		}

		var req UpdateRoleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
			return
		}

		var user models.User
		if err := db.First(&user, parsedUserID).Error; err != nil {
			utils.ErrorResponse(c, http.StatusNotFound, "User not found")
			return
		}

		db.Model(&user).Update("role", req.Role)

		utils.SuccessResponse(c, http.StatusOK, "User role updated", toAdminUserResponse(user))
	}
}

// @Summary Delete user (admin)
// @Description Hapus user - admin only
// @Tags Admin
// @Produce json
// @Security Bearer
// @Param user_id path string true "User ID"
// @Success 200 {object} utils.Response "User deleted"
// @Failure 404 {object} utils.Response "User not found"
// @Router /api/admin/users/{user_id} [delete]
func DeleteUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")
		parsedUserID, err := uuid.Parse(userID)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
			return
		}

		var user models.User
		if err := db.First(&user, parsedUserID).Error; err != nil {
			utils.ErrorResponse(c, http.StatusNotFound, "User not found")
			return
		}

		if user.Role == "admin" {
			utils.ErrorResponse(c, http.StatusForbidden, "Cannot delete admin user")
			return
		}

		if err := db.Delete(&user).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete user")
			return
		}

		utils.SuccessResponse(c, http.StatusOK, "User deleted successfully", nil)
	}
}

// @Summary Suspend user (admin)
// @Description Tangguhkan akun user - admin only
// @Tags Admin
// @Accept json
// @Produce json
// @Security Bearer
// @Param user_id path string true "User ID"
// @Param request body SuspendRequest true "Suspend data"
// @Success 200 {object} utils.Response "User suspended"
// @Router /api/admin/users/{user_id}/suspend [put]
func SuspendUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")
		parsedUserID, err := uuid.Parse(userID)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
			return
		}

		var user models.User
		if err := db.First(&user, parsedUserID).Error; err != nil {
			utils.ErrorResponse(c, http.StatusNotFound, "User not found")
			return
		}

		if user.Role == "admin" {
			utils.ErrorResponse(c, http.StatusForbidden, "Cannot suspend admin user")
			return
		}

		var req SuspendRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
			return
		}

		updates := map[string]interface{}{
			"account_status": "suspended",
			"suspend_reason": req.Reason,
		}

		if req.ExpiresAt != nil {
			parsedTime, err := time.Parse("2006-01-02T15:04:05Z", *req.ExpiresAt)
			if err != nil {
				utils.ErrorResponse(c, http.StatusBadRequest, "Invalid expires_at format. Use 2006-01-02T15:04:05Z")
				return
			}
			updates["suspend_expires_at"] = &parsedTime
		} else {
			utils.ErrorResponse(c, http.StatusBadRequest, "expires_at is required for suspend")
			return
		}

		db.Model(&user).Updates(updates)
		db.First(&user, parsedUserID)

		utils.SuccessResponse(c, http.StatusOK, "User suspended", toAdminUserResponse(user))
	}
}

// @Summary Warn user (admin)
// @Description Kirim peringatan ke user - admin only
// @Tags Admin
// @Accept json
// @Produce json
// @Security Bearer
// @Param user_id path string true "User ID"
// @Param request body WarnRequest true "Warning message"
// @Success 200 {object} utils.Response "Warning sent"
// @Router /api/admin/users/{user_id}/warn [put]
func WarnUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")
		parsedUserID, err := uuid.Parse(userID)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
			return
		}

		var user models.User
		if err := db.First(&user, parsedUserID).Error; err != nil {
			utils.ErrorResponse(c, http.StatusNotFound, "User not found")
			return
		}

		var req WarnRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
			return
		}

		db.Model(&user).Updates(map[string]interface{}{
			"account_status":  "warning",
			"warning_message": req.Message,
			"warning_count":   user.WarningCount + 1,
		})

		db.First(&user, parsedUserID)

		utils.SuccessResponse(c, http.StatusOK, "Warning sent to user", toAdminUserResponse(user))
	}
}

// @Summary Deactivate user (admin)
// @Description Nonaktifkan akun user - admin only
// @Tags Admin
// @Produce json
// @Security Bearer
// @Param user_id path string true "User ID"
// @Success 200 {object} utils.Response "User deactivated"
// @Router /api/admin/users/{user_id}/deactivate [put]
func DeactivateUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")
		parsedUserID, err := uuid.Parse(userID)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
			return
		}

		var user models.User
		if err := db.First(&user, parsedUserID).Error; err != nil {
			utils.ErrorResponse(c, http.StatusNotFound, "User not found")
			return
		}

		if user.Role == "admin" {
			utils.ErrorResponse(c, http.StatusForbidden, "Cannot deactivate admin user")
			return
		}

		db.Model(&user).Updates(map[string]interface{}{
			"account_status":     "deactivated",
			"suspend_reason":     "",
			"suspend_expires_at": nil,
		})

		db.First(&user, parsedUserID)

		utils.SuccessResponse(c, http.StatusOK, "User deactivated", toAdminUserResponse(user))
	}
}

// @Summary Reactivate user (admin)
// @Description Aktifkan kembali akun user - admin only
// @Tags Admin
// @Produce json
// @Security Bearer
// @Param user_id path string true "User ID"
// @Success 200 {object} utils.Response "User reactivated"
// @Router /api/admin/users/{user_id}/reactivate [put]
func ReactivateUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")
		parsedUserID, err := uuid.Parse(userID)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
			return
		}

		var user models.User
		if err := db.First(&user, parsedUserID).Error; err != nil {
			utils.ErrorResponse(c, http.StatusNotFound, "User not found")
			return
		}

		db.Model(&user).Updates(map[string]interface{}{
			"account_status":     "active",
			"suspend_reason":     "",
			"suspend_expires_at": nil,
			"warning_message":    "",
		})

		db.First(&user, parsedUserID)

		utils.SuccessResponse(c, http.StatusOK, "User reactivated", toAdminUserResponse(user))
	}
}
