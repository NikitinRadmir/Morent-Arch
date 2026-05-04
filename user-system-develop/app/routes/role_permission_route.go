package routes

import (
	"user-system/app/controller"

	"github.com/gin-gonic/gin"
)

func RegisterRolePermissionRoutes(router *gin.RouterGroup, rolePermissionController *controller.RolePermissionController, authMiddleware gin.HandlerFunc) {
	// Маршруты для разрешений ролей
	rolePermissions := router.Group("/role-permissions")
	rolePermissions.Use(authMiddleware)
	{
		// Назначение/удаление разрешений
		rolePermissions.POST("/assign", rolePermissionController.AssignPermission)

		// Получение разрешений роли и ролей разрешения
		rolePermissions.GET("/roles/:roleId/permissions", rolePermissionController.GetRolePermissions)
		rolePermissions.GET("/permissions/:permissionId/roles", rolePermissionController.GetPermissionRoles)

		// Проверка разрешения
		rolePermissions.GET("/roles/:roleId/permissions/:permissionId/check", rolePermissionController.CheckPermission)

		// Массовые операции
		rolePermissions.POST("/roles/:roleId/permissions/bulk-assign", rolePermissionController.BulkAssignPermissions)
		rolePermissions.POST("/roles/:roleId/permissions/bulk-remove", rolePermissionController.BulkRemovePermissions)
	}
}
