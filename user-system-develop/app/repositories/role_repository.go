package repositories

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"user-system/app/exceptions"
	"user-system/app/models"
)

type RoleRepository interface {
	Create(role *models.Role) error
	GetByID(id uuid.UUID) (*models.Role, error)
	GetByNameAndCompany(name string, companyID uuid.UUID) (*models.Role, error)
	GetByCompany(companyID uuid.UUID) ([]models.Role, error)
	Update(role *models.Role) error
	Delete(id uuid.UUID) error
	AddPermission(roleID, permissionID uuid.UUID) error
	RemovePermission(roleID, permissionID uuid.UUID) error
	GetPermissions(roleID uuid.UUID) ([]models.Permission, error)
	AssignRoleToUser(userRole *models.UserRole) error
	RevokeRoleFromUser(userID, roleID uuid.UUID) error
	GetUserRoles(userID uuid.UUID) ([]models.UserRole, error)
	GetUsersByRole(roleID uuid.UUID) ([]models.User, error)
	HasRole(userID, roleID uuid.UUID) (bool, error)
	CreateSystemRoles(companyID uuid.UUID) error
	GetRoleWithUsers(roleID uuid.UUID) ([]models.UserRole, error)
}

type roleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) Create(role *models.Role) error {
	return r.db.Create(role).Error
}

func (r *roleRepository) GetByID(id uuid.UUID) (*models.Role, error) {
	var role models.Role
	err := r.db.Preload("Permissions").First(&role, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &exceptions.RoleNotFoundError{RoleName: id.String()}
		}
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) GetByNameAndCompany(name string, companyID uuid.UUID) (*models.Role, error) {
	var role models.Role
	err := r.db.Preload("Permissions").Where("name = ? AND company_id = ?", name, companyID).First(&role).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &exceptions.RoleNotFoundError{RoleName: name}
		}
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) GetByCompany(companyID uuid.UUID) ([]models.Role, error) {
	var roles []models.Role
	err := r.db.Preload("Permissions").Where("company_id = ?", companyID).Find(&roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *roleRepository) Update(role *models.Role) error {
	result := r.db.Save(role)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return &exceptions.RoleNotFoundError{RoleName: role.ID.String()}
	}
	return nil
}

func (r *roleRepository) Delete(id uuid.UUID) error {
	result := r.db.Where("id = ?", id).Delete(&models.Role{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return &exceptions.RoleNotFoundError{RoleName: id.String()}
	}
	return nil
}

func (r *roleRepository) AddPermission(roleID, permissionID uuid.UUID) error {
	return r.db.Exec("INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?)", roleID, permissionID).Error
}

func (r *roleRepository) RemovePermission(roleID, permissionID uuid.UUID) error {
	return r.db.Exec("DELETE FROM role_permissions WHERE role_id = ? AND permission_id = ?", roleID, permissionID).Error
}

func (r *roleRepository) GetPermissions(roleID uuid.UUID) ([]models.Permission, error) {
	var role models.Role
	err := r.db.Preload("Permissions").First(&role, "id = ?", roleID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &exceptions.RoleNotFoundError{RoleName: roleID.String()}
		}
		return nil, err
	}
	return role.Permissions, nil
}

func (r *roleRepository) AssignRoleToUser(userRole *models.UserRole) error {
	existingUserRole := &models.UserRole{}
	err := r.db.Where("user_id = ? AND role_id = ?", userRole.UserID, userRole.RoleID).First(existingUserRole).Error

	if err == nil {
		existingUserRole.AssignedBy = userRole.AssignedBy
		existingUserRole.AssignedAt = time.Now()
		return r.db.Save(existingUserRole).Error
	}

	return r.db.Create(userRole).Error
}

func (r *roleRepository) RevokeRoleFromUser(userID, roleID uuid.UUID) error {
	result := r.db.Where("user_id = ? AND role_id = ?", userID, roleID).Delete(&models.UserRole{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return &exceptions.UserRoleNotFoundError{UserID: userID.String(), RoleID: roleID.String()}
	}
	return nil
}

func (r *roleRepository) GetUserRoles(userID uuid.UUID) ([]models.UserRole, error) {
	var userRoles []models.UserRole
	err := r.db.Preload("Role").Preload("Role.Permissions").Where("user_id = ?", userID).Find(&userRoles).Error
	if err != nil {
		return nil, err
	}
	return userRoles, nil
}

func (r *roleRepository) GetUsersByRole(roleID uuid.UUID) ([]models.User, error) {
	var users []models.User
	err := r.db.Joins("JOIN user_roles ON users.id = user_roles.user_id").
		Where("user_roles.role_id = ?", roleID).
		Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *roleRepository) HasRole(userID, roleID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.UserRole{}).
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Count(&count).Error

	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *roleRepository) CreateSystemRoles(companyID uuid.UUID) error {
	systemRoles := []models.Role{
		{
			ID:          uuid.New(),
			CompanyID:   companyID,
			Name:        "super_admin",
			Description: "Полный доступ ко всем функциям системы",
			IsSystem:    true,
		},
		{
			ID:          uuid.New(),
			CompanyID:   companyID,
			Name:        "admin",
			Description: "Администратор компании",
			IsSystem:    true,
		},
		{
			ID:          uuid.New(),
			CompanyID:   companyID,
			Name:        "manager",
			Description: "Менеджер компании",
			IsSystem:    true,
		},
		{
			ID:          uuid.New(),
			CompanyID:   companyID,
			Name:        "employee",
			Description: "Сотрудник компании",
			IsSystem:    true,
		},
	}

	for _, role := range systemRoles {
		existingRole, err := r.GetByNameAndCompany(role.Name, companyID)
		if err != nil || existingRole == nil {
			if err := r.Create(&role); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *roleRepository) GetRoleWithUsers(roleID uuid.UUID) ([]models.UserRole, error) {
	var userRoles []models.UserRole
	err := r.db.
		Where("role_id = ?", roleID).
		Preload("User").
		Order("assigned_at DESC").
		Find(&userRoles).Error
	return userRoles, err
}
