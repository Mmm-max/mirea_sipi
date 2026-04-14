package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func ParseUUIDParam(c *gin.Context, paramName, errorMessage string) (uuid.UUID, bool) {
	value, err := uuid.Parse(c.Param(paramName))
	if err != nil {
		Error(c, http.StatusBadRequest, "validation_error", errorMessage)
		return uuid.Nil, false
	}

	return value, true
}
