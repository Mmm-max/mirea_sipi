package auth

import (
	"errors"
)

type ErrorCode string

var (
	ErrNotFound      = errors.New("resource not found")
	ErrAlreadyExists = errors.New("resource already exists")
)

const (
	ErrCodeValidation          ErrorCode = "validation_error"
	ErrCodeEmailAlreadyExists  ErrorCode = "email_already_exists"
	ErrCodeInvalidCredentials  ErrorCode = "invalid_credentials"
	ErrCodeInvalidRefreshToken ErrorCode = "invalid_refresh_token"
	ErrCodeRefreshTokenExpired ErrorCode = "refresh_token_expired"
	ErrCodeRefreshTokenReused  ErrorCode = "refresh_token_reused"
	ErrCodeForbidden           ErrorCode = "forbidden"
	ErrCodeInternal            ErrorCode = "internal_error"
)

type ServiceError struct {
	Code    ErrorCode
	Message string
	Err     error
}

func (e *ServiceError) Error() string {
	if e == nil {
		return ""
	}

	return e.Message
}

func (e *ServiceError) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Err
}

func (e *ServiceError) ErrorCode() string {
	if e == nil {
		return "internal_error"
	}
	return string(e.Code)
}

func (e *ServiceError) ClientMessage() string {
	if e == nil {
		return "internal server error"
	}
	return e.Message
}

func newServiceError(code ErrorCode, message string, err error) *ServiceError {
	return &ServiceError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}
