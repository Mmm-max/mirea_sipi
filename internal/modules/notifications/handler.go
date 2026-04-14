package notifications

import (
	"net/http"

	httpmiddleware "sipi/internal/http/middleware"
	"sipi/internal/platform/httpx"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service ServicePort
}

func NewHandler(service ServicePort) *Handler {
	return &Handler{service: service}
}

// ListNotifications godoc
// @Summary List my notifications
// @Description Returns notifications for the current authenticated user
// @Tags notifications
// @Security BearerAuth
// @Produce json
// @Success 200 {object} NotificationsListEnvelope
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /notifications [get]
func (h *Handler) ListNotifications(c *gin.Context) {
	userID, ok := httpmiddleware.RequireUserID(c)
	if !ok {
		return
	}
	items, err := h.service.ListNotifications(c.Request.Context(), ListNotificationsQuery{UserID: userID})
	if err != nil {
		httpx.HandleServiceError(c, err)
		return
	}
	httpx.OK(c, ToListResponse(items))
}

// MarkRead godoc
// @Summary Mark notification as read
// @Description Marks one notification as read for the current authenticated user
// @Tags notifications
// @Security BearerAuth
// @Produce json
// @Param id path string true "Notification ID"
// @Success 200 {object} MessageEnvelope
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 404 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /notifications/{id}/read [post]
func (h *Handler) MarkRead(c *gin.Context) {
	userID, ok := httpmiddleware.RequireUserID(c)
	if !ok {
		return
	}
	command, err := NewMarkReadCommand(userID, c.Param("id"))
	if err != nil {
		httpx.Error(c, http.StatusBadRequest, "validation_error", "invalid notification id")
		return
	}
	if err := h.service.MarkRead(c.Request.Context(), command); err != nil {
		httpx.HandleServiceError(c, err)
		return
	}
	httpx.OK(c, MessageData{Message: "notification marked as read"})
}

// MarkAllRead godoc
// @Summary Mark all notifications as read
// @Description Marks all notifications as read for the current authenticated user
// @Tags notifications
// @Security BearerAuth
// @Produce json
// @Success 200 {object} ReadAllEnvelope
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /notifications/read-all [post]
func (h *Handler) MarkAllRead(c *gin.Context) {
	userID, ok := httpmiddleware.RequireUserID(c)
	if !ok {
		return
	}
	updated, err := h.service.MarkAllRead(c.Request.Context(), MarkAllReadCommand{UserID: userID})
	if err != nil {
		httpx.HandleServiceError(c, err)
		return
	}
	httpx.OK(c, ReadAllResponse{Updated: updated})
}
