package meetings

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

// CreateMeeting godoc
// @Summary Create meeting
// @Description Creates a new meeting
// @Tags meetings
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CreateMeetingRequest true "Create meeting request"
// @Success 201 {object} MeetingEnvelope
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /meetings [post]
func (h *Handler) CreateMeeting(c *gin.Context) {
	userID, ok := httpmiddleware.RequireUserID(c)
	if !ok {
		return
	}
	var request CreateMeetingRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpx.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	command, err := request.ToCommand(userID)
	if err != nil {
		httpx.Error(c, http.StatusBadRequest, "validation_error", "invalid datetime format, expected RFC3339")
		return
	}
	meeting, err := h.service.CreateMeeting(c.Request.Context(), command)
	if err != nil {
		httpx.HandleServiceError(c, err)
		return
	}
	httpx.Created(c, toMeetingResponse(meeting))
}

// ListMeetings godoc
// @Summary List meetings
// @Description Returns meetings where the current user is organizer or participant
// @Tags meetings
// @Security BearerAuth
// @Produce json
// @Success 200 {object} MeetingListEnvelope
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /meetings [get]
func (h *Handler) ListMeetings(c *gin.Context) {
	userID, ok := httpmiddleware.RequireUserID(c)
	if !ok {
		return
	}
	items, err := h.service.ListMeetings(c.Request.Context(), ListMeetingsQuery{UserID: userID})
	if err != nil {
		httpx.HandleServiceError(c, err)
		return
	}
	httpx.OK(c, MeetingListResponse{Items: toMeetingResponses(items)})
}

// GetMeeting godoc
// @Summary Get meeting by id
// @Description Returns a meeting by id
// @Tags meetings
// @Security BearerAuth
// @Produce json
// @Param id path string true "Meeting ID"
// @Success 200 {object} MeetingEnvelope
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 404 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /meetings/{id} [get]
func (h *Handler) GetMeeting(c *gin.Context) {
	userID, ok := httpmiddleware.RequireUserID(c)
	if !ok {
		return
	}
	meetingID, ok := parseMeetingID(c)
	if !ok {
		return
	}
	meeting, err := h.service.GetMeeting(c.Request.Context(), GetMeetingQuery{ID: meetingID, UserID: userID})
	if err != nil {
		httpx.HandleServiceError(c, err)
		return
	}
	httpx.OK(c, toMeetingResponse(meeting))
}

// UpdateMeeting godoc
// @Summary Update meeting
// @Description Updates a meeting
// @Tags meetings
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Meeting ID"
// @Param request body UpdateMeetingRequest true "Update meeting request"
// @Success 200 {object} MeetingEnvelope
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 403 {object} httpx.ErrorResponse
// @Failure 404 {object} httpx.ErrorResponse
// @Failure 409 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /meetings/{id} [patch]
func (h *Handler) UpdateMeeting(c *gin.Context) {
	userID, ok := httpmiddleware.RequireUserID(c)
	if !ok {
		return
	}
	meetingID, ok := parseMeetingID(c)
	if !ok {
		return
	}
	var request UpdateMeetingRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpx.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	command, err := request.ToCommand(meetingID, userID)
	if err != nil {
		httpx.Error(c, http.StatusBadRequest, "validation_error", "invalid datetime format, expected RFC3339")
		return
	}
	meeting, err := h.service.UpdateMeeting(c.Request.Context(), command)
	if err != nil {
		httpx.HandleServiceError(c, err)
		return
	}
	httpx.OK(c, toMeetingResponse(meeting))
}

// DeleteMeeting godoc
// @Summary Delete meeting
// @Description Deletes a meeting
// @Tags meetings
// @Security BearerAuth
// @Produce json
// @Param id path string true "Meeting ID"
// @Success 200 {object} MessageEnvelope
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 403 {object} httpx.ErrorResponse
// @Failure 404 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /meetings/{id} [delete]
func (h *Handler) DeleteMeeting(c *gin.Context) {
	userID, ok := httpmiddleware.RequireUserID(c)
	if !ok {
		return
	}
	meetingID, ok := parseMeetingID(c)
	if !ok {
		return
	}
	if err := h.service.DeleteMeeting(c.Request.Context(), meetingID, userID); err != nil {
		httpx.HandleServiceError(c, err)
		return
	}
	httpx.OK(c, MessageData{Message: "meeting deleted successfully"})
}

