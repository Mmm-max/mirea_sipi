package events

import (
	"net/http"
	"time"

	httpmiddleware "sipi/internal/http/middleware"
	"sipi/internal/platform/httpx"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service ServicePort
}

func NewHandler(service ServicePort) *Handler {
	return &Handler{service: service}
}

// Create godoc
// @Summary Create event
// @Description Creates a manual event for the authenticated user
// @Tags events
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CreateEventRequest true "Create event request"
// @Success 201 {object} EventEnvelope
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 409 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /events [post]
func (h *Handler) Create(c *gin.Context) {
	userID, ok := httpmiddleware.RequireUserID(c)
	if !ok {
		return
	}

	var request CreateEventRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpx.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	command, err := request.ToCommand(userID)
	if err != nil {
		httpx.Error(c, http.StatusBadRequest, "validation_error", "invalid date format, expected RFC3339")
		return
	}

	event, err := h.service.Create(c.Request.Context(), command)
	if err != nil {
		httpx.HandleServiceError(c, err)
		return
	}

	httpx.Created(c, toEventResponse(event))
}

// List godoc
// @Summary List events
// @Description Returns events of the authenticated user with optional date range filtering
// @Tags events
// @Security BearerAuth
// @Produce json
// @Param date_from query string false "RFC3339 lower bound"
// @Param date_to query string false "RFC3339 upper bound"
// @Success 200 {object} EventListEnvelope
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /events [get]
func (h *Handler) List(c *gin.Context) {
	userID, ok := httpmiddleware.RequireUserID(c)
	if !ok {
		return
	}

	query, err := buildListQuery(c, userID)
	if err != nil {
		httpx.Error(c, http.StatusBadRequest, "validation_error", "invalid date range format, expected RFC3339")
		return
	}

	events, err := h.service.List(c.Request.Context(), query)
	if err != nil {
		httpx.HandleServiceError(c, err)
		return
	}

	httpx.OK(c, EventListResponse{Items: toEventResponses(events)})
}

// GetByID godoc
// @Summary Get event by id
// @Description Returns an event by id for the authenticated user
// @Tags events
// @Security BearerAuth
// @Produce json
// @Param id path string true "Event ID"
// @Success 200 {object} EventEnvelope
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 404 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /events/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	userID, ok := httpmiddleware.RequireUserID(c)
	if !ok {
		return
	}

	id, ok := httpx.ParseUUIDParam(c, "id", "invalid event id")
	if !ok {
		return
	}

	event, err := h.service.Get(c.Request.Context(), GetEventQuery{
		ID:          id,
		OwnerUserID: userID,
	})
	if err != nil {
		httpx.HandleServiceError(c, err)
		return
	}

	httpx.OK(c, toEventResponse(event))
}

// Update godoc
// @Summary Update event
// @Description Updates an event of the authenticated user
// @Tags events
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Event ID"
// @Param request body UpdateEventRequest true "Update event request"
// @Success 200 {object} EventEnvelope
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 404 {object} httpx.ErrorResponse
// @Failure 409 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /events/{id} [patch]
func (h *Handler) Update(c *gin.Context) {
	userID, ok := httpmiddleware.RequireUserID(c)
	if !ok {
		return
	}

	id, ok := httpx.ParseUUIDParam(c, "id", "invalid event id")
	if !ok {
		return
	}

	var request UpdateEventRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpx.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	command, err := request.ToCommand(id, userID)
	if err != nil {
		httpx.Error(c, http.StatusBadRequest, "validation_error", "invalid date format, expected RFC3339")
		return
	}

	event, err := h.service.Update(c.Request.Context(), command)
	if err != nil {
		httpx.HandleServiceError(c, err)
		return
	}

	httpx.OK(c, toEventResponse(event))
}

// Delete godoc
// @Summary Delete event
// @Description Deletes an event of the authenticated user
// @Tags events
// @Security BearerAuth
// @Produce json
// @Param id path string true "Event ID"
// @Success 200 {object} MessageEnvelope
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 404 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /events/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	userID, ok := httpmiddleware.RequireUserID(c)
	if !ok {
		return
	}

	id, ok := httpx.ParseUUIDParam(c, "id", "invalid event id")
	if !ok {
		return
	}

	if err := h.service.Delete(c.Request.Context(), DeleteEventCommand{
		ID:          id,
		OwnerUserID: userID,
	}); err != nil {
		httpx.HandleServiceError(c, err)
		return
	}

	httpx.OK(c, MessageData{Message: "event deleted successfully"})
}

func buildListQuery(c *gin.Context, ownerUserID uuid.UUID) (ListEventsQuery, error) {
	query := ListEventsQuery{OwnerUserID: ownerUserID}

	if value := c.Query("date_from"); value != "" {
		parsed, err := time.Parse(time.RFC3339, value)
		if err != nil {
			return ListEventsQuery{}, err
		}
		query.DateFrom = &parsed
	}

	if value := c.Query("date_to"); value != "" {
		parsed, err := time.Parse(time.RFC3339, value)
		if err != nil {
			return ListEventsQuery{}, err
		}
		query.DateTo = &parsed
	}

	return query, nil
}

func toEventResponse(event *Event) EventResponse {
	return EventResponse{
		ID:              event.ID.String(),
		OwnerUserID:     event.OwnerUserID.String(),
		SourceType:      string(event.SourceType),
		ExternalUID:     event.ExternalUID,
		Title:           event.Title,
		Description:     event.Description,
		StartAt:         event.StartAt.UTC().Format(time.RFC3339),
		EndAt:           event.EndAt.UTC().Format(time.RFC3339),
		Priority:        string(event.Priority),
		IsReschedulable: event.IsReschedulable,
		VisibilityHint:  event.VisibilityHint,
		CreatedAt:       event.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:       event.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func toEventResponses(events []Event) []EventResponse {
	items := make([]EventResponse, 0, len(events))
	for i := range events {
		event := events[i]
		items = append(items, toEventResponse(&event))
	}

	return items
}
