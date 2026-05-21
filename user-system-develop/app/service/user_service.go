package service

import (
	"errors"
	"log/slog"
	"time"
	"user-system/app/exceptions"

	"user-system/app/dto"
	"user-system/app/models"
	"user-system/app/repositories"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService interface {
	GetUsers(page, limit int, companyID uuid.UUID, isActive *bool) ([]dto.UserResponse, int64, error)
	GetUserByID(id uuid.UUID) (*dto.UserResponse, error)
	CreateUser(req dto.CreateUserRequest, createdBy uuid.UUID) (*dto.UserResponse, error)
	UpdateUser(id uuid.UUID, req dto.UpdateUserRequest, updatedBy uuid.UUID) (*dto.UserResponse, error)
	DeleteUser(id, deletedBy uuid.UUID) error
	ChangePassword(id uuid.UUID, oldPassword, newPassword string, changedBy uuid.UUID) error
	GetAdminDashboard(companyID uuid.UUID) (map[string]interface{}, error)
	Authenticate(email, password string) (*models.User, error)
	GetUsersByRole(roleID uuid.UUID) (*dto.UserByRoleResponse, error)
}

type userService struct {
	userRepo     repositories.UserRepository
	companyRepo  repositories.CompanyRepository
	roleRepo     repositories.RoleRepository
	passwordCost int
}

func NewUserService(
	userRepo repositories.UserRepository,
	companyRepo repositories.CompanyRepository,
	roleRepo repositories.RoleRepository,
) UserService {
	return &userService{
		userRepo:     userRepo,
		companyRepo:  companyRepo,
		roleRepo:     roleRepo,
		passwordCost: bcrypt.DefaultCost,
	}
}

func (s *userService) GetUsers(page, limit int, companyID uuid.UUID, isActive *bool) ([]dto.UserResponse, int64, error) {
	offset := (page - 1) * limit

	users, err := s.userRepo.FindAll(limit, offset, companyID, isActive)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.userRepo.Count(companyID, isActive)
	if err != nil {
		return nil, 0, err
	}

	var userResponses []dto.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, dto.UserResponse{
			ID:         user.ID,
			CompanyID:  user.CompanyID,
			Email:      user.Email,
			FirstName:  user.FirstName,
			LastName:   user.LastName,
			Phone:      user.Phone,
			Department: user.Department,
			Position:   user.Position,
			IsActive:   user.IsActive,
			LastLogin:  user.LastLogin,
			CreatedAt:  user.CreatedAt,
			UpdatedAt:  user.UpdatedAt,
		})
	}

	return userResponses, total, nil
}

func (s *userService) GetUserByID(id uuid.UUID) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	return &dto.UserResponse{
		ID:         user.ID,
		CompanyID:  user.CompanyID,
		Email:      user.Email,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		Phone:      user.Phone,
		Department: user.Department,
		Position:   user.Position,
		IsActive:   user.IsActive,
		LastLogin:  user.LastLogin,
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
	}, nil
}

func (s *userService) CreateUser(req dto.CreateUserRequest, createdBy uuid.UUID) (*dto.UserResponse, error) {
	company, err := s.companyRepo.GetByID(req.CompanyID)
	if err != nil {
		return nil, errors.New("company not found")
	}

	existingUser, _ := s.userRepo.FindByEmailAndCompany(req.Email, req.CompanyID)
	if existingUser != nil {
		return nil, errors.New("user with this email already exists in this company")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), s.passwordCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		CompanyID:  req.CompanyID,
		Email:      req.Email,
		Password:   string(hashedPassword),
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		Phone:      req.Phone,
		Department: req.Department,
		Position:   req.Position,
		IsActive:   true,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	if company.OwnerID == nil {
		company.OwnerID = &user.ID
		if err := s.companyRepo.Update(company); err != nil {
			slog.Warn("failed to set company owner", "error", err)
		}

		if err := s.assignSuperAdminRole(user); err != nil {
			slog.Warn("failed to assign super admin role", "error", err)
		}
	}

	return &dto.UserResponse{
		ID:         user.ID,
		CompanyID:  user.CompanyID,
		Email:      user.Email,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		Phone:      user.Phone,
		Department: user.Department,
		Position:   user.Position,
		IsActive:   user.IsActive,
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
	}, nil
}

func (s *userService) UpdateUser(id uuid.UUID, req dto.UpdateUserRequest, updatedBy uuid.UUID) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if updatedBy != user.ID {
		updater, err := s.userRepo.FindByID(updatedBy)
		if err != nil {
			return nil, errors.New("updater not found")
		}

		if updater.CompanyID != user.CompanyID && !s.hasSuperAdminAccess(updater) {
			return nil, errors.New("permission denied: cannot update user from another company")
		}

		if !s.hasAdminAccess(updater) {
			return nil, errors.New("permission denied: admin role required")
		}
	}

	if req.Email != "" && req.Email != user.Email {
		existingUser, _ := s.userRepo.FindByEmailAndCompany(req.Email, user.CompanyID)
		if existingUser != nil && existingUser.ID != id {
			return nil, errors.New("email already in use in this company")
		}
		user.Email = req.Email
	}

	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Department != "" {
		user.Department = req.Department
	}
	if req.Position != "" {
		user.Position = req.Position
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	return &dto.UserResponse{
		ID:         user.ID,
		CompanyID:  user.CompanyID,
		Email:      user.Email,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		Phone:      user.Phone,
		Department: user.Department,
		Position:   user.Position,
		IsActive:   user.IsActive,
		LastLogin:  user.LastLogin,
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
	}, nil
}

