package service

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"morent-backend/internal/models"
)

var (
	ErrUserExists         = errors.New("user with this email already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidToken       = errors.New("invalid or expired token")
)

type AuthService struct {
	userRepo    AuthUserRepository
	sessionRepo AuthSessionRepository
}

type AuthUserRepository interface {
	GetByEmail(email string) (*models.User, error)
	Create(user *models.User) error
	GetByID(id uint) (*models.User, error)
	Update(user *models.User) error
}

type AuthSessionRepository interface {
	Create(userID uint, token string) error
	GetByToken(token string) (*models.Session, error)
	DeleteByToken(token string) error
}

func NewAuthService(userRepo AuthUserRepository, sessionRepo AuthSessionRepository) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}
}

func (s *AuthService) Register(name, email, password string) (*models.UserResponse, string, error) {
	if strings.TrimSpace(name) == "" || strings.TrimSpace(email) == "" || strings.TrimSpace(password) == "" {
		return nil, "", errors.New("all fields are required")
	}

	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	existing, err := s.userRepo.GetByEmail(normalizedEmail)
	if err != nil {
		return nil, "", err
	}
	if existing != nil {
		return nil, "", ErrUserExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	user := models.User{
		Name:         strings.TrimSpace(name),
		Email:        normalizedEmail,
		PasswordHash: string(hash),
		AvatarURL:    "https://avatars.mds.yandex.net/i?id=18025267d7d94e6289d82fda9b36eea0_l-5256838-images-thumbs&n=13",
		Nickname:     strings.TrimSpace(name),
		Position:     "",
		Role:		  "user",
	}

	if errCreate := s.userRepo.Create(&user); errCreate != nil {
		return nil, "", errCreate
	}

	resp := user.ToResponse()
	token, errSession := s.createSession(user.ID)
	if errSession != nil {
		return nil, "", errSession
	}
	return &resp, token, nil
}

func (s *AuthService) Login(email, password string) (*models.UserResponse, string, error) {
	if strings.TrimSpace(email) == "" || strings.TrimSpace(password) == "" {
		return nil, "", errors.New("email and password are required")
	}

	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	user, err := s.userRepo.GetByEmail(normalizedEmail)
	if err != nil {
		return nil, "", err
	}
	if user == nil {
		return nil, "", ErrInvalidCredentials
	}

	if errCompare := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); errCompare != nil {
		return nil, "", ErrInvalidCredentials
	}

	resp := user.ToResponse()
	token, errSession := s.createSession(user.ID)
	if errSession != nil {
		return nil, "", errSession
	}
	return &resp, token, nil
}

func (s *AuthService) createSession(userID uint) (string, error) {
	token := uuid.NewString()
	if err := s.sessionRepo.Create(userID, token); err != nil {
		return "", err
	}
	return token, nil
}

func (s *AuthService) GetUserByToken(token string) (*models.User, error) {
	if strings.TrimSpace(token) == "" {
		return nil, ErrInvalidToken
	}
	session, err := s.sessionRepo.GetByToken(token)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, ErrInvalidToken
	}
	return &session.User, nil
}

func (s *AuthService) Logout(token string) error {
	if strings.TrimSpace(token) == "" {
		return ErrInvalidToken
	}
	return s.sessionRepo.DeleteByToken(token)
}

func (s *AuthService) UpdateProfile(userID uint, name, nickname, position, avatarURL *string) (*models.UserResponse, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidToken
	}

	if name != nil {
		user.Name = strings.TrimSpace(*name)
	}
	if nickname != nil {
		user.Nickname = strings.TrimSpace(*nickname)
	}
	if position != nil {
		user.Position = strings.TrimSpace(*position)
	}
	if avatarURL != nil {
		user.AvatarURL = strings.TrimSpace(*avatarURL)
	}

	if errSave := s.userRepo.Update(user); errSave != nil {
		return nil, errSave
	}
	resp := user.ToResponse()
	return &resp, nil
}

func (s *AuthService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	if strings.TrimSpace(newPassword) == "" {
		return errors.New("new password is required")
	}
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrInvalidToken
	}

	if errCompare := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); errCompare != nil {
		return ErrInvalidCredentials
	}

	hash, errHash := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if errHash != nil {
		return errHash
	}
	user.PasswordHash = string(hash)
	return s.userRepo.Update(user)
}
