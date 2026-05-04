package service

import (
	"fmt"
	"user-system/app/token"
	"time"

	"user-system/app/dto"
	"user-system/app/exceptions"
	"user-system/app/models"
	"user-system/app/repositories"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(input dto.RegisterRequest) (*dto.AuthResponse, error)
	Login(input dto.UserLoginRequest) (*dto.AuthResponse, error)
	Logout(userID uuid.UUID) error
	RefreshToken(refreshToken string) (*dto.TokenResponse, error)
	ValidateToken(tokenStr string) (*token.Claims, error)
	ChangePassword(userID uuid.UUID, input dto.ChangePasswordRequest) error
}

type authService struct {
	userRepo    repositories.UserRepository
	companyRepo repositories.CompanyRepository
	roleRepo    repositories.RoleRepository
}

func NewAuthService(
	userRepo repositories.UserRepository,
	companyRepo repositories.CompanyRepository,
	roleRepo repositories.RoleRepository,
) AuthService {
	return &authService{
		userRepo:    userRepo,
		companyRepo: companyRepo,
		roleRepo:    roleRepo,
	}
}

func (s *authService) Register(input dto.RegisterRequest) (*dto.AuthResponse, error) {
	company, err := s.companyRepo.GetByName(input.CompanyName)
	if err != nil {
		return nil, err
	}
	if company == nil {
		now := time.Now()
		domain := "company-" + uuid.New().String()
		company = &models.Company{
			ID:          uuid.New(),
			Name:        input.CompanyName,
			Description: "",
			Domain:      domain,
			CreatedAt:   now,
			UpdatedAt:   now,
			IsActive:    true,
		}
		if createErr := s.companyRepo.Create(company); createErr != nil {
			// гонка: другой запрос успел создать компанию с тем же именем
			company, err = s.companyRepo.GetByName(input.CompanyName)
			if err != nil {
				return nil, err
			}
			if company == nil {
				return nil, fmt.Errorf("create company: %w", createErr)
			}
		}
	}

	existingUser, err := s.userRepo.FindByEmail(input.Email)
	if err == nil && existingUser != nil {
		return nil, exceptions.NewUserAlreadyExistsError(input.Email)
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		CompanyID:  company.ID,
		Email:      input.Email,
		Password:   string(hashed),
		FirstName:  input.FirstName,
		LastName:   input.LastName,
		Phone:      input.Phone,
		Department: input.Department,
		Position:   input.Position,
		IsActive:   true,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	roles, err := s.getUserRoles(user.ID)
	if err != nil {
		roles = []string{"user"}
	}

	accessToken, err := token.GenerateTokenWithRoles(user.ID.String(), roles)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateRefreshToken(user.ID.String())
	if err != nil {
		return nil, err
	}

	userDTO := &dto.UserDTO{
		ID:         user.ID,
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
	}

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         userDTO,
		Roles:        roles,
		ExpiresIn:    3600,
		TokenType:    "Bearer",
	}, nil
}

func (s *authService) Login(input dto.UserLoginRequest) (*dto.AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(input.Email)
	if err != nil {
		return nil, exceptions.NewInvalidCredentialsError()
	}

	if !user.IsActive {
		return nil, &exceptions.AccountDisabledError{}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return nil, exceptions.NewInvalidCredentialsError()
	}

	roles, err := s.getUserRoles(user.ID)
	if err != nil {
		roles = []string{"user"}
	}

	accessToken, err := token.GenerateTokenWithRoles(user.ID.String(), roles)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateRefreshToken(user.ID.String())
	if err != nil {
		return nil, err
	}

	now := time.Now()
	user.LastLogin = &now
	s.userRepo.Update(user)

	userDTO := &dto.UserDTO{
		ID:         user.ID,
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
	}

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         userDTO,
		Roles:        roles,
		ExpiresIn:    3600,
		TokenType:    "Bearer",
	}, nil
}

func (s *authService) Logout(userID uuid.UUID) error {
	return nil
}

func (s *authService) RefreshToken(refreshToken string) (*dto.TokenResponse, error) {
	claims, err := s.validateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	userID, _ := uuid.Parse(claims.UserId)
	roles, err := s.getUserRoles(userID)
	if err != nil {
		roles = []string{"user"}
	}

	newAccessToken, err := token.GenerateTokenWithRoles(claims.UserId, roles)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := s.generateRefreshToken(claims.UserId)
	if err != nil {
		return nil, err
	}

	return &dto.TokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    3600,
		TokenType:    "Bearer",
	}, nil
}

func (s *authService) ValidateToken(tokenStr string) (*token.Claims, error) {
	claims, err := token.ValidateToken(tokenStr)
	if err != nil {
		if err.Error() == "token has invalid claims: token is expired" {
			return nil, &exceptions.TokenExpiredError{}
		}
		return nil, &exceptions.InvalidTokenError{
			Message: err.Error(),
		}
	}
	return claims, nil
}

func (s *authService) ChangePassword(userID uuid.UUID, input dto.ChangePasswordRequest) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.OldPassword)); err != nil {
		return exceptions.NewInvalidCredentialsError()
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Password = string(hashed)
	return s.userRepo.Update(user)
}

func (s *authService) getUserRoles(userID uuid.UUID) ([]string, error) {
	userRoles, err := s.roleRepo.GetUserRoles(userID)
	if err != nil {
		return []string{"user"}, nil
	}

	var roles []string
	for _, role := range userRoles {
		roles = append(roles, role.Role.Name)
	}
	return roles, nil
}

func (s *authService) generateRefreshToken(userID string) (string, error) {
	return "", nil
}

func (s *authService) validateRefreshToken(tokenStr string) (*token.Claims, error) {
	return s.ValidateToken(tokenStr)
}
