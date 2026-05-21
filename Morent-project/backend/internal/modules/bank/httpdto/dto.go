package httpdto

import "morent-backend/internal/bank"

type RegisterRequest struct {
	Phone       string `json:"phone" validate:"required"`
	Password    string `json:"password" validate:"required,min=6"`
	DisplayName string `json:"displayName"`
}

type LoginRequest struct {
	Phone    string `json:"phone" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	Token   string              `json:"token"`
	Profile bank.ProfileResponse `json:"profile"`
}

type AmountRequest struct {
	Amount float64 `json:"amount" validate:"required"`
}

type TransferRequest struct {
	RecipientCardNumber string  `json:"recipientCardNumber" validate:"required"`
	Amount              float64 `json:"amount" validate:"required"`
}

type OperationResponse struct {
	Message string              `json:"message"`
	Profile bank.ProfileResponse `json:"profile"`
}
