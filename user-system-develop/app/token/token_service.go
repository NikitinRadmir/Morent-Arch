package token

import (
	"user-system/app/models"
)

type TokenService struct {
	secret string
}

func NewTokenService(secret string) *TokenService {
	return &TokenService{
		secret: secret,
	}
}

func (s *TokenService) GenerateToken(user *models.User) (string, error) {
	roles := []string{}

	return GenerateTokenWithRoles(user.ID.String(), roles)
}

func (s *TokenService) ValidateToken(tokenString string) (*Claims, error) {
	return ValidateToken(tokenString)
}
