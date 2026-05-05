package exceptions

import "fmt"

type UserNotFoundError struct {
	UserID string
}

func (e *UserNotFoundError) Error() string {
	return fmt.Sprintf("user with ID %s not found", e.UserID)
}

type InvalidCredentialsError struct{}

func (e *InvalidCredentialsError) Error() string {
	return "invalid credentials"
}

type UserAlreadyExistsError struct {
	Email string
}

func (e *UserAlreadyExistsError) Error() string {
	return fmt.Sprintf("user with email %s already exists", e.Email)
}

type RoleNotFoundError struct {
	RoleName string
}

func (e *RoleNotFoundError) Error() string {
	return fmt.Sprintf("role %s not found", e.RoleName)
}

type PermissionDeniedError struct {
	UserID     string
	Permission string
}

func (e *PermissionDeniedError) Error() string {
	return fmt.Sprintf("user %s does not have permission %s", e.UserID, e.Permission)
}

type CompanyNotFoundError struct {
	CompanyID string
}

func (e *CompanyNotFoundError) Error() string {
	return fmt.Sprintf("company with ID %s not found", e.CompanyID)
}

type CompanyAlreadyExistsError struct {
	Domain string
}

func (e *CompanyAlreadyExistsError) Error() string {
	return fmt.Sprintf("company with domain %s already exists", e.Domain)
}
