package calendarimport

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

// ImportICS godoc
// @Summary Import calendar from ICS file
// @Description Uploads an .ics file, validates iCalendar payload, stores import history and creates or updates imported events
// @Tags calendar-import
// @Security BearerAuth
// @Accept mpfd
// @Produce json
// @Param file formData file true "ICS file"
// @Success 201 {object} ImportJobEnvelope
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /calendar-import/ics [post]
func (h *Handler) ImportICS(c *gin.Context) {
	userID, ok := httpmiddleware.RequireUserID(c)
	if !ok {
		return
	}

	command, err := BindImportICSCommand(c, userID)
	if err != nil {
		httpx.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	job, err := h.service.ImportICS(c.Request.Context(), command)
	if err != nil {
		httpx.HandleServiceError(c, err)
		return
	}

	httpx.Created(c, toImportJobResponse(job))
}

// ListHistory godoc
// @Summary List calendar import history
// @Description Returns import history for the authenticated user
// @Tags calendar-import
// @Security BearerAuth
// @Produce json
// @Success 200 {object} ImportJobListEnvelope
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /calendar-import/history [get]
func (h *Handler) ListHistory(c *gin.Context) {
	userID, ok := httpmiddleware.RequireUserID(c)
	if !ok {
		return
	}

	jobs, err := h.service.ListHistory(c.Request.Context(), ListImportJobsQuery{UserID: userID})
	if err != nil {
		httpx.HandleServiceError(c, err)
		return
	}

	httpx.OK(c, ImportJobListResponse{Items: toImportJobResponses(jobs)})
}

// GetHistoryByID godoc
// @Summary Get calendar import history item
// @Description Returns a specific import history record for the authenticated user
// @Tags calendar-import
// @Security BearerAuth
// @Produce json
// @Param id path string true "Import job ID"
// @Success 200 {object} ImportJobEnvelope
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 404 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /calendar-import/history/{id} [get]
func (h *Handler) GetHistoryByID(c *gin.Context) {
	userID, ok := httpmiddleware.RequireUserID(c)
	if !ok {
		return
	}

	id, ok := httpx.ParseUUIDParam(c, "id", "invalid import job id")
	if !ok {
		return
	}

	job, err := h.service.GetHistoryByID(c.Request.Context(), GetImportJobQuery{
		UserID: userID,
		ID:     id,
	})
	if err != nil {
		httpx.HandleServiceError(c, err)
		return
	}

	httpx.OK(c, toImportJobResponse(job))
}
