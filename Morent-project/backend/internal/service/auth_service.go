package service

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"morent-backend/internal/messaging"
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
	events      messaging.UserEventPublisher
	companyName string
	log         *slog.Logger
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
	GetLatestByUserID(userID uint) (*models.Session, error)
	DeleteByToken(token string) error
}

func NewAuthService(
	userRepo AuthUserRepository,
	sessionRepo AuthSessionRepository,
	events messaging.UserEventPublisher,
	companyName string,
	log *slog.Logger,
) *AuthService {
	if events == nil {
		events = messaging.NoopPublisher{}
	}
	if log == nil {
		log = slog.Default()
	}
	return &AuthService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		events:      events,
		companyName: strings.TrimSpace(companyName),
		log:         log,
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
		AvatarURL:    defaultAvatarURL(),
		Nickname:     strings.TrimSpace(name),
		Position:     "",
		Role:         "user",
	}

	if errCreate := s.userRepo.Create(&user); errCreate != nil {
		return nil, "", errCreate
	}

	s.publishRegistered(user)

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
	s.publishProfileUpdated(user, name, nickname, position, avatarURL, nil)
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
	if err := s.userRepo.Update(user); err != nil {
		return err
	}
	s.publishPasswordChanged(user)
	return nil
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

func defaultAvatarURL() string {
	return "https://avatars.mds.yandex.net/i?id=18025267d7d94e6289d82fda9b36eea0_l-5256838-images-thumbs&n=13"
}

func (s *AuthService) publishRegistered(user models.User) {
	if !s.events.Enabled() {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := s.events.PublishUserRegistered(ctx, messaging.UserRegisteredEvent{
			MorentUserID: user.ID,
			Email:        user.Email,
			Name:         user.Name,
			Nickname:     user.Nickname,
			Position:     user.Position,
			AvatarURL:    user.AvatarURL,
			Role:         user.Role,
			PasswordHash: user.PasswordHash,
			CompanyName:  s.companyName,
		})
		if err != nil {
			s.log.Warn("failed to publish user.registered", "user_id", user.ID, "error", err)
		}
	}()
}

func (s *AuthService) publishProfileUpdated(
	user *models.User,
	name, nickname, position, avatarURL *string,
	isActive *bool,
) {
	if !s.events.Enabled() {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := s.events.PublishUserProfileUpdated(ctx, messaging.UserProfileUpdatedEvent{
			MorentUserID: user.ID,
			Email:        user.Email,
			Name:         name,
			Nickname:     nickname,
			Position:     position,
			AvatarURL:    avatarURL,
			Role:         nil,
			IsActive:     isActive,
		})
		if err != nil {
			s.log.Warn("failed to publish user.profile_updated", "user_id", user.ID, "error", err)
		}
	}()
}

func (s *AuthService) publishPasswordChanged(user *models.User) {
	if !s.events.Enabled() {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := s.events.PublishUserPasswordChanged(ctx, messaging.UserPasswordChangedEvent{
			MorentUserID: user.ID,
			Email:        user.Email,
			PasswordHash: user.PasswordHash,
		})
		if err != nil {
			s.log.Warn("failed to publish user.password_changed", "user_id", user.ID, "error", err)
		}
	}()
}
