package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateUserRequest struct {
	CompanyID  uuid.UUID `json:"company_id" binding:"required"`
	Email      string    `json:"email" binding:"required,email"`
	Password   string    `json:"password" binding:"required,min=8"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Phone      string    `json:"phone"`
	Department string    `json:"department"`
	Position   string    `json:"position"`
}

type UpdateUserRequest struct {
	Email      string `json:"email" binding:"required,email"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Phone      string `json:"phone"`
	Department string `json:"department"`
	Position   string `json:"position"`
	IsActive   *bool  `json:"is_active"`
}

type UserResponse struct {
	ID         uuid.UUID  `json:"id"`
	CompanyID  uuid.UUID  `json:"company_id"`
	Email      string     `json:"email"`
	FirstName  string     `json:"first_name"`
	LastName   string     `json:"last_name"`
	Phone      string     `json:"phone"`
	Department string     `json:"department"`
	Position   string     `json:"position"`
	IsActive   bool       `json:"is_active"`
	LastLogin  *time.Time `json:"last_login,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type UserWithCompanyResponse struct {
	UserResponse
	CompanyName string `json:"company_name"`
}

type UserLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}
