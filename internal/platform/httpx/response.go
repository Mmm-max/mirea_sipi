package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ResponseMeta struct {
	RequestID string `json:"request_id,omitempty"`
}

type SuccessResponse struct {
	Success bool         `json:"success"`
	Data    any          `json:"data,omitempty"`
	Meta    ResponseMeta `json:"meta,omitempty"`
}

type ErrorResponse struct {
	Success bool         `json:"success"`
	Error   ErrorBody    `json:"error"`
	Meta    ResponseMeta `json:"meta,omitempty"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

func Success(c *gin.Context, statusCode int, data any) {
	c.JSON(statusCode, SuccessResponse{
		Success: true,
		Data:    data,
		Meta:    responseMeta(c),
	})
}

func OK(c *gin.Context, payload any) {
	Success(c, http.StatusOK, payload)
}

func Created(c *gin.Context, payload any) {
	Success(c, http.StatusCreated, payload)
}

func Error(c *gin.Context, statusCode int, code, message string) {
	c.JSON(statusCode, ErrorResponse{
		Success: false,
		Error: ErrorBody{
			Code:    code,
			Message: message,
		},
		Meta: responseMeta(c),
	})
}

func responseMeta(c *gin.Context) ResponseMeta {
	requestID, _ := c.Get(ContextRequestIDKey)

	return ResponseMeta{
		RequestID: toString(requestID),
	}
}

func toString(value any) string {
	if value == nil {
		return ""
	}

	stringValue, ok := value.(string)
	if !ok {
		return ""
	}

	return stringValue
}
