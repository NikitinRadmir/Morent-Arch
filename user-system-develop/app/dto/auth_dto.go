// dto/auth.go
package dto

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type RegisterRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	FirstName   string `json:"first_name" binding:"required"`
	LastName    string `json:"last_name" binding:"required"`
	CompanyName string `json:"company_name" binding:"required"`
	Phone       string `json:"phone"`
	Department  string `json:"department"`
	Position    string `json:"position"`
}

func (r *RegisterRequest) Validate() error {
	// Можно добавить дополнительные проверки
	return nil
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (l *LoginRequest) Validate() error {
	return nil
}

func (c *ChangePasswordRequest) Validate() error {
	if c.OldPassword == c.NewPassword {
		return errors.New("new password must be different from old password")
	}
	return nil
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type UserDTO struct {
	ID         uuid.UUID  `json:"id"`
	Email      string     `json:"email"`
	FirstName  string     `json:"first_name"`
	LastName   string     `json:"last_name"`
	Phone      string     `json:"phone,omitempty"`
	Department string     `json:"department,omitempty"`
	Position   string     `json:"position,omitempty"`
	IsActive   bool       `json:"is_active"`
	LastLogin  *time.Time `json:"last_login,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type AuthResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token,omitempty"`
	User         *UserDTO `json:"user"`
	Roles        []string `json:"roles"`
	ExpiresIn    int64    `json:"expires_in"`
	TokenType    string   `json:"token_type"`
}
