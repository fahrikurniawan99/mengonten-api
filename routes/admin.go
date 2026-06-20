package routes

import (
	"net/http"

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
	ID         uuid.UUID `json:"id"`
	Email      string    `json:"email"`
	Username   string    `json:"username"`
	Role       string    `json:"role"`
	IsVerified bool      `json:"is_verified"`
	CreatedAt  string    `json:"created_at"`
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
			response[i] = AdminUserResponse{
				ID:         u.ID,
				Email:      u.Email,
				Username:   u.Username,
				Role:       u.Role,
				IsVerified: u.IsVerified,
				CreatedAt:  u.CreatedAt.Format("2006-01-02 15:04:05"),
			}
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

		utils.SuccessResponse(c, http.StatusOK, "User retrieved", AdminUserResponse{
			ID:         user.ID,
			Email:      user.Email,
			Username:   user.Username,
			Role:       user.Role,
			IsVerified: user.IsVerified,
			CreatedAt:  user.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
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

		utils.SuccessResponse(c, http.StatusOK, "User role updated", AdminUserResponse{
			ID:         user.ID,
			Email:      user.Email,
			Username:   user.Username,
			Role:       req.Role,
			IsVerified: user.IsVerified,
			CreatedAt:  user.CreatedAt.Format("2006-01-02 15:04:05"),
		})
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
