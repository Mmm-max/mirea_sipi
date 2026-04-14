package events

import (
	"errors"
)

type ErrorCode string

const (
	ErrCodeValidation ErrorCode = "validation_error"
	ErrCodeNotFound   ErrorCode = "not_found"
	ErrCodeForbidden  ErrorCode = "forbidden"
	ErrCodeConflict   ErrorCode = "conflict"
	ErrCodeInternal   ErrorCode = "internal_error"
)

var ErrNotFound = errors.New("resource not found")

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
