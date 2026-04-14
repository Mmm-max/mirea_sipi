package routes

import (
	"net/http"

	"sipi/internal/platform/httpx"

	"github.com/gin-gonic/gin"
)

type HealthPayload struct {
	Status string `json:"status" example:"ok"`
}

func registerHealthRoutes(router *gin.Engine) {
	router.GET("/health", healthHandler)
}

// healthHandler godoc
// @Summary Health check
// @Description Returns service health status
// @Tags system
// @Produce json
// @Success 200 {object} httpx.SuccessResponse
// @Router /health [get]
func healthHandler(c *gin.Context) {
	httpx.Success(c, http.StatusOK, HealthPayload{
		Status: "ok",
	})
}
