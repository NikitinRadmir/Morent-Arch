//go:build wireinject
// +build wireinject

package di

import (
	"user-system/app/controller"
	"user-system/app/repositories"
	"user-system/app/rpc"
	"user-system/app/service"

	"github.com/google/wire"
	"gorm.io/gorm"
)

func InitAuthController(db *gorm.DB) *controller.AuthController {
	wire.Build(
		repositories.NewUserRepository,
		repositories.NewCompanyRepository,
		repositories.NewRoleRepository,
		service.NewAuthService,
		controller.NewAuthController,
	)
	return nil
}

func InitUserController(db *gorm.DB) *controller.UserController {
	wire.Build(
		repositories.NewUserRepository,
		repositories.NewCompanyRepository,
		repositories.NewRoleRepository,
		service.NewUserService,
		controller.NewUserController,
	)
	return nil
}

func InitRoleController(db *gorm.DB) *controller.RoleController {
	wire.Build(
		repositories.NewRoleRepository,
		repositories.NewCompanyRepository,
		repositories.NewPermissionRepository,
		service.NewRoleService,
		controller.NewRoleController,
	)
	return nil
}

func InitCompanyController(db *gorm.DB) *controller.CompanyController {
	wire.Build(
		repositories.NewCompanyRepository,
		service.NewCompanyService,
		controller.NewCompanyController,
	)
	return nil
}

func InitRolePermissionController(db *gorm.DB) *controller.RolePermissionController {
	wire.Build(
		repositories.NewRolePermissionRepository,
		repositories.NewRoleRepository,
		repositories.NewPermissionRepository,
		service.NewRolePermissionService,
		controller.NewRolePermissionController,
	)
	return nil
}

func InitUserService(db *gorm.DB) service.UserService {
	wire.Build(
		repositories.NewUserRepository,
		repositories.NewCompanyRepository,
		repositories.NewRoleRepository,
		service.NewUserService,
	)
	return nil
}

func InitRoleService(db *gorm.DB) service.RoleService {
	wire.Build(
		repositories.NewRoleRepository,
		repositories.NewCompanyRepository,
		repositories.NewPermissionRepository,
		service.NewRoleService,
	)
	return nil
}

func InitRPCMethods(db *gorm.DB) *rpc.RPCMethods {
	wire.Build(
		repositories.NewCompanyRepository,
		repositories.NewUserRepository,
		repositories.NewRoleRepository,
		repositories.NewPermissionRepository,
		repositories.NewRolePermissionRepository,

		service.NewCompanyService,
		service.NewUserService,
		service.NewRoleService,
		service.NewRolePermissionService,

		rpc.NewRPCMethods,
	)
	return nil
}

func InitRPCServer(db *gorm.DB) *rpc.RPCServer {
	wire.Build(
		repositories.NewCompanyRepository,
		repositories.NewUserRepository,
		repositories.NewRoleRepository,
		repositories.NewPermissionRepository,
		repositories.NewRolePermissionRepository,

		service.NewCompanyService,
		service.NewUserService,
		service.NewRoleService,
		service.NewRolePermissionService,

		rpc.NewRPCMethods,
		rpc.NewRPCServer,
	)
	return nil
}

//func InitGraphQLResolver(db *gorm.DB) *resolvers.Resolver {
//	wire.Build(
//		repositories.NewUserRepository,
//		repositories.NewRoleRepository,
//		repositories.NewPermissionRepository,
//		repositories.NewCompanyRepository,
//		repositories.NewRolePermissionRepository,
//		service.NewUserService,
//		service.NewRoleService,
//		service.NewRolePermissionService,
//		resolvers.NewResolver,
//	)
//	return &resolvers.Resolver{}
//}
