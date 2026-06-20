package routes

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"mengonten-api/config"
	"mengonten-api/models"
	"mengonten-api/utils"
	"mengonten-api/worker"
)

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"required,min=3"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  UserResponse `json:"user"`
}

type UserResponse struct {
	ID             uuid.UUID `json:"id"`
	Email          string    `json:"email"`
	Username       string    `json:"username"`
	Role           string    `json:"role"`
	AccountStatus  string    `json:"account_status"`
	WarningMessage string    `json:"warning_message,omitempty"`
	IsVerified     bool      `json:"is_verified"`
}

// @Summary Register new user
// @Description Create new user account with email, username, and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration data"
// @Success 201 {object} utils.Response{data=AuthResponse} "User registered successfully"
// @Failure 400 {object} utils.Response "Invalid request format"
// @Failure 409 {object} utils.Response "Email or username already exists"
// @Failure 500 {object} utils.Response "Server error"
// @Router /api/auth/register [post]
func Register(db *gorm.DB, emailSender *worker.EmailSender) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
			return
		}

		var existingUser models.User
		if err := db.Where("email = ? OR username = ?", req.Email, req.Username).First(&existingUser).Error; err == nil {
			utils.ErrorResponse(c, http.StatusConflict, "Email or username already exists")
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to hash password")
			return
		}

		verificationToken := uuid.New().String()
		tokenExpires := time.Now().Add(24 * time.Hour)

		user := models.User{
			Email:             req.Email,
			Username:          req.Username,
			Password:          string(hashedPassword),
			VerificationToken: verificationToken,
			TokenExpiresAt:    &tokenExpires,
		}

		if err := db.Create(&user).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create user")
			return
		}

		if emailSender != nil {
			go emailSender.SendVerificationEmail(user.Email, user.Username, verificationToken)
		}

		utils.SuccessResponse(c, http.StatusCreated, "User registered. Please check your email for verification link.", nil)
	}
}

// @Summary User login
// @Description Authenticate user with email and password, returns JWT token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} utils.Response{data=AuthResponse} "Login successful"
// @Failure 400 {object} utils.Response "Invalid request format"
// @Failure 401 {object} utils.Response "Invalid email or password"
// @Failure 500 {object} utils.Response "Server error"
// @Router /api/auth/login [post]
func Login(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
			return
		}

		var user models.User
		if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid email or password")
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid email or password")
			return
		}

		if !user.IsVerified {
			utils.ErrorResponse(c, http.StatusForbidden, "Email not verified. Please check your inbox.")
			return
		}

		token := generateToken(user.ID)
		utils.SuccessResponse(c, http.StatusOK, "Login successful", AuthResponse{
			Token: token,
			User: UserResponse{
				ID:             user.ID,
				Email:          user.Email,
				Username:       user.Username,
				Role:           user.Role,
				AccountStatus:  user.AccountStatus,
				WarningMessage: user.WarningMessage,
				IsVerified:     user.IsVerified,
			},
		})
	}
}

// @Summary User logout
// @Description Logout user (clears session)
// @Tags Auth
// @Success 200 {object} utils.Response "Logout successful"
// @Router /api/auth/logout [post]
func Logout(c *gin.Context) {
	utils.SuccessResponse(c, http.StatusOK, "Logout successful", nil)
}

// @Summary Get user profile
// @Description Retrieve authenticated user profile information
// @Tags Auth
// @Security Bearer
// @Produce json
// @Success 200 {object} utils.Response{data=UserResponse} "Profile retrieved"
// @Failure 401 {object} utils.Response "User not authenticated"
// @Failure 404 {object} utils.Response "User not found"
// @Router /api/profile [get]
func GetProfile(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
			return
		}

		var user models.User
		if err := db.First(&user, userID).Error; err != nil {
			utils.ErrorResponse(c, http.StatusNotFound, "User not found")
			return
		}

		utils.SuccessResponse(c, http.StatusOK, "Profile retrieved", UserResponse{
			ID:             user.ID,
			Email:          user.Email,
			Username:       user.Username,
			Role:           user.Role,
			AccountStatus:  user.AccountStatus,
			WarningMessage: user.WarningMessage,
			IsVerified:     user.IsVerified,
		})
	}
}

func generateToken(userID uuid.UUID) string {
	jwtConfig := config.GetJWTConfig()
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     time.Now().Add(jwtConfig.Expiration).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(jwtConfig.Secret))
	return tokenString
}

type VerifyEmailRequest struct {
	Token string `json:"token" binding:"required"`
}

// @Summary Verify email address
// @Description Verify user email dengan token dari email
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body VerifyEmailRequest true "Verification token"
// @Success 200 {object} utils.Response "Email verified successfully"
// @Failure 400 {object} utils.Response "Invalid or expired token"
// @Router /api/auth/verify-email [post]
func VerifyEmail(db *gorm.DB, emailSender *worker.EmailSender) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req VerifyEmailRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
			return
		}

		var user models.User
		if err := db.Where("verification_token = ?", req.Token).First(&user).Error; err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid verification token")
			return
		}

		if user.IsVerified {
			utils.SuccessResponse(c, http.StatusOK, "Email already verified", nil)
			return
		}

		if user.TokenExpiresAt != nil && time.Now().After(*user.TokenExpiresAt) {
			utils.ErrorResponse(c, http.StatusBadRequest, "Verification token has expired. Please request a new one.")
			return
		}

		now := time.Now()
		db.Model(&user).Updates(map[string]interface{}{
			"is_verified":        true,
			"verification_token": "",
			"token_expires_at":   nil,
			"verified_at":        &now,
		})

		utils.SuccessResponse(c, http.StatusOK, "Email verified successfully. You can now login.", nil)
	}
}

type ResendVerificationRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// @Summary Resend verification email
// @Description Kirim ulang email verifikasi (cooldown 5 menit)
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body ResendVerificationRequest true "Email address"
// @Success 200 {object} utils.Response "Verification email sent"
// @Failure 400 {object} utils.Response "Invalid request"
// @Failure 429 {object} utils.Response "Rate limited - cooldown active"
// @Router /api/auth/resend-verification [post]
func ResendVerification(db *gorm.DB, emailSender *worker.EmailSender) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ResendVerificationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
			return
		}

		var user models.User
		if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
			utils.SuccessResponse(c, http.StatusOK, "If this email is registered, a verification link has been sent.", nil)
			return
		}

		if user.IsVerified {
			utils.ErrorResponse(c, http.StatusBadRequest, "Email is already verified")
			return
		}

		if user.TokenExpiresAt != nil && time.Now().Before(*user.TokenExpiresAt) {
			remaining := time.Until(*user.TokenExpiresAt)
			if remaining > (23*time.Hour + 55*time.Minute) {
				utils.ErrorResponse(c, http.StatusTooManyRequests,
					"Please wait before requesting a new verification email. Try again in 5 minutes.")
				return
			}
		}

		newToken := uuid.New().String()
		newExpiry := time.Now().Add(24 * time.Hour)

		db.Model(&user).Updates(map[string]interface{}{
			"verification_token": newToken,
			"token_expires_at":   newExpiry,
		})

		if emailSender != nil {
			go emailSender.SendVerificationEmail(user.Email, user.Username, newToken)
		}

		utils.SuccessResponse(c, http.StatusOK, "Verification email sent. Please check your inbox.", nil)
	}
}