func (s *userService) DeleteUser(id, deletedBy uuid.UUID) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return errors.New("user not found")
	}

	if deletedBy != user.ID {
		deleter, err := s.userRepo.FindByID(deletedBy)
		if err != nil {
			return errors.New("permission denied")
		}

		if deleter.CompanyID != user.CompanyID && !s.hasSuperAdminAccess(deleter) {
			return errors.New("permission denied")
		}
	}

	if err := s.userRepo.Delete(id); err != nil {
		return err
	}

	if user.IsActive {
		user.IsActive = false
		s.userRepo.Update(user)
	}

	slog.Info("user soft deleted", "user_id", id, "deleted_by", deletedBy)
	return nil
}

func (s *userService) ChangePassword(id uuid.UUID, oldPassword, newPassword string, changedBy uuid.UUID) error {
	if changedBy != id {
		return errors.New("permission denied")
	}

	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.New("invalid old password")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), s.passwordCost)
	if err != nil {
		return err
	}

	user.Password = string(hashedPassword)
	return s.userRepo.Update(user)
}

func (s *userService) GetAdminDashboard(companyID uuid.UUID) (map[string]interface{}, error) {
	totalUsers, err := s.userRepo.Count(companyID, nil)
	if err != nil {
		return nil, err
	}

	activeUsers, err := s.userRepo.Count(companyID, boolPtr(true))
	if err != nil {
		return nil, err
	}

	recentUsers, err := s.userRepo.CountRecent(30, companyID)
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_users":       totalUsers,
		"active_users":      activeUsers,
		"inactive_users":    totalUsers - activeUsers,
		"new_users_30_days": recentUsers,
	}

	return stats, nil
}

func (s *userService) Authenticate(email, password string) (*models.User, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid credentials")
		}
		return nil, err
	}

	if !user.IsActive {
		return nil, errors.New("account is disabled")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	now := time.Now()
	user.LastLogin = &now
	if err := s.userRepo.Update(user); err != nil {
		slog.Warn("failed to update last login", "user_id", user.ID, "error", err)
	}

	return user, nil
}

func (s *userService) hasSuperAdminAccess(user *models.User) bool {
	company, err := s.companyRepo.GetByID(user.CompanyID)
	if err != nil {
		return false
	}

	if company.OwnerID != nil && *company.OwnerID == user.ID {
		return true
	}

	return s.checkUserRole(user, "super_admin")
}

func (s *userService) hasAdminAccess(user *models.User) bool {
	return s.checkUserRole(user, "admin") || s.hasSuperAdminAccess(user)
}

func (s *userService) checkUserRole(user *models.User, roleName string) bool {
	role, err := s.roleRepo.GetByNameAndCompany(roleName, user.CompanyID)
	if err != nil {
		return false
	}

	for _, userRole := range user.UserRoles {
		if userRole.RoleID == role.ID {
			return true
		}
	}

	return false
}

func (s *userService) assignSuperAdminRole(user *models.User) error {
	role, err := s.roleRepo.GetByNameAndCompany("super_admin", user.CompanyID)
	if err != nil {
		role = &models.Role{
			ID:        uuid.New(),
			CompanyID: user.CompanyID,
			Name:      "super_admin",
			IsSystem:  true,
		}
		if err := s.roleRepo.Create(role); err != nil {
			return err
		}
	}

	userRole := &models.UserRole{
		ID:         uuid.New(),
		UserID:     user.ID,
		RoleID:     role.ID,
		AssignedBy: user.ID,
		AssignedAt: time.Now(),
	}

	return s.roleRepo.AssignRoleToUser(userRole)
}

func boolPtr(b bool) *bool {
	return &b
}

func (s *userService) GetUsersByRole(roleID uuid.UUID) (*dto.UserByRoleResponse, error) {
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		return nil, &exceptions.RoleNotFoundError{RoleName: roleID.String()}
	}

	users, err := s.roleRepo.GetUsersByRole(roleID)
	if err != nil {
		return nil, err
	}

	var userDTOs []dto.UserWithDetails
	for _, user := range users {
		userDTOs = append(userDTOs, dto.UserWithDetails{
			ID:         user.ID,
			Email:      user.Email,
			FirstName:  user.FirstName,
			LastName:   user.LastName,
			Department: user.Department,
			Position:   user.Position,
			IsActive:   user.IsActive,
			LastLogin:  user.LastLogin,
		})
	}

	return &dto.UserByRoleResponse{
		RoleID:     role.ID,
		RoleName:   role.Name,
		Users:      userDTOs,
		TotalCount: len(users),
	}, nil
}
