package privacy

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"sipi/internal/platform/jwt"
)

func RegisterRoutes(api *gin.RouterGroup, db *gorm.DB, tokenManager *jwt.Manager) {
	_ = api
	_ = db
	_ = tokenManager
}
