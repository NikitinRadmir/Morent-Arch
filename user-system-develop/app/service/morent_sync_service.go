package service

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	morentevents "morent-events"
	"user-system/app/models"
	"user-system/app/repositories"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MorentSyncService применяет события Morent → user-system.
type MorentSyncService struct {
	db           *gorm.DB
	userRepo     repositories.UserRepository
	companyRepo  repositories.CompanyRepository
	roleRepo     repositories.RoleRepository
	processedRepo repositories.ProcessedEventRepository
	defaultCompany string
}

func NewMorentSyncService(
	db *gorm.DB,
	userRepo repositories.UserRepository,
	companyRepo repositories.CompanyRepository,
	roleRepo repositories.RoleRepository,
	processedRepo repositories.ProcessedEventRepository,
	defaultCompany string,
) *MorentSyncService {
	if strings.TrimSpace(defaultCompany) == "" {
		defaultCompany = "Morent"
	}
	return &MorentSyncService{
		db:             db,
		userRepo:       userRepo,
		companyRepo:    companyRepo,
		roleRepo:       roleRepo,
		processedRepo:  processedRepo,
		defaultCompany: defaultCompany,
	}
}

func (s *MorentSyncService) Handle(raw []byte) error {
	env, err := morentevents.UnmarshalEnvelope(raw)
	if err != nil {
		return err
	}

	exists, err := s.processedRepo.Exists(env.EventID)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	switch env.EventType {
	case morentevents.EventUserRegistered:
		data, err := morentevents.UnmarshalData[morentevents.UserRegistered](env)
		if err != nil {
			return err
		}
		err = s.handleRegistered(*data)
	case morentevents.EventUserProfileUpdated:
		data, err := morentevents.UnmarshalData[morentevents.UserProfileUpdated](env)
		if err != nil {
			return err
		}
		err = s.handleProfileUpdated(*data)
	case morentevents.EventUserPasswordChanged:
		data, err := morentevents.UnmarshalData[morentevents.UserPasswordChanged](env)
		if err != nil {
			return err
		}
		err = s.handlePasswordChanged(*data)
	case morentevents.EventUserDeactivated:
		data, err := morentevents.UnmarshalData[morentevents.UserDeactivated](env)
		if err != nil {
			return err
		}
		err = s.handleDeactivated(*data)
	default:
		return fmt.Errorf("unsupported event type: %s", env.EventType)
	}

	if err != nil {
		return err
	}
	return s.processedRepo.Mark(env.EventID, env.EventType)
}

func (s *MorentSyncService) handleRegistered(data morentevents.UserRegistered) error {
	if data.MorentUserID == 0 || strings.TrimSpace(data.Email) == "" {
		return errors.New("invalid user.registered payload")
	}
	if strings.TrimSpace(data.PasswordHash) == "" {
		return errors.New("password_hash is required (plaintext passwords are not allowed)")
	}

	email := strings.ToLower(strings.TrimSpace(data.Email))
	if existing, _ := s.userRepo.FindByMorentUserIDUnscoped(data.MorentUserID); existing != nil {
		existing.Password = data.PasswordHash
		existing.IsActive = true
		existing.DeletedAt = gorm.DeletedAt{}
		first, last := splitName(data.Name)
		existing.FirstName = first
		existing.LastName = last
		existing.Position = strings.TrimSpace(data.Position)
		return s.userRepo.UpdateUnscoped(existing)
	}

	existing, err := s.userRepo.FindByEmailUnscoped(email)
	if err != nil {
		return err
	}
	if existing != nil {
		if existing.MorentUserID != nil && *existing.MorentUserID != data.MorentUserID {
			return fmt.Errorf("email %s is already linked to morent_user_id %d", email, *existing.MorentUserID)
		}
		existing.MorentUserID = &data.MorentUserID
		existing.Password = data.PasswordHash
		existing.IsActive = true
		existing.DeletedAt = gorm.DeletedAt{}
		first, last := splitName(data.Name)
		existing.FirstName = first
		existing.LastName = last
		existing.Position = strings.TrimSpace(data.Position)
		if err := s.userRepo.UpdateUnscoped(existing); err != nil {
			return err
		}
		return s.assignRole(existing.ID, existing.CompanyID, mapMorentRole(data.Role))
	}

	company, err := s.ensureCompany(firstNonEmpty(data.CompanyName, s.defaultCompany))
	if err != nil {
		return err
	}

	first, last := splitName(data.Name)
	user := &models.User{
		MorentUserID: &data.MorentUserID,
		CompanyID:    company.ID,
		Email:        email,
		Password:     data.PasswordHash,
		FirstName:    first,
		LastName:     last,
		Department:   "",
		Position:     strings.TrimSpace(data.Position),
		IsActive:     true,
	}
	if err := s.userRepo.Create(user); err != nil {
		return err
	}
	return s.assignRole(user.ID, company.ID, mapMorentRole(data.Role))
}

