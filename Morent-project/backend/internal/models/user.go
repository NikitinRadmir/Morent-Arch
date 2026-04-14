package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name         string `gorm:"size:100;not null" json:"name"`
	Email        string `gorm:"size:150;not null;uniqueIndex" json:"email"`
	PasswordHash string `gorm:"not null" json:"-"`
	AvatarURL    string `gorm:"size:255" json:"avatarUrl"`
	Nickname     string `gorm:"size:100" json:"nickname"`
	Position     string `gorm:"size:120" json:"position"`
	Role         string `gorm:"size:120" json:"role"`
}

type UserResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	AvatarURL string    `json:"avatarUrl"`
	Nickname  string    `json:"nickname"`
	Position  string    `json:"position"`
	Role 	  string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
}

func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		AvatarURL: u.AvatarURL,
		Nickname:  u.Nickname,
		Position:  u.Position,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
	}
}
