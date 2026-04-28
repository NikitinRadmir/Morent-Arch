package errors

import stderrs "errors"

type Code string

const (
	CodeValidation Code = "VALIDATION_FAILED"
	CodeNotFound   Code = "NOT_FOUND"
	CodeInternal   Code = "INTERNAL_ERROR"
)

type DomainError struct {
	Code    Code
	Message string
}

func (e DomainError) Error() string {
	return string(e.Code) + ": " + e.Message
}

func New(code Code, msg string) error {
	return &DomainError{Code: code, Message: msg}
}

func IsDomain(err error) (DomainError, bool) {
	if err == nil {
		return DomainError{}, false
	}
	var de *DomainError
	if stderrs.As(err, &de) {
		return *de, true
	}
	return DomainError{}, false
}
