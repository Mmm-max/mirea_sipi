package scheduling

import (
	"errors"
	"io"
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

// SearchSlots godoc
// @Summary Search recommended meeting slots
// @Description Searches and stores top recommended slots for the specified meeting
// @Tags scheduling
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Meeting ID"
// @Param request body SearchSlotsRequest false "Search slots request"
// @Success 200 {object} SearchSlotsEnvelope
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 403 {object} httpx.ErrorResponse
// @Failure 404 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /meetings/{id}/search-slots [post]
func (h *Handler) SearchSlots(c *gin.Context) {
	userID, ok := httpmiddleware.RequireUserID(c)
	if !ok {
		return
	}
	meetingID, ok := parseMeetingID(c)
	if !ok {
		return
	}
	var request SearchSlotsRequest
	if err := c.ShouldBindJSON(&request); err != nil && !errors.Is(err, io.EOF) {
		httpx.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	result, err := h.service.SearchSlots(c.Request.Context(), request.ToCommand(meetingID, userID))
	if err != nil {
		httpx.HandleServiceError(c, err)
		return
	}
	httpx.OK(c, toSearchSlotsResponse(result))
}

// GetSlots godoc
// @Summary Get stored meeting slots
// @Description Returns last stored slots for the specified meeting
// @Tags scheduling
// @Security BearerAuth
// @Produce json
// @Param id path string true "Meeting ID"
// @Success 200 {object} SearchSlotsEnvelope
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 404 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /meetings/{id}/slots [get]
func (h *Handler) GetSlots(c *gin.Context) {
	userID, ok := httpmiddleware.RequireUserID(c)
	if !ok {
		return
	}
	meetingID, ok := parseMeetingID(c)
	if !ok {
		return
	}
	result, err := h.service.GetSlots(c.Request.Context(), GetSlotsQuery{
		MeetingID:       meetingID,
		RequesterUserID: userID,
	})
	if err != nil {
		httpx.HandleServiceError(c, err)
		return
	}
	httpx.OK(c, toSearchSlotsResponse(result))
}

// SelectSlot godoc
// @Summary Select meeting slot
// @Description Selects one of the previously generated slots for the meeting
// @Tags scheduling
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Meeting ID"
// @Param request body SelectSlotRequest true "Select slot request"
// @Success 200 {object} httpx.SuccessResponse
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 403 {object} httpx.ErrorResponse
// @Failure 404 {object} httpx.ErrorResponse
// @Failure 409 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /meetings/{id}/select-slot [post]
func (h *Handler) SelectSlot(c *gin.Context) {
	userID, ok := httpmiddleware.RequireUserID(c)
	if !ok {
		return
	}
	meetingID, ok := parseMeetingID(c)
	if !ok {
		return
	}
	var request SelectSlotRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpx.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	command, err := request.ToCommand(meetingID, userID)
	if err != nil {
		httpx.Error(c, http.StatusBadRequest, "validation_error", "invalid meeting_slot_id")
		return
	}
	if _, err := h.service.SelectSlot(c.Request.Context(), command); err != nil {
		httpx.HandleServiceError(c, err)
		return
	}
	httpx.OK(c, MessageData{Message: "slot selected successfully"})
}

func toSearchSlotsResponse(result *SearchResult) SearchSlotsResponse {
	response := SearchSlotsResponse{
		MeetingID: result.MeetingID.String(),
		Slots:     make([]MeetingSlotResponse, 0, len(result.Slots)),
	}
	for _, slot := range result.Slots {
		response.Slots = append(response.Slots, MeetingSlotResponse{
			ID:        slot.ID.String(),
			StartAt:   slot.StartAt.UTC().Format(time.RFC3339),
			EndAt:     slot.EndAt.UTC().Format(time.RFC3339),
			Score:     slot.Score,
			Rank:      slot.Rank,
			Conflicts: toConflictResponses(slot.Conflicts),
		})
	}
	return response
}

func toConflictResponses(conflicts []SlotConflict) []SlotConflictResponse {
	result := make([]SlotConflictResponse, 0, len(conflicts))
	for _, conflict := range conflicts {
		result = append(result, SlotConflictResponse{
			UserID:          formatUUIDPointer(conflict.UserID),
			EventID:         formatUUIDPointer(conflict.EventID),
			ResourceID:      formatUUIDPointer(conflict.ResourceID),
			ConflictType:    conflict.ConflictType,
			VisibleTitle:    conflict.VisibleTitle,
			VisiblePriority: conflict.VisiblePriority,
		})
	}
	return result
}

func formatUUIDPointer(value *uuid.UUID) *string {
	if value == nil {
		return nil
	}
	formatted := value.String()
	return &formatted
}

func parseMeetingID(c *gin.Context) (uuid.UUID, bool) {
	return httpx.ParseUUIDParam(c, "id", "invalid meeting id")
}
