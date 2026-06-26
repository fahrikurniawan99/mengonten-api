package routes

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"mengonten-api/config"
	"mengonten-api/models"
	"mengonten-api/utils"
	"mengonten-api/worker"
)

type RegisterRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type LoginRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type VerifyOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp" binding:"required"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

type UserResponse struct {
	ID             uuid.UUID `json:"id"`
	Email          string    `json:"email"`
	Role           string    `json:"role"`
	AccountStatus  string    `json:"account_status"`
	WarningMessage string    `json:"warning_message,omitempty"`
	IsVerified     bool      `json:"is_verified"`
}

// @Summary Register new user
// @Description Daftar dengan email, kirim verifikasi email
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Email"
// @Success 201 {object} utils.Response "Registration email sent"
// @Router /api/auth/register [post]
func Register(db *gorm.DB, emailSender *worker.EmailSender) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
			return
		}

		var existingUser models.User
		if err := db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
			utils.ErrorResponse(c, http.StatusConflict, "Email already exists")
			return
		}

		verificationToken := uuid.New().String()
		tokenExpires := time.Now().Add(24 * time.Hour)

		user := models.User{
			Email:             req.Email,
			VerificationToken: verificationToken,
			TokenExpiresAt:    &tokenExpires,
		}

		if err := db.Create(&user).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create user")
			return
		}

		if emailSender != nil {
			go emailSender.SendVerificationEmail(user.Email, verificationToken)
		}

		utils.SuccessResponse(c, http.StatusCreated, "User registered. Please check your email for verification link.", nil)
	}
}

// @Summary User login (send OTP)
// @Description Kirim kode OTP ke email untuk login
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Email"
// @Success 200 {object} utils.Response "OTP sent"
// @Router /api/auth/login [post]
func Login(db *gorm.DB, emailSender *worker.EmailSender) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
			return
		}

		var user models.User
		if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
			utils.SuccessResponse(c, http.StatusOK, "If this email is registered, an OTP has been sent.", nil)
			return
		}

		if !user.IsVerified {
			utils.SuccessResponse(c, http.StatusOK, "If this email is registered, an OTP has been sent.", nil)
			return
		}

		otp := generateOTP()
		otpExpires := time.Now().Add(5 * time.Minute)

		db.Model(&user).Updates(map[string]interface{}{
			"otp_code":       otp,
			"otp_expires_at": otpExpires,
		})

		if emailSender != nil {
			go emailSender.SendOTPEmail(user.Email, otp)
		}

		utils.SuccessResponse(c, http.StatusOK, "If this email is registered, an OTP has been sent.", nil)
	}
}

// @Summary Verify OTP
// @Description Masukkan kode OTP untuk login
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body VerifyOTPRequest true "Email & OTP"
// @Success 200 {object} utils.Response{data=AuthResponse} "Login successful"
// @Router /api/auth/verify-otp [post]
func VerifyOTP(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req VerifyOTPRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
			return
		}

		var user models.User
		if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid email or OTP")
			return
		}

		if user.OTPCode == "" || user.OTPExpiresAt == nil {
			utils.ErrorResponse(c, http.StatusUnauthorized, "No OTP requested. Please request a new one.")
			return
		}

		if time.Now().After(*user.OTPExpiresAt) {
			db.Model(&user).Updates(map[string]interface{}{
				"otp_code":       "",
				"otp_expires_at": nil,
			})
			utils.ErrorResponse(c, http.StatusUnauthorized, "OTP has expired. Please request a new one.")
			return
		}

		if user.OTPCode != req.OTP {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid OTP code")
			return
		}

		db.Model(&user).Updates(map[string]interface{}{
			"otp_code":       "",
			"otp_expires_at": nil,
		})

		token := generateToken(user.ID)
		utils.SuccessResponse(c, http.StatusOK, "Login successful", AuthResponse{
			Token: token,
			User: UserResponse{
				ID:             user.ID,
				Email:          user.Email,
				Role:           user.Role,
				AccountStatus:  user.AccountStatus,
				WarningMessage: user.WarningMessage,
				IsVerified:     user.IsVerified,
			},
		})
	}
}

