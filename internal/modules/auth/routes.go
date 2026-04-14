package auth

import (
	httpmiddleware "sipi/internal/http/middleware"
	"sipi/internal/platform/jwt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(group *gin.RouterGroup, db *gorm.DB, tokenManager *jwt.Manager) {
	repository := NewRepository(db)
	service := NewService(repository, NewTokenManagerAdapter(tokenManager))
	handler := NewHandler(service)

	group.POST("/register", handler.Register)
	group.POST("/login", handler.Login)
	group.POST("/refresh", handler.Refresh)
	group.POST("/logout", httpmiddleware.Auth(tokenManager), handler.Logout)
	group.POST("/logout-all", httpmiddleware.Auth(tokenManager), handler.LogoutAll)
}
