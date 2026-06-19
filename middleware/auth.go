package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"mengonten-api/config"
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
