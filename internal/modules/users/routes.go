package users

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
	group.GET("/me", handler.GetMe)
	group.PATCH("/me", handler.UpdateMe)
	group.PUT("/me/working-hours", handler.ReplaceWorkingHours)
	group.GET("/me/working-hours", handler.ListWorkingHours)
	group.POST("/me/unavailability", handler.CreateUnavailability)
	group.GET("/me/unavailability", handler.ListUnavailability)
	group.DELETE("/me/unavailability/:id", handler.DeleteUnavailability)
}
