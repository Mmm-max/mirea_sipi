package httpx

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type ServiceError interface {
	error
	ErrorCode() string
	ClientMessage() string
}

func HandleServiceError(c *gin.Context, err error) {
	var serviceErr ServiceError
	if !errors.As(err, &serviceErr) {
		Error(c, http.StatusInternalServerError, "internal_error", "internal server error")
		return
	}

	Error(c, mapServiceErrorStatus(serviceErr.ErrorCode()), serviceErr.ErrorCode(), serviceErr.ClientMessage())
}

func mapServiceErrorStatus(code string) int {
	switch strings.TrimSpace(strings.ToLower(code)) {
	case "validation_error":
		return http.StatusBadRequest
	case "not_found":
		return http.StatusNotFound
	case "forbidden":
		return http.StatusForbidden
	case "conflict", "email_already_exists":
		return http.StatusConflict
	case "invalid_credentials", "invalid_refresh_token", "refresh_token_expired", "refresh_token_reused":
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}
