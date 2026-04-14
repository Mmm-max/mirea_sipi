package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"sipi/internal/platform/httpx"
	"sipi/internal/platform/jwt"
	"sipi/internal/platform/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const contextUserIDKey = "userID"

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.NewString()
		}

		c.Set(httpx.ContextRequestIDKey, requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)
		c.Next()
	}
}

func RequestLogger(log logger.FieldLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()
		c.Next()

		log.Info(
			"http request",
			"method", c.Request.Method,
			"path", c.FullPath(),
			"raw_path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency", time.Since(startedAt).String(),
			"client_ip", c.ClientIP(),
			"request_id", RequestIDFromContext(c),
		)
	}
}

func Recovery(log logger.FieldLogger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		log.Error(
			"panic recovered",
			"error", fmt.Sprint(recovered),
			"request_id", RequestIDFromContext(c),
			"path", c.Request.URL.Path,
			"method", c.Request.Method,
		)
		httpx.Error(c, http.StatusInternalServerError, "internal_error", "internal server error")
		c.Abort()
	})
}

func Auth(tokenManager *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			httpx.Error(c, http.StatusUnauthorized, "unauthorized", "authorization header is required")
			c.Abort()
			return
		}

		parts := strings.Fields(header)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			httpx.Error(c, http.StatusUnauthorized, "unauthorized", "bearer token is required")
			c.Abort()
			return
		}

		token := strings.TrimSpace(parts[1])
		claims, err := tokenManager.Parse(token)
		if err != nil || claims.Type != "access" {
			httpx.Error(c, http.StatusUnauthorized, "unauthorized", "invalid access token")
			c.Abort()
			return
		}

		c.Set(contextUserIDKey, claims.UserID.String())
		c.Next()
	}
}

func RequestIDFromContext(c *gin.Context) string {
	requestID, _ := c.Get(httpx.ContextRequestIDKey)
	value, ok := requestID.(string)
	if !ok {
		return ""
	}

	return value
}

func UserIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	rawUserID, exists := c.Get(contextUserIDKey)
	if !exists {
		return uuid.Nil, false
	}

	userID, ok := rawUserID.(string)
	if !ok {
		return uuid.Nil, false
	}

	parsed, err := uuid.Parse(userID)
	if err != nil {
		return uuid.Nil, false
	}

	return parsed, true
}

func RequireUserID(c *gin.Context) (uuid.UUID, bool) {
	userID, ok := UserIDFromContext(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, "unauthorized", "invalid access token")
		return uuid.Nil, false
	}

	return userID, true
}
