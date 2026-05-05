package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"user-system/app/token"
)

func GraphQLAuthMiddleware(tokenService *token.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/graphql/playground" {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		claims, err := tokenService.ValidateToken(parts[1])
		if err != nil {
			c.Next()
			return
		}

		userID, _ := uuid.Parse(claims.UserId)
		ctx := context.WithValue(c.Request.Context(), "user_id", userID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
