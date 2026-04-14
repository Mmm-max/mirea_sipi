package scheduling

import (
	httpmiddleware "sipi/internal/http/middleware"
	"sipi/internal/modules/events"
	"sipi/internal/modules/meetings"
	"sipi/internal/modules/notifications"
	"sipi/internal/modules/resources"
	"sipi/internal/modules/users"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"sipi/internal/platform/jwt"
)

func RegisterRoutes(group *gin.RouterGroup, db *gorm.DB, tokenManager *jwt.Manager) {
	repository := NewRepository(db)
	meetingStore := NewMeetingStore(meetings.NewRepository(db))
	availability := NewAvailabilityStore(users.NewRepository(db), events.NewRepository(db), resources.NewRepository(db))
	eventSync := NewNoopEventSync()
	notificationService := notifications.NewService(notifications.NewRepository(db))
	notificationHook := NewNotificationHook(notificationService)
	service := NewService(repository, meetingStore, availability, eventSync, notificationHook)
	handler := NewHandler(service)

	group.Use(httpmiddleware.Auth(tokenManager))
	group.POST("/:id/search-slots", handler.SearchSlots)
	group.POST("/:id/select-slot", handler.SelectSlot)
	group.GET("/:id/slots", handler.GetSlots)
}
