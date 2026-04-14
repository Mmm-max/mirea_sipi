package httpx

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AppError struct {
	Code       string
	Message    string
	StatusCode int
	Err        error
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}

	return e.Message
}

func (e *AppError) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Err
}

func NewAppError(statusCode int, code, message string, err error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
	}
}

func HandleError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		Error(c, appErr.StatusCode, appErr.Code, appErr.Message)
		return
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		Error(c, http.StatusNotFound, "not_found", "resource not found")
		return
	}

	Error(c, http.StatusInternalServerError, "internal_error", "internal server error")
}
