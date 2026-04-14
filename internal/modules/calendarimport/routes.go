package calendarimport

import (
	httpmiddleware "sipi/internal/http/middleware"
	"sipi/internal/modules/events"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"sipi/internal/platform/jwt"
)

func RegisterRoutes(group *gin.RouterGroup, db *gorm.DB, tokenManager *jwt.Manager) {
	repository := NewRepository(db)
	eventStore := NewEventStore(events.NewRepository(db))
	parser := NewICSParser()
	service := NewService(repository, parser, eventStore)
	handler := NewHandler(service)

	group.Use(httpmiddleware.Auth(tokenManager))
	group.POST("/ics", handler.ImportICS)
	group.GET("/history", handler.ListHistory)
	group.GET("/history/:id", handler.GetHistoryByID)
}
