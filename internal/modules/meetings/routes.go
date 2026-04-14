package meetings

import (
	httpmiddleware "sipi/internal/http/middleware"
	"sipi/internal/modules/notifications"
	"sipi/internal/modules/resources"
	"sipi/internal/modules/users"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"sipi/internal/platform/jwt"
)

func RegisterRoutes(group *gin.RouterGroup, db *gorm.DB, tokenManager *jwt.Manager) {
	repository := NewRepository(db)
	userDirectory := NewUserDirectory(users.NewRepository(db))
	resourceDirectory := NewResourceDirectory(resources.NewRepository(db))
	notificationService := notifications.NewService(notifications.NewRepository(db))
	notificationHook := NewNotificationHook(notificationService)
	service := NewService(repository, userDirectory, resourceDirectory, notificationHook)
	handler := NewHandler(service)

	group.Use(httpmiddleware.Auth(tokenManager))
	group.POST("", handler.CreateMeeting)
	group.GET("", handler.ListMeetings)
	group.GET("/:id", handler.GetMeeting)
	group.PATCH("/:id", handler.UpdateMeeting)
	group.DELETE("/:id", handler.DeleteMeeting)
	group.POST("/:id/participants", handler.AddParticipants)
	group.DELETE("/:id/participants/:userId", handler.RemoveParticipant)
	group.POST("/:id/resources", handler.AddResources)
	group.DELETE("/:id/resources/:resourceId", handler.RemoveResource)
	group.POST("/:id/respond", handler.RespondInvitation)
	group.POST("/:id/request-alternative", handler.RequestAlternative)
}