// AddParticipants godoc
// @Summary Add meeting participants
// @Description Adds participants to a meeting
// @Tags meetings
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Meeting ID"
// @Param request body AddParticipantsRequest true "Participants request"
// @Success 200 {object} MeetingEnvelope
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 403 {object} httpx.ErrorResponse
// @Failure 404 {object} httpx.ErrorResponse
// @Failure 409 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /meetings/{id}/participants [post]
func (h *Handler) AddParticipants(c *gin.Context) {
	userID, ok := httpmiddleware.RequireUserID(c)
	if !ok {
		return
	}
	meetingID, ok := parseMeetingID(c)
	if !ok {
		return
	}
	var request AddParticipantsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpx.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	command, err := request.ToCommand(meetingID, userID)
	if err != nil {
		httpx.Error(c, http.StatusBadRequest, "validation_error", "invalid participant id")
		return
	}
	meeting, err := h.service.AddParticipants(c.Request.Context(), command)
	if err != nil {
		httpx.HandleServiceError(c, err)
		return
	}
	httpx.OK(c, toMeetingResponse(meeting))
}

// RemoveParticipant godoc
// @Summary Remove meeting participant
// @Description Removes a participant from a meeting
// @Tags meetings
// @Security BearerAuth
// @Produce json
// @Param id path string true "Meeting ID"
// @Param userId path string true "Participant user ID"
// @Success 200 {object} MessageEnvelope
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 403 {object} httpx.ErrorResponse
// @Failure 404 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /meetings/{id}/participants/{userId} [delete]
func (h *Handler) RemoveParticipant(c *gin.Context) {
	userID, ok := httpmiddleware.RequireUserID(c)
	if !ok {
		return
	}
	meetingID, ok := parseMeetingID(c)
	if !ok {
		return
	}
	participantUserID, ok := parseUUIDParam(c, "userId", "invalid participant user id")
	if !ok {
		return
	}
	if err := h.service.RemoveParticipant(c.Request.Context(), RemoveParticipantCommand{
		MeetingID:         meetingID,
		ActingUserID:      userID,
		ParticipantUserID: participantUserID,
	}); err != nil {
		httpx.HandleServiceError(c, err)
		return
	}
	httpx.OK(c, MessageData{Message: "participant removed successfully"})
}

// AddResources godoc
// @Summary Add meeting resources
// @Description Adds resources to a meeting
// @Tags meetings
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Meeting ID"
// @Param request body AddResourcesRequest true "Resources request"
// @Success 200 {object} MeetingEnvelope
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 403 {object} httpx.ErrorResponse
// @Failure 404 {object} httpx.ErrorResponse
// @Failure 409 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /meetings/{id}/resources [post]
func (h *Handler) AddResources(c *gin.Context) {
	userID, ok := httpmiddleware.RequireUserID(c)
	if !ok {
		return
	}
	meetingID, ok := parseMeetingID(c)
	if !ok {
		return
	}
	var request AddResourcesRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpx.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	command, err := request.ToCommand(meetingID, userID)
	if err != nil {
		httpx.Error(c, http.StatusBadRequest, "validation_error", "invalid resource id")
		return
	}
	meeting, err := h.service.AddResources(c.Request.Context(), command)
	if err != nil {
		httpx.HandleServiceError(c, err)
		return
	}
	httpx.OK(c, toMeetingResponse(meeting))
}

// RemoveResource godoc
// @Summary Remove meeting resource
// @Description Removes a resource from a meeting
// @Tags meetings
// @Security BearerAuth
// @Produce json
// @Param id path string true "Meeting ID"
// @Param resourceId path string true "Resource ID"
// @Success 200 {object} MessageEnvelope
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 403 {object} httpx.ErrorResponse
// @Failure 404 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /meetings/{id}/resources/{resourceId} [delete]
func (h *Handler) RemoveResource(c *gin.Context) {
	userID, ok := httpmiddleware.RequireUserID(c)
	if !ok {
		return
	}
	meetingID, ok := parseMeetingID(c)
	if !ok {
		return
	}
	resourceID, ok := parseUUIDParam(c, "resourceId", "invalid resource id")
	if !ok {
		return
	}
	if err := h.service.RemoveResource(c.Request.Context(), RemoveResourceCommand{
		MeetingID:       meetingID,
		OrganizerUserID: userID,
		ResourceID:      resourceID,
	}); err != nil {
		httpx.HandleServiceError(c, err)
		return
	}
	httpx.OK(c, MessageData{Message: "resource removed successfully"})
}

