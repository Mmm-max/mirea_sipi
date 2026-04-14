package routes

import (
	"sipi/internal/config"
	"sipi/internal/http/middleware"
	"sipi/internal/modules/auth"
	"sipi/internal/modules/calendarimport"
	"sipi/internal/modules/events"
	"sipi/internal/modules/meetings"
	"sipi/internal/modules/notifications"
	"sipi/internal/modules/resources"
	"sipi/internal/modules/scheduling"
	"sipi/internal/modules/users"
	"sipi/internal/platform/jwt"
	"sipi/internal/platform/logger"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Dependencies struct {
	Config       *config.Config
	DB           *gorm.DB
	Logger       *logger.Logger
	TokenManager *jwt.Manager
}

func Setup(deps Dependencies) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(middleware.RequestID())
	router.Use(middleware.RequestLogger(deps.Logger))
	router.Use(middleware.Recovery(deps.Logger))

	registerHealthRoutes(router)
	registerSwaggerRoutes(router)
	auth.RegisterRoutes(router.Group("/auth"), deps.DB, deps.TokenManager)
	users.RegisterRoutes(router.Group("/users"), deps.DB, deps.TokenManager)
	events.RegisterRoutes(router.Group("/events"), deps.DB, deps.TokenManager)
	calendarimport.RegisterRoutes(router.Group("/calendar-import"), deps.DB, deps.TokenManager)
	meetings.RegisterRoutes(router.Group("/meetings"), deps.DB, deps.TokenManager)
	scheduling.RegisterRoutes(router.Group("/meetings"), deps.DB, deps.TokenManager)
	resources.RegisterRoutes(router.Group("/resources"), deps.DB, deps.TokenManager)
	notifications.RegisterRoutes(router.Group("/notifications"), deps.DB, deps.TokenManager)

	return router
}

func registerSwaggerRoutes(router *gin.Engine) {
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
}
