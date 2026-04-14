package resources

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
	group.POST("", handler.CreateResource)
	group.GET("", handler.ListResources)
	group.GET("/:id", handler.GetResource)
	group.PATCH("/:id", handler.UpdateResource)
	group.DELETE("/:id", handler.DeleteResource)
	group.GET("/:id/availability", handler.GetAvailability)
	group.POST("/:id/bookings", handler.CreateBooking)
	group.DELETE("/:id/bookings/:bookingId", handler.CancelBooking)
}
