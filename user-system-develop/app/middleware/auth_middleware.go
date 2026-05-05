package middleware

import (
	"net/http"
	"strings"

	"user-system/app/token"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func AuthRequired(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Authorization header is required",
		})
		return
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Invalid authorization format. Expected 'Bearer <token>'",
		})
		return
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := token.ValidateToken(tokenStr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Invalid or expired token",
		})
		return
	}

	userUUID, err := uuid.Parse(claims.UserId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Invalid user id in token",
		})
		return
	}

	c.Set("userID", userUUID)
	c.Set("roles", claims.Roles)
	c.Set("claims", claims)

	c.Next()
}
