package routes

import (
	"context"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
)

func RegisterGraphQLRoutes(router *gin.Engine, graphqlServer *handler.Server, tokenService interface{}) {
	router.POST("/query", func(c *gin.Context) {
		if graphqlServer == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "graphql server is not initialized",
			})
			return
		}

		if userID, exists := c.Get("user_id"); exists {
			ctx := context.WithValue(c.Request.Context(), "user_id", userID)
			c.Request = c.Request.WithContext(ctx)
		}
		graphqlServer.ServeHTTP(c.Writer, c.Request)
	})

	router.GET("/", func(c *gin.Context) {
		if graphqlServer == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "graphql server is not initialized",
			})
			return
		}
		playground.Handler("GraphQL playground", "/query").ServeHTTP(c.Writer, c.Request)
	})

	router.GET("/graphql", func(c *gin.Context) {
		if graphqlServer == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "graphql server is not initialized",
			})
			return
		}
		playground.Handler("GraphQL playground", "/query").ServeHTTP(c.Writer, c.Request)
	})
}
