package exceptions

type AuthErrorCode string

const (
	ErrInvalidCredentials AuthErrorCode = "INVALID_CREDENTIALS"
	ErrUserAlreadyExists  AuthErrorCode = "USER_ALREADY_EXISTS"
	ErrCompanyNotFound    AuthErrorCode = "COMPANY_NOT_FOUND"
	ErrAccountDisabled    AuthErrorCode = "ACCOUNT_DISABLED"
	ErrInvalidToken       AuthErrorCode = "INVALID_TOKEN"
	ErrTokenExpired       AuthErrorCode = "TOKEN_EXPIRED"
)

type AuthError struct {
	Code    AuthErrorCode
	Message string
}

func (e *AuthError) Error() string {
	return e.Message
}

func NewAuthError(code AuthErrorCode, message string) *AuthError {
	return &AuthError{
		Code:    code,
		Message: message,
	}
}

func NewInvalidCredentialsError() *AuthError {
	return &AuthError{
		Code:    ErrInvalidCredentials,
		Message: "invalid credentials",
	}
}

func NewUserAlreadyExistsError(email string) *AuthError {
	return &AuthError{
		Code:    ErrUserAlreadyExists,
		Message: "user with email " + email + " already exists",
	}
}

type AccountDisabledError struct{}

func (e *AccountDisabledError) Error() string {
	return "account is disabled"
}

type InvalidTokenError struct {
	Message string
}

func (e *InvalidTokenError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "invalid token"
}

type TokenExpiredError struct{}

func (e *TokenExpiredError) Error() string {
	return "token has expired"
}
