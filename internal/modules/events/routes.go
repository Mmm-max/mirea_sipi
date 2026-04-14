package events

import (
	httpmiddleware "sipi/internal/http/middleware"
	"sipi/internal/platform/jwt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(group *gin.RouterGroup, db *gorm.DB, tokenManager *jwt.Manager) {
	repository := NewRepository(db)
	service := NewService(repository)
	handler := NewHandler(service)

	group.Use(httpmiddleware.Auth(tokenManager))
	group.POST("", handler.Create)
	group.GET("", handler.List)
	group.GET("/:id", handler.GetByID)
	group.PATCH("/:id", handler.Update)
	group.DELETE("/:id", handler.Delete)
}
