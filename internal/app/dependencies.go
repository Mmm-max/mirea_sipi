package app

import (
	"sipi/internal/config"
	"sipi/internal/platform/jwt"
	"sipi/internal/platform/logger"

	"gorm.io/gorm"
)

type Dependencies struct {
	Config       *config.Config
	DB           *gorm.DB
	Logger       *logger.Logger
	TokenManager *jwt.Manager
}
