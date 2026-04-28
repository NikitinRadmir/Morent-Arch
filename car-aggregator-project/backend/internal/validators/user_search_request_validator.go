package validators

import (
	"context"

	"car-aggregator/internal/dtos"
	"car-aggregator/internal/errors"
)

type UserSearchRequestValidator interface {
	ValidateSearch(ctx context.Context, in dtos.UserSearchQuery) error
}

type DefaultUserSearchRequestValidator struct {
}

func NewUserSearchRequestValidator() *DefaultUserSearchRequestValidator {
	return &DefaultUserSearchRequestValidator{}
}

func (v *DefaultUserSearchRequestValidator) ValidateSearch(ctx context.Context, in dtos.UserSearchQuery) error {
	query := in.Query

	// TODO: Add more validation rules
	if query == "" {
		return errors.New(errors.CodeValidation, "query empty")
	}
	return nil
}
