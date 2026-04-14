package notifications

import (
	httpmiddleware "sipi/internal/http/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"sipi/internal/platform/jwt"
)

func RegisterRoutes(group *gin.RouterGroup, db *gorm.DB, tokenManager *jwt.Manager) {
	repository := NewRepository(db)
	service := NewService(repository)
	handler := NewHandler(service)

	group.Use(httpmiddleware.Auth(tokenManager))
	group.GET("", handler.ListNotifications)
	group.POST("/read-all", handler.MarkAllRead)
	group.POST("/:id/read", handler.MarkRead)
}