func (s *MorentSyncService) handleProfileUpdated(data morentevents.UserProfileUpdated) error {
	user, err := s.findUser(data.MorentUserID, data.Email)
	if err != nil {
		return err
	}
	if data.Name != nil {
		first, last := splitName(*data.Name)
		user.FirstName = first
		user.LastName = last
	}
	if data.Position != nil {
		user.Position = strings.TrimSpace(*data.Position)
	}
	if data.IsActive != nil {
		user.IsActive = *data.IsActive
	}
	return s.userRepo.Update(user)
}

func (s *MorentSyncService) handlePasswordChanged(data morentevents.UserPasswordChanged) error {
	if strings.TrimSpace(data.PasswordHash) == "" {
		return errors.New("password_hash is required")
	}
	user, err := s.findUser(data.MorentUserID, data.Email)
	if err != nil {
		return err
	}
	user.Password = data.PasswordHash
	return s.userRepo.Update(user)
}

func (s *MorentSyncService) handleDeactivated(data morentevents.UserDeactivated) error {
	user, err := s.findUserIncludeDeleted(data.MorentUserID, data.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	// Освобождаем email для повторной регистрации в Morent (unique index в user-system).
	user.IsActive = false
	user.Email = fmt.Sprintf("deleted-%s-%s", user.ID.String(), strings.ToLower(strings.TrimSpace(data.Email)))
	if user.MorentUserID != nil {
		user.MorentUserID = nil
	}
	return s.userRepo.UpdateUnscoped(user)
}

func (s *MorentSyncService) findUserIncludeDeleted(morentUserID uint, email string) (*models.User, error) {
	if morentUserID > 0 {
		user, err := s.userRepo.FindByMorentUserID(morentUserID)
		if err == nil && user != nil {
			return user, nil
		}
	}
	normalized := strings.ToLower(strings.TrimSpace(email))
	if normalized == "" {
		return nil, gorm.ErrRecordNotFound
	}
	user, err := s.userRepo.FindByEmailUnscoped(normalized)
	if err != nil || user == nil {
		return nil, fmt.Errorf("user not found for email %s", normalized)
	}
	return user, nil
}

func (s *MorentSyncService) findUser(morentUserID uint, email string) (*models.User, error) {
	if morentUserID > 0 {
		user, err := s.userRepo.FindByMorentUserID(morentUserID)
		if err == nil && user != nil {
			return user, nil
		}
	}
	normalized := strings.ToLower(strings.TrimSpace(email))
	if normalized == "" {
		return nil, errors.New("user not found: missing identifiers")
	}
	user, err := s.userRepo.FindByEmail(normalized)
	if err != nil || user == nil {
		return nil, fmt.Errorf("user not found for email %s", normalized)
	}
	return user, nil
}

func (s *MorentSyncService) ensureCompany(name string) (*models.Company, error) {
	company, err := s.companyRepo.GetByName(name)
	if err != nil {
		return nil, err
	}
	if company != nil {
		_ = s.roleRepo.CreateSystemRoles(company.ID)
		return company, nil
	}
	now := time.Now()
	company = &models.Company{
		ID:        uuid.New(),
		Name:      name,
		Domain:    "morent-" + uuid.New().String(),
		CreatedAt: now,
		UpdatedAt: now,
		IsActive:  true,
	}
	if err := s.companyRepo.Create(company); err != nil {
		company, err = s.companyRepo.GetByName(name)
		if err != nil || company == nil {
			return nil, err
		}
	}
	if err := s.roleRepo.CreateSystemRoles(company.ID); err != nil {
		slog.Warn("create system roles failed", "company_id", company.ID, "error", err)
	}
	return company, nil
}

func (s *MorentSyncService) assignRole(userID, companyID uuid.UUID, roleName string) error {
	role, err := s.roleRepo.GetByNameAndCompany(roleName, companyID)
	if err != nil || role == nil {
		role, err = s.roleRepo.GetByNameAndCompany("employee", companyID)
		if err != nil || role == nil {
			return nil
		}
	}
	has, err := s.roleRepo.HasRole(userID, role.ID)
	if err != nil || has {
		return err
	}
	return s.roleRepo.AssignRoleToUser(&models.UserRole{
		ID:         uuid.New(),
		UserID:     userID,
		RoleID:     role.ID,
		AssignedAt: time.Now(),
	})
}

func splitName(full string) (string, string) {
	full = strings.TrimSpace(full)
	if full == "" {
		return "User", ""
	}
	parts := strings.Fields(full)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], strings.Join(parts[1:], " ")
}

func mapMorentRole(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "admin":
		return "admin"
	case "manager":
		return "manager"
	default:
		return "employee"
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
