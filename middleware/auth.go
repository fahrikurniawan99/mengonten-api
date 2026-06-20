package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"mengonten-api/config"
	"mengonten-api/models"
	"mengonten-api/utils"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Missing authorization header")
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid authorization header format")
			c.Abort()
			return
		}

		jwtConfig := config.GetJWTConfig()
		claims := jwt.MapClaims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtConfig.Secret), nil
		})

		if err != nil || !token.Valid {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid or expired token")
			c.Abort()
			return
		}

		userID, ok := claims["user_id"]
		if !ok {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid token claims")
			c.Abort()
			return
		}

		userIDStr, ok := userID.(string)
		if !ok {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user_id format")
			c.Abort()
			return
		}

		parsedUUID, err := uuid.Parse(userIDStr)
		if err != nil {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user_id UUID")
			c.Abort()
			return
		}

		c.Set("user_id", parsedUUID)
		c.Next()
	}
}

func AuthWithAccountStatusCheck(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Missing authorization header")
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid authorization header format")
			c.Abort()
			return
		}

		jwtConfig := config.GetJWTConfig()
		claims := jwt.MapClaims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtConfig.Secret), nil
		})

		if err != nil || !token.Valid {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid or expired token")
			c.Abort()
			return
		}

		userID, ok := claims["user_id"]
		if !ok {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid token claims")
			c.Abort()
			return
		}

		userIDStr, ok := userID.(string)
		if !ok {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user_id format")
			c.Abort()
			return
		}

		parsedUUID, err := uuid.Parse(userIDStr)
		if err != nil {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user_id UUID")
			c.Abort()
			return
		}

		var user models.User
		if err := db.First(&user, parsedUUID).Error; err != nil {
			utils.ErrorResponse(c, http.StatusUnauthorized, "User not found")
			c.Abort()
			return
		}

		switch user.AccountStatus {
		case "suspended":
			if user.SuspendExpiresAt != nil && time.Now().After(*user.SuspendExpiresAt) {
				db.Model(&user).Updates(map[string]interface{}{
					"account_status":     "active",
					"suspend_reason":     "",
					"suspend_expires_at": nil,
				})
				c.Set("user_id", parsedUUID)
				c.Set("user_role", user.Role)
				c.Set("account_status", "active")
				c.Next()
				return
			}
			msg := "Akun anda ditangguhkan"
			if user.SuspendReason != "" {
				msg += ": " + user.SuspendReason
			}
			utils.ErrorResponse(c, http.StatusForbidden, msg)
			c.Abort()
			return

		case "deactivated":
			utils.ErrorResponse(c, http.StatusForbidden, "Akun anda tidak aktif. Hubungi admin untuk mengaktifkan kembali.")
			c.Abort()
			return
		}

		c.Set("user_id", parsedUUID)
		c.Set("user_role", user.Role)
		c.Set("account_status", user.AccountStatus)
		if user.WarningMessage != "" {
			c.Set("warning_message", user.WarningMessage)
		}
		c.Next()
	}
}
