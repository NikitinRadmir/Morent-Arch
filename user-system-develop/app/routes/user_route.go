package routes

import (
	"user-system/app/controller"

	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(router *gin.RouterGroup, userController *controller.UserController, authMiddleware gin.HandlerFunc) {
	users := router.Group("/users")
	{
		protected := users.Group("")
		protected.Use(authMiddleware)
		{
			protected.GET("", userController.GetUsers)
			protected.POST("", userController.CreateUser)
			// Статичные сегменты до /:id, иначе Gin сопоставит "profile" как :id
			protected.GET("/profile", userController.GetUserProfile)
			protected.PUT("/profile", userController.UpdateUserProfile)

			protected.GET("/:id", userController.GetUserByID)
			protected.PUT("/:id", userController.UpdateUser)
			protected.DELETE("/:id", userController.DeleteUser)
			protected.PUT("/:id/change-password", userController.ChangePassword)
		}
	}
}
