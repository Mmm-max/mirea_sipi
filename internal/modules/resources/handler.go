package resources

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

// CreateResource godoc
// @Summary Create resource
// @Description Creates a new resource
// @Tags resources
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CreateResourceRequest true "Create resource request"
// @Success 201 {object} ResourceEnvelope
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /resources [post]
func (h *Handler) CreateResource(c *gin.Context) {
	userID, ok := httpmiddleware.RequireUserID(c)
	if !ok {
		return
	}

	var request CreateResourceRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpx.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	resource, err := h.service.CreateResource(c.Request.Context(), request.ToCommand(userID))
	if err != nil {
		httpx.HandleServiceError(c, err)
		return
	}

	httpx.Created(c, toResourceResponse(resource))
}

// ListResources godoc
// @Summary List resources
// @Description Returns all resources
// @Tags resources
// @Security BearerAuth
// @Produce json
// @Success 200 {object} ResourceListEnvelope
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /resources [get]
func (h *Handler) ListResources(c *gin.Context) {
	if _, ok := httpmiddleware.RequireUserID(c); !ok {
		return
	}

	resources, err := h.service.ListResources(c.Request.Context(), ListResourcesQuery{})
	if err != nil {
		httpx.HandleServiceError(c, err)
		return
	}

	httpx.OK(c, ResourceListResponse{Items: toResourceResponses(resources)})
}

// GetResource godoc
// @Summary Get resource by id
// @Description Returns a resource by id
// @Tags resources
// @Security BearerAuth
// @Produce json
// @Param id path string true "Resource ID"
// @Success 200 {object} ResourceEnvelope
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 404 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /resources/{id} [get]
func (h *Handler) GetResource(c *gin.Context) {
	if _, ok := httpmiddleware.RequireUserID(c); !ok {
		return
	}

	id, ok := httpx.ParseUUIDParam(c, "id", "invalid resource id")
	if !ok {
		return
	}

	resource, err := h.service.GetResource(c.Request.Context(), GetResourceQuery{ID: id})
	if err != nil {
		httpx.HandleServiceError(c, err)
		return
	}

	httpx.OK(c, toResourceResponse(resource))
}

// UpdateResource godoc
// @Summary Update resource
// @Description Updates a resource
// @Tags resources
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Resource ID"
// @Param request body UpdateResourceRequest true "Update resource request"
// @Success 200 {object} ResourceEnvelope
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 404 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /resources/{id} [patch]
func (h *Handler) UpdateResource(c *gin.Context) {
	if _, ok := httpmiddleware.RequireUserID(c); !ok {
		return
	}

	id, ok := httpx.ParseUUIDParam(c, "id", "invalid resource id")
	if !ok {
		return
	}

	var request UpdateResourceRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpx.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	resource, err := h.service.UpdateResource(c.Request.Context(), request.ToCommand(id))
	if err != nil {
		httpx.HandleServiceError(c, err)
		return
	}

	httpx.OK(c, toResourceResponse(resource))
}

// DeleteResource godoc
// @Summary Delete resource
// @Description Deletes a resource
// @Tags resources
// @Security BearerAuth
// @Produce json
// @Param id path string true "Resource ID"
// @Success 200 {object} MessageEnvelope
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 404 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /resources/{id} [delete]
func (h *Handler) DeleteResource(c *gin.Context) {
	if _, ok := httpmiddleware.RequireUserID(c); !ok {
		return
	}

	id, ok := httpx.ParseUUIDParam(c, "id", "invalid resource id")
	if !ok {
		return
	}

	if err := h.service.DeleteResource(c.Request.Context(), id); err != nil {
		httpx.HandleServiceError(c, err)
		return
	}

	httpx.OK(c, MessageData{Message: "resource deleted successfully"})
}

// GetAvailability godoc
// @Summary Get resource availability
// @Description Returns resource availability and bookings for the specified date range
// @Tags resources
// @Security BearerAuth
// @Produce json
// @Param id path string true "Resource ID"
// @Param date_from query string true "RFC3339 lower bound"
// @Param date_to query string true "RFC3339 upper bound"
// @Success 200 {object} AvailabilityEnvelope
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 404 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /resources/{id}/availability [get]
func (h *Handler) GetAvailability(c *gin.Context) {
	if _, ok := httpmiddleware.RequireUserID(c); !ok {
		return
	}

	id, ok := httpx.ParseUUIDParam(c, "id", "invalid resource id")
	if !ok {
		return
	}

	query, err := buildAvailabilityQuery(c, id)
	if err != nil {
		httpx.Error(c, http.StatusBadRequest, "validation_error", "invalid date range format, expected RFC3339")
		return
	}

	availability, err := h.service.GetAvailability(c.Request.Context(), query)
	if err != nil {
		httpx.HandleServiceError(c, err)
		return
	}

	httpx.OK(c, toAvailabilityResponse(availability))
}