// @Summary Admin login (send OTP)
// @Description Kirim kode OTP ke email admin untuk login
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Email"
// @Success 200 {object} utils.Response "OTP sent"
// @Router /api/auth/admin/login [post]
func AdminLogin(db *gorm.DB, emailSender *worker.EmailSender) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
			return
		}

		var user models.User
		if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
			utils.SuccessResponse(c, http.StatusOK, "If this email is registered as admin, an OTP has been sent.", nil)
			return
		}

		if user.Role != "admin" {
			utils.SuccessResponse(c, http.StatusOK, "If this email is registered as admin, an OTP has been sent.", nil)
			return
		}

		if !user.IsVerified {
			utils.SuccessResponse(c, http.StatusOK, "If this email is registered as admin, an OTP has been sent.", nil)
			return
		}

		otp := generateOTP()
		otpExpires := time.Now().Add(5 * time.Minute)

		db.Model(&user).Updates(map[string]interface{}{
			"otp_code":       otp,
			"otp_expires_at": otpExpires,
		})

		if emailSender != nil {
			go emailSender.SendOTPEmail(user.Email, otp)
		}

		utils.SuccessResponse(c, http.StatusOK, "If this email is registered as admin, an OTP has been sent.", nil)
	}
}

// @Summary User logout
// @Description Logout user
// @Tags Auth
// @Success 200 {object} utils.Response "Logout successful"
// @Router /api/auth/logout [post]
func Logout(c *gin.Context) {
	utils.SuccessResponse(c, http.StatusOK, "Logout successful", nil)
}

// @Summary Get user profile
// @Description Ambil profil user yang sedang login
// @Tags Auth
// @Security Bearer
// @Produce json
// @Success 200 {object} utils.Response{data=UserResponse} "Profile retrieved"
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

func generateOTP() string {
	return fmt.Sprintf("%06d", 100000+rand.Intn(900000))
}

type VerifyEmailRequest struct {
	Token string `json:"token" binding:"required"`
}

// @Summary Verify email address
// @Description Verifikasi email dan auto-login
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body VerifyEmailRequest true "Verification token"
// @Success 200 {object} utils.Response{data=AuthResponse} "Email verified"
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
			token := generateToken(user.ID)
			utils.SuccessResponse(c, http.StatusOK, "Email already verified", AuthResponse{
				Token: token,
				User: UserResponse{
					ID:             user.ID,
					Email:          user.Email,
					Role:           user.Role,
					AccountStatus:  user.AccountStatus,
					WarningMessage: user.WarningMessage,
					IsVerified:     true,
				},
			})
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

		var freePlan models.SubscriptionPlan
		if err := db.First(&freePlan, "id = ?", "85052fb1-7a58-4951-957f-36ce2e5588f6").Error; err == nil {
			var existing models.UserSubscription
			if db.Where("user_id = ? AND plan_name = ? AND status = 'active'", user.ID, freePlan.Name).First(&existing).Error != nil {
				startDate := time.Now()
				endDate := startDate.AddDate(0, 0, freePlan.DurationDays)

				var rules []models.SubscriptionRule
				db.Where("plan_id = ?", freePlan.ID).Find(&rules)
				rulesMap := map[string]string{}
				for _, r := range rules {
					rulesMap[r.RuleKey] = r.RuleValue
				}
				rulesJSON, _ := json.Marshal(rulesMap)

				subscription := models.UserSubscription{
					UserID:       user.ID,
					PlanName:     freePlan.Name,
					PlanType:     freePlan.Type,
					PlanBenefits: freePlan.Benefits,
					PlanPrice:    freePlan.FinalPrice,
					PlanDuration: freePlan.DurationDays,
					Rules:        string(rulesJSON),
					Status:       "active",
					StartDate:    startDate,
					EndDate:      endDate,
				}
				db.Create(&subscription)
			}
		}

		token := generateToken(user.ID)
		utils.SuccessResponse(c, http.StatusOK, "Email verified successfully", AuthResponse{
			Token: token,
			User: UserResponse{
				ID:             user.ID,
				Email:          user.Email,
				Role:           user.Role,
				AccountStatus:  user.AccountStatus,
				WarningMessage: user.WarningMessage,
				IsVerified:     true,
			},
		})
	}
}

type ResendVerificationRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// @Summary Resend verification email
// @Description Kirim ulang email verifikasi
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body ResendVerificationRequest true "Email address"
// @Success 200 {object} utils.Response "Verification email sent"
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

		newToken := uuid.New().String()
		newExpiry := time.Now().Add(24 * time.Hour)

		db.Model(&user).Updates(map[string]interface{}{
			"verification_token": newToken,
			"token_expires_at":   newExpiry,
		})

		if emailSender != nil {
			go emailSender.SendVerificationEmail(user.Email, newToken)
		}

		utils.SuccessResponse(c, http.StatusOK, "Verification email sent. Please check your inbox.", nil)
	}
}
