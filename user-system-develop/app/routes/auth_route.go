package routes

import (
	"user-system/app/controller"

	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(router *gin.RouterGroup, authController *controller.AuthController, authMiddleware gin.HandlerFunc) {
	auth := router.Group("/auth")
	{
		// Публичные маршруты (не требуют аутентификации)
		auth.POST("/register", authController.Register)
		auth.POST("/login", authController.Login)

		// Защищенные маршруты (требуют аутентификации)
		protected := auth.Group("")
		protected.Use(authMiddleware)
		{
			protected.POST("/logout", authController.Logout)
			protected.POST("/refresh", authController.RefreshToken)
			protected.PUT("/change-password", authController.ChangePassword)
		}
	}
}