// RespondInvitation godoc
// @Summary Respond to invitation
// @Description Updates current participant response status
// @Tags meetings
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Meeting ID"
// @Param request body RespondInvitationRequest true "Invitation response request"
// @Success 200 {object} ParticipantEnvelope
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 403 {object} httpx.ErrorResponse
// @Failure 404 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /meetings/{id}/respond [post]
func (h *Handler) RespondInvitation(c *gin.Context) {
	userID, ok := httpmiddleware.RequireUserID(c)
	if !ok {
		return
	}
	meetingID, ok := parseMeetingID(c)
	if !ok {
		return
	}
	var request RespondInvitationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpx.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	participant, err := h.service.RespondInvitation(c.Request.Context(), request.ToCommand(meetingID, userID))
	if err != nil {
		httpx.HandleServiceError(c, err)
		return
	}
	httpx.OK(c, toParticipantResponse(participant))
}

// RequestAlternative godoc
// @Summary Request alternative meeting time
// @Description Allows a participant to request an alternative time
// @Tags meetings
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Meeting ID"
// @Param request body RequestAlternativeRequest true "Alternative time request"
// @Success 200 {object} ParticipantEnvelope
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 403 {object} httpx.ErrorResponse
// @Failure 404 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /meetings/{id}/request-alternative [post]
func (h *Handler) RequestAlternative(c *gin.Context) {
	userID, ok := httpmiddleware.RequireUserID(c)
	if !ok {
		return
	}
	meetingID, ok := parseMeetingID(c)
	if !ok {
		return
	}
	var request RequestAlternativeRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpx.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	command, err := request.ToCommand(meetingID, userID)
	if err != nil {
		httpx.Error(c, http.StatusBadRequest, "validation_error", "invalid datetime format, expected RFC3339")
		return
	}
	participant, err := h.service.RequestAlternative(c.Request.Context(), command)
	if err != nil {
		httpx.HandleServiceError(c, err)
		return
	}
	httpx.OK(c, toParticipantResponse(participant))
}

func toMeetingResponse(meeting *Meeting) MeetingResponse {
	return MeetingResponse{
		ID:                meeting.ID.String(),
		OrganizerUserID:   meeting.OrganizerUserID.String(),
		Title:             meeting.Title,
		Description:       meeting.Description,
		DurationMinutes:   meeting.DurationMinutes,
		Priority:          meeting.Priority,
		Status:            string(meeting.Status),
		SearchRangeStart:  meeting.SearchRangeStart.UTC().Format(time.RFC3339),
		SearchRangeEnd:    meeting.SearchRangeEnd.UTC().Format(time.RFC3339),
		EarliestStartTime: meeting.EarliestStartTime,
		LatestStartTime:   meeting.LatestStartTime,
		RecurrenceRule:    meeting.RecurrenceRule,
		SelectedStartAt:   formatTimePointer(meeting.SelectedStartAt),
		SelectedEndAt:     formatTimePointer(meeting.SelectedEndAt),
		Participants:      toParticipantResponses(meeting.Participants),
		Resources:         toMeetingResourceResponses(meeting.Resources),
		CreatedAt:         meeting.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:         meeting.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func toMeetingResponses(items []Meeting) []MeetingResponse {
	result := make([]MeetingResponse, 0, len(items))
	for i := range items {
		meeting := items[i]
		result = append(result, toMeetingResponse(&meeting))
	}
	return result
}

func toParticipantResponse(item *MeetingParticipant) MeetingParticipantResponse {
	return MeetingParticipantResponse{
		UserID:             item.UserID.String(),
		ResponseStatus:     string(item.ResponseStatus),
		VisibilityOverride: item.VisibilityOverride,
		AlternativeStartAt: formatTimePointer(item.AlternativeStartAt),
		AlternativeEndAt:   formatTimePointer(item.AlternativeEndAt),
		AlternativeComment: item.AlternativeComment,
	}
}

func toParticipantResponses(items []MeetingParticipant) []MeetingParticipantResponse {
	result := make([]MeetingParticipantResponse, 0, len(items))
	for i := range items {
		item := items[i]
		result = append(result, toParticipantResponse(&item))
	}
	return result
}

func toMeetingResourceResponses(items []MeetingResource) []MeetingResourceResponse {
	result := make([]MeetingResourceResponse, 0, len(items))
	for i := range items {
		result = append(result, MeetingResourceResponse{ResourceID: items[i].ResourceID.String()})
	}
	return result
}

func formatTimePointer(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.UTC().Format(time.RFC3339)
	return &formatted
}

func parseMeetingID(c *gin.Context) (uuid.UUID, bool) {
	return parseUUIDParam(c, "id", "invalid meeting id")
}

func parseUUIDParam(c *gin.Context, paramName, errorMessage string) (uuid.UUID, bool) {
	return httpx.ParseUUIDParam(c, paramName, errorMessage)
}
