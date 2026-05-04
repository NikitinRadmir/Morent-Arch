package routes

import (
	"user-system/app/controller"

	"github.com/gin-gonic/gin"
)

func RegisterRoleRoutes(router *gin.RouterGroup, roleController *controller.RoleController, authMiddleware gin.HandlerFunc) {
	roles := router.Group("/roles")
	roles.Use(authMiddleware)
	{
		// CRUD операций для ролей
		roles.POST("", roleController.CreateRole)
		roles.GET("/:id", roleController.GetRole)
		roles.PUT("/:id", roleController.UpdateRole)
		roles.DELETE("/:id", roleController.DeleteRole)

		// Роли компании
		roles.GET("/company/:companyId", roleController.GetCompanyRoles)
	}
}
