package users

import (
	"net/http"
	"time"

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

// GetMe godoc
// @Summary Get current user profile
// @Description Returns the profile of the authenticated user
// @Tags users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} ProfileEnvelope
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 404 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /users/me [get]
func (h *Handler) GetMe(c *gin.Context) {
	userID, ok := httpmiddleware.RequireUserID(c)
	if !ok {
		return
	}

	result, err := h.service.GetProfile(c.Request.Context(), GetProfileQuery{UserID: userID})
	if err != nil {
		httpx.HandleServiceError(c, err)
		return
	}

	httpx.OK(c, newProfileResponse(result))
}

// UpdateMe godoc
// @Summary Update current user profile
// @Description Updates profile fields of the authenticated user
// @Tags users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body UpdateProfileRequest true "Update profile request"
// @Success 200 {object} ProfileEnvelope
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 404 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /users/me [patch]
func (h *Handler) UpdateMe(c *gin.Context) {
	userID, ok := httpmiddleware.RequireUserID(c)
	if !ok {
		return
	}

	var request UpdateProfileRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpx.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	result, err := h.service.UpdateProfile(c.Request.Context(), request.ToCommand(userID))
	if err != nil {
		httpx.HandleServiceError(c, err)
		return
	}

	httpx.OK(c, newProfileResponse(result))
}

// ReplaceWorkingHours godoc
// @Summary Replace working hours
// @Description Replaces all working hours for the authenticated user
// @Tags users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body ReplaceWorkingHoursRequest true "Working hours request"
// @Success 200 {object} WorkingHoursEnvelope
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /users/me/working-hours [put]
func (h *Handler) ReplaceWorkingHours(c *gin.Context) {
	userID, ok := httpmiddleware.RequireUserID(c)
	if !ok {
		return
	}

	var request ReplaceWorkingHoursRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpx.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	result, err := h.service.ReplaceWorkingHours(c.Request.Context(), request.ToCommand(userID))
	if err != nil {
		httpx.HandleServiceError(c, err)
		return
	}

	httpx.OK(c, WorkingHoursListResponse{Items: newWorkingHoursResponses(result)})
}

// ListWorkingHours godoc
// @Summary Get working hours
// @Description Returns working hours for the authenticated user
// @Tags users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} WorkingHoursEnvelope
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /users/me/working-hours [get]
func (h *Handler) ListWorkingHours(c *gin.Context) {
	userID, ok := httpmiddleware.RequireUserID(c)
	if !ok {
		return
	}

	result, err := h.service.ListWorkingHours(c.Request.Context(), ListWorkingHoursQuery{UserID: userID})
	if err != nil {
		httpx.HandleServiceError(c, err)
		return
	}

	httpx.OK(c, WorkingHoursListResponse{Items: newWorkingHoursResponses(result)})
}

// CreateUnavailability godoc
// @Summary Create unavailability period
// @Description Creates a new unavailability period for the authenticated user
// @Tags users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CreateUnavailabilityRequest true "Create unavailability request"
// @Success 201 {object} UnavailabilityEnvelope
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 409 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /users/me/unavailability [post]
func (h *Handler) CreateUnavailability(c *gin.Context) {
	userID, ok := httpmiddleware.RequireUserID(c)
	if !ok {
		return
	}

	var request CreateUnavailabilityRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpx.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	command, err := request.ToCommand(userID)
	if err != nil {
		httpx.Error(c, http.StatusBadRequest, "validation_error", "invalid start_at format, expected RFC3339")
		return
	}
	result, err := h.service.CreateUnavailabilityPeriod(c.Request.Context(), command)
	if err != nil {
		httpx.HandleServiceError(c, err)
		return
	}

	httpx.Created(c, newUnavailabilityResponse(result))
}

// ListUnavailability godoc
// @Summary List unavailability periods
// @Description Returns unavailability periods for the authenticated user
// @Tags users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} UnavailabilityListEnvelope
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /users/me/unavailability [get]
func (h *Handler) ListUnavailability(c *gin.Context) {
	userID, ok := httpmiddleware.RequireUserID(c)
	if !ok {
		return
	}

	result, err := h.service.ListUnavailabilityPeriods(c.Request.Context(), ListUnavailabilityQuery{UserID: userID})
	if err != nil {
		httpx.HandleServiceError(c, err)
		return
	}

	httpx.OK(c, UnavailabilityListResponse{Items: newUnavailabilityResponses(result)})
}

// DeleteUnavailability godoc
// @Summary Delete unavailability period
// @Description Deletes an unavailability period of the authenticated user
// @Tags users
// @Security BearerAuth
// @Produce json
// @Param id path string true "Unavailability ID"
// @Success 200 {object} MessageEnvelope
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 403 {object} httpx.ErrorResponse
// @Failure 404 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /users/me/unavailability/{id} [delete]
func (h *Handler) DeleteUnavailability(c *gin.Context) {
	userID, ok := httpmiddleware.RequireUserID(c)
	if !ok {
		return
	}

	id, ok := httpx.ParseUUIDParam(c, "id", "invalid unavailability id")
	if !ok {
		return
	}

	if err := h.service.DeleteUnavailabilityPeriod(c.Request.Context(), DeleteUnavailabilityCommand{
		UserID: userID,
		ID:     id,
	}); err != nil {
		httpx.HandleServiceError(c, err)
		return
	}

	httpx.OK(c, MessageData{Message: "unavailability period deleted successfully"})
}

func newProfileResponse(result *ProfileResult) ProfileResponse {
	return ProfileResponse{
		ID:        result.ID.String(),
		Email:     result.Email,
		FullName:  result.FullName,
		Timezone:  result.Timezone,
		CreatedAt: result.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: result.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func newWorkingHoursResponses(items []WorkingHours) []WorkingHoursResponse {
	response := make([]WorkingHoursResponse, 0, len(items))
	for _, item := range items {
		response = append(response, WorkingHoursResponse{
			ID:           item.ID.String(),
			Weekday:      item.Weekday,
			StartTime:    item.StartTime,
			EndTime:      item.EndTime,
			IsWorkingDay: item.IsWorkingDay,
		})
	}

	return response
}

func newUnavailabilityResponse(item *UnavailabilityPeriod) UnavailabilityResponse {
	return UnavailabilityResponse{
		ID:        item.ID.String(),
		Type:      string(item.Type),
		Title:     item.Title,
		StartAt:   item.StartAt.UTC().Format(time.RFC3339),
		EndAt:     item.EndAt.UTC().Format(time.RFC3339),
		Comment:   item.Comment,
		CreatedAt: item.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: item.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func newUnavailabilityResponses(items []UnavailabilityPeriod) []UnavailabilityResponse {
	response := make([]UnavailabilityResponse, 0, len(items))
	for _, item := range items {
		copied := item
		response = append(response, newUnavailabilityResponse(&copied))
	}

	return response
}
