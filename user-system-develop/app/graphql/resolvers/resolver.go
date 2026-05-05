package resolvers

// THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

import (
	"user-system/app/dto"
	"user-system/app/graphql/generated"
	"user-system/app/models"
	"user-system/app/service"
	"user-system/app/token"
	"context"

	"github.com/google/uuid"
)

type Resolver struct {
	userService           service.UserService
	roleService           service.RoleService
	rolePermissionService service.RolePermissionService
}

func NewResolver(
	userService service.UserService,
	roleService service.RoleService,
	rolePermissionService service.RolePermissionService,
) *Resolver {
	return &Resolver{
		userService:           userService,
		roleService:           roleService,
		rolePermissionService: rolePermissionService,
	}
}

// Login is the resolver for the login field.
func (r *mutationResolver) Login(ctx context.Context, email string, password string) (*generated.AuthPayload, error) {
	user, err := r.userService.Authenticate(email, password)
	if err != nil {
		return nil, err
	}

	roles, err := r.roleService.GetUserRoles(user.ID)
	if err != nil {
		roles = []models.Role{}
	}

	var roleNames []string
	for _, role := range roles {
		roleNames = append(roleNames, role.Name)
	}

	tokenString, err := token.GenerateTokenWithRoles(user.ID.String(), roleNames)
	if err != nil {
		return nil, err
	}

	return &generated.AuthPayload{
		Token: tokenString,
		User:  user,
	}, nil
}

// Logout is the resolver for the logout field.
func (r *mutationResolver) Logout(ctx context.Context) (bool, error) {
	return true, nil
}

// Roles is the resolver for the roles field.
func (r *permissionResolver) Roles(ctx context.Context, obj *models.Permission) ([]*models.Role, error) {
	roles, err := r.rolePermissionService.GetRolesWithPermission(obj.ID)
	if err != nil {
		return nil, err
	}

	var result []*models.Role
	for i := range roles {
		result = append(result, &roles[i])
	}

	return result, nil
}

// Me is the resolver for the me field.
func (r *queryResolver) Me(ctx context.Context) (*models.User, error) {
	// TODO: Получить userID из JWT
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	userDTO, err := r.userService.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	return convertUserDTOToModel(userDTO), nil
}

// User is the resolver for the user field.
func (r *queryResolver) User(ctx context.Context, id uuid.UUID) (*models.User, error) {
	userDTO, err := r.userService.GetUserByID(id)
	if err != nil {
		return nil, err
	}

	return convertUserDTOToModel(userDTO), nil
}

// Users is the resolver for the users field.
func (r *queryResolver) Users(ctx context.Context, limit *int, offset *int, companyID *uuid.UUID, isActive *bool) (*generated.UserList, error) {
	page := 1
	perPage := 10

	if limit != nil {
		perPage = *limit
	}
	if offset != nil {
		page = (*offset / perPage) + 1
	}

	var companyIDValue uuid.UUID
	if companyID != nil {
		companyIDValue = *companyID
	} else {
		companyIDValue = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	}

	users, total, err := r.userService.GetUsers(page, perPage, companyIDValue, isActive)
	if err != nil {
		return nil, err
	}

	var items []*models.User
	for _, userDTO := range users {
		items = append(items, convertUserDTOToModel(&userDTO))
	}

	return &generated.UserList{
		Items: items,
		Total: int(total),
	}, nil
}

// Roles is the resolver for the roles field.
func (r *queryResolver) Roles(ctx context.Context, companyID *uuid.UUID) ([]*models.Role, error) {
	var companyIDValue uuid.UUID
	if companyID != nil {
		companyIDValue = *companyID
	} else {
		companyIDValue = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	}

	roles, err := r.roleService.GetCompanyRoles(companyIDValue)
	if err != nil {
		return nil, err
	}

	var result []*models.Role
	for i := range roles {
		result = append(result, &roles[i])
	}

	return result, nil
}

// Permissions is the resolver for the permissions field.
func (r *queryResolver) Permissions(ctx context.Context) ([]*models.Permission, error) {
	// TODO: Реализовать получение всех разрешений
	return []*models.Permission{}, nil
}

// Permissions is the resolver for the permissions field.
func (r *roleResolver) Permissions(ctx context.Context, obj *models.Role) ([]*models.Permission, error) {
	permissions, err := r.rolePermissionService.GetPermissionsForRole(obj.ID)
	if err != nil {
		return nil, err
	}

	var result []*models.Permission
	for i := range permissions {
		result = append(result, &permissions[i])
	}

	return result, nil
}

// Users is the resolver for the users field.
func (r *roleResolver) Users(ctx context.Context, obj *models.Role) ([]*models.User, error) {
	roleUsers, err := r.roleService.GetRoleUsers(obj.ID)
	if err != nil {
		return nil, err
	}

	var result []*models.User
	for _, userInfo := range roleUsers.Users {
		userDTO, err := r.userService.GetUserByID(userInfo.ID)
		if err != nil {
			continue
		}
		result = append(result, convertUserDTOToModel(userDTO))
	}

	return result, nil
}

// Roles is the resolver for the roles field.
func (r *userResolver) Roles(ctx context.Context, obj *models.User) ([]*models.Role, error) {
	roles, err := r.roleService.GetUserRoles(obj.ID)
	if err != nil {
		return nil, err
	}

	var result []*models.Role
	for i := range roles {
		result = append(result, &roles[i])
	}

	return result, nil
}

// Permissions is the resolver for the permissions field.
func (r *userResolver) Permissions(ctx context.Context, obj *models.User) ([]*models.Permission, error) {
	roles, err := r.roleService.GetUserRoles(obj.ID)
	if err != nil {
		return nil, err
	}

	permissionMap := make(map[uuid.UUID]*models.Permission)

	for _, role := range roles {
		permissions, err := r.rolePermissionService.GetPermissionsForRole(role.ID)
		if err != nil {
			continue
		}

		for _, perm := range permissions {
			if _, exists := permissionMap[perm.ID]; !exists {
				permissionMap[perm.ID] = &models.Permission{
					ID:          perm.ID,
					Name:        perm.Name,
					Description: perm.Description,
				}
			}
		}
	}

	var result []*models.Permission
	for _, perm := range permissionMap {
		result = append(result, perm)
	}

	return result, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Permission returns generated.PermissionResolver implementation.
func (r *Resolver) Permission() generated.PermissionResolver { return &permissionResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// Role returns generated.RoleResolver implementation.
func (r *Resolver) Role() generated.RoleResolver { return &roleResolver{r} }

// User returns generated.UserResolver implementation.
func (r *Resolver) User() generated.UserResolver { return &userResolver{r} }

func convertUserDTOToModel(userDTO *dto.UserResponse) *models.User {
	return &models.User{
		ID:         userDTO.ID,
		CompanyID:  userDTO.CompanyID,
		Email:      userDTO.Email,
		FirstName:  userDTO.FirstName,
		LastName:   userDTO.LastName,
		Phone:      userDTO.Phone,
		Department: userDTO.Department,
		Position:   userDTO.Position,
		IsActive:   userDTO.IsActive,
		LastLogin:  userDTO.LastLogin,
		CreatedAt:  userDTO.CreatedAt,
		UpdatedAt:  userDTO.UpdatedAt,
	}
}

type mutationResolver struct{ *Resolver }
type permissionResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type roleResolver struct{ *Resolver }
type userResolver struct{ *Resolver }