// CreateBooking godoc
// @Summary Create resource booking
// @Description Creates a booking for the selected resource
// @Tags resources
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Resource ID"
// @Param request body CreateBookingRequest true "Create booking request"
// @Success 201 {object} BookingEnvelope
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 404 {object} httpx.ErrorResponse
// @Failure 409 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /resources/{id}/bookings [post]
func (h *Handler) CreateBooking(c *gin.Context) {
	userID, ok := httpmiddleware.RequireUserID(c)
	if !ok {
		return
	}

	resourceID, ok := httpx.ParseUUIDParam(c, "id", "invalid resource id")
	if !ok {
		return
	}

	var request CreateBookingRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpx.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	command, err := request.ToCommand(resourceID, userID)
	if err != nil {
		httpx.Error(c, http.StatusBadRequest, "validation_error", "invalid date or event_id format")
		return
	}

	booking, err := h.service.CreateBooking(c.Request.Context(), command)
	if err != nil {
		httpx.HandleServiceError(c, err)
		return
	}

	httpx.Created(c, toBookingResponse(booking))
}

// CancelBooking godoc
// @Summary Cancel resource booking
// @Description Cancels a resource booking
// @Tags resources
// @Security BearerAuth
// @Produce json
// @Param id path string true "Resource ID"
// @Param bookingId path string true "Booking ID"
// @Success 200 {object} MessageEnvelope
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 403 {object} httpx.ErrorResponse
// @Failure 404 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /resources/{id}/bookings/{bookingId} [delete]
func (h *Handler) CancelBooking(c *gin.Context) {
	userID, ok := httpmiddleware.RequireUserID(c)
	if !ok {
		return
	}

	resourceID, ok := httpx.ParseUUIDParam(c, "id", "invalid resource id")
	if !ok {
		return
	}
	bookingID, ok := httpx.ParseUUIDParam(c, "bookingId", "invalid booking id")
	if !ok {
		return
	}

	if err := h.service.CancelBooking(c.Request.Context(), CancelBookingCommand{
		ResourceID: resourceID,
		BookingID:  bookingID,
		UserID:     userID,
	}); err != nil {
		httpx.HandleServiceError(c, err)
		return
	}

	httpx.OK(c, MessageData{Message: "resource booking cancelled successfully"})
}

func buildAvailabilityQuery(c *gin.Context, resourceID uuid.UUID) (GetResourceAvailabilityQuery, error) {
	dateFrom, err := time.Parse(time.RFC3339, c.Query("date_from"))
	if err != nil {
		return GetResourceAvailabilityQuery{}, err
	}
	dateTo, err := time.Parse(time.RFC3339, c.Query("date_to"))
	if err != nil {
		return GetResourceAvailabilityQuery{}, err
	}

	return GetResourceAvailabilityQuery{
		ResourceID: resourceID,
		DateFrom:   dateFrom,
		DateTo:     dateTo,
	}, nil
}

func toResourceResponse(resource *Resource) ResourceResponse {
	var ownerUserID *string
	if resource.OwnerUserID != nil {
		value := resource.OwnerUserID.String()
		ownerUserID = &value
	}

	return ResourceResponse{
		ID:          resource.ID.String(),
		Name:        resource.Name,
		Type:        string(resource.Type),
		Description: resource.Description,
		Capacity:    resource.Capacity,
		Location:    resource.Location,
		OwnerUserID: ownerUserID,
		CreatedAt:   resource.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   resource.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func toResourceResponses(resources []Resource) []ResourceResponse {
	result := make([]ResourceResponse, 0, len(resources))
	for i := range resources {
		resource := resources[i]
		result = append(result, toResourceResponse(&resource))
	}

	return result
}

func toBookingResponse(booking *ResourceBooking) BookingResponse {
	var eventID *string
	if booking.EventID != nil {
		value := booking.EventID.String()
		eventID = &value
	}

	return BookingResponse{
		ID:             booking.ID.String(),
		ResourceID:     booking.ResourceID.String(),
		EventID:        eventID,
		BookedByUserID: booking.BookedByUserID.String(),
		StartAt:        booking.StartAt.UTC().Format(time.RFC3339),
		EndAt:          booking.EndAt.UTC().Format(time.RFC3339),
		Title:          booking.Title,
		CreatedAt:      booking.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:      booking.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func toAvailabilityResponse(availability *ResourceAvailability) AvailabilityResponse {
	bookings := make([]BookingResponse, 0, len(availability.Bookings))
	for i := range availability.Bookings {
		booking := availability.Bookings[i]
		bookings = append(bookings, toBookingResponse(&booking))
	}

	return AvailabilityResponse{
		ResourceID:  availability.ResourceID.String(),
		DateFrom:    availability.DateFrom.UTC().Format(time.RFC3339),
		DateTo:      availability.DateTo.UTC().Format(time.RFC3339),
		IsAvailable: availability.IsAvailable,
		Bookings:    bookings,
	}
}
