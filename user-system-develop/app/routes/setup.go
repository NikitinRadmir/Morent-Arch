package routes

import (
	"user-system/app/controller"
	"user-system/app/middleware"
	"user-system/app/rpc"
	"user-system/app/swagger"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine,
	userController *controller.UserController,
	roleController *controller.RoleController,
	companyController *controller.CompanyController,
	rolePermissionController *controller.RolePermissionController,
	authController *controller.AuthController,
	rpcServer *rpc.RPCServer,
	graphqlServer *handler.Server,
) {

	api := router.Group("/api")

	swagger.Mount(router)

	RegisterAuthRoutes(api, authController, middleware.AuthRequired)

	api.POST("/companies", companyController.CreateCompany)

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	protected := api.Group("")
	protected.Use(middleware.AuthRequired)
	{
		RegisterUserRoutes(protected, userController, middleware.AuthRequired)
		RegisterRoleRoutes(protected, roleController, middleware.AuthRequired)
		RegisterCompanyRoutes(protected, companyController)
		RegisterRolePermissionRoutes(protected, rolePermissionController, middleware.AuthRequired)
	}

	RegisterRPCRoutes(router, rpcServer)
	RegisterGraphQLRoutes(router, graphqlServer, nil)
}
