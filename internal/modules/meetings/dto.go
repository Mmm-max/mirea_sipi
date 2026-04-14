package meetings

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type CreateMeetingRequest struct {
	Title             string  `json:"title" binding:"required,min=1,max=255" example:"Sprint planning"`
	Description       string  `json:"description" binding:"omitempty,max=5000" example:"Weekly sprint planning meeting"`
	DurationMinutes   int     `json:"duration_minutes" binding:"required,min=1" example:"60"`
	Priority          string  `json:"priority" binding:"required,oneof=low medium high critical" example:"high"`
	Status            *string `json:"status,omitempty" binding:"omitempty,oneof=draft scheduled cancelled" example:"draft"`
	SearchRangeStart  string  `json:"search_range_start" binding:"required" example:"2026-03-27T09:00:00Z"`
	SearchRangeEnd    string  `json:"search_range_end" binding:"required" example:"2026-03-27T18:00:00Z"`
	EarliestStartTime *string `json:"earliest_start_time,omitempty" example:"09:00"`
	LatestStartTime   *string `json:"latest_start_time,omitempty" example:"18:00"`
	RecurrenceRule    *string `json:"recurrence_rule,omitempty" example:"FREQ=WEEKLY;COUNT=4"`
	SelectedStartAt   *string `json:"selected_start_at,omitempty" example:"2026-03-27T10:00:00Z"`
	SelectedEndAt     *string `json:"selected_end_at,omitempty" example:"2026-03-27T11:00:00Z"`
}

type UpdateMeetingRequest struct {
	Title             *string `json:"title,omitempty" binding:"omitempty,min=1,max=255" example:"Updated sprint planning"`
	Description       *string `json:"description,omitempty" binding:"omitempty,max=5000" example:"Updated description"`
	DurationMinutes   *int    `json:"duration_minutes,omitempty" binding:"omitempty,min=1" example:"90"`
	Priority          *string `json:"priority,omitempty" binding:"omitempty,oneof=low medium high critical" example:"critical"`
	Status            *string `json:"status,omitempty" binding:"omitempty,oneof=draft scheduled cancelled" example:"scheduled"`
	SearchRangeStart  *string `json:"search_range_start,omitempty" example:"2026-03-28T09:00:00Z"`
	SearchRangeEnd    *string `json:"search_range_end,omitempty" example:"2026-03-28T18:00:00Z"`
	EarliestStartTime *string `json:"earliest_start_time,omitempty" example:"10:00"`
	ClearEarliest     *bool   `json:"clear_earliest_start_time,omitempty" example:"false"`
	LatestStartTime   *string `json:"latest_start_time,omitempty" example:"17:00"`
	ClearLatest       *bool   `json:"clear_latest_start_time,omitempty" example:"false"`
	RecurrenceRule    *string `json:"recurrence_rule,omitempty" example:"FREQ=DAILY;COUNT=3"`
	ClearRecurrence   *bool   `json:"clear_recurrence_rule,omitempty" example:"false"`
	SelectedStartAt   *string `json:"selected_start_at,omitempty" example:"2026-03-28T12:00:00Z"`
	SelectedEndAt     *string `json:"selected_end_at,omitempty" example:"2026-03-28T13:30:00Z"`
	ClearSelected     *bool   `json:"clear_selected_time,omitempty" example:"false"`
}

type AddParticipantsRequest struct {
	UserIDs []string `json:"user_ids" binding:"required,min=1,dive,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
}

type AddResourcesRequest struct {
	ResourceIDs []string `json:"resource_ids" binding:"required,min=1,dive,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
}

type RespondInvitationRequest struct {
	ResponseStatus string `json:"response_status" binding:"required,oneof=pending accepted declined tentative" example:"accepted"`
}

type RequestAlternativeRequest struct {
	ProposedStartAt string  `json:"proposed_start_at" binding:"required" example:"2026-03-27T14:00:00Z"`
	ProposedEndAt   string  `json:"proposed_end_at" binding:"required" example:"2026-03-27T15:00:00Z"`
	Comment         *string `json:"comment,omitempty" binding:"omitempty,max=1000" example:"Can we move this after lunch?"`
}

type MeetingParticipantResponse struct {
	UserID             string  `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	ResponseStatus     string  `json:"response_status" example:"pending"`
	VisibilityOverride *string `json:"visibility_override,omitempty" example:"busy"`
	AlternativeStartAt *string `json:"alternative_start_at,omitempty" example:"2026-03-27T14:00:00Z"`
	AlternativeEndAt   *string `json:"alternative_end_at,omitempty" example:"2026-03-27T15:00:00Z"`
	AlternativeComment *string `json:"alternative_comment,omitempty" example:"Can we move this after lunch?"`
}

type MeetingResourceResponse struct {
	ResourceID string `json:"resource_id" example:"550e8400-e29b-41d4-a716-446655440000"`
}

type MeetingResponse struct {
	ID                string                       `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	OrganizerUserID   string                       `json:"organizer_user_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	Title             string                       `json:"title" example:"Sprint planning"`
	Description       string                       `json:"description" example:"Weekly sprint planning meeting"`
	DurationMinutes   int                          `json:"duration_minutes" example:"60"`
	Priority          string                       `json:"priority" example:"high"`
	Status            string                       `json:"status" example:"scheduled"`
	SearchRangeStart  string                       `json:"search_range_start" example:"2026-03-27T09:00:00Z"`
	SearchRangeEnd    string                       `json:"search_range_end" example:"2026-03-27T18:00:00Z"`
	EarliestStartTime *string                      `json:"earliest_start_time,omitempty" example:"09:00"`
	LatestStartTime   *string                      `json:"latest_start_time,omitempty" example:"18:00"`
	RecurrenceRule    *string                      `json:"recurrence_rule,omitempty" example:"FREQ=WEEKLY;COUNT=4"`
	SelectedStartAt   *string                      `json:"selected_start_at,omitempty" example:"2026-03-27T10:00:00Z"`
	SelectedEndAt     *string                      `json:"selected_end_at,omitempty" example:"2026-03-27T11:00:00Z"`
	Participants      []MeetingParticipantResponse `json:"participants"`
	Resources         []MeetingResourceResponse    `json:"resources"`
	CreatedAt         string                       `json:"created_at" example:"2026-03-26T12:00:00Z"`
	UpdatedAt         string                       `json:"updated_at" example:"2026-03-26T12:00:00Z"`
}

type ParticipantEnvelope struct {
	Success bool                       `json:"success" example:"true"`
	Data    MeetingParticipantResponse `json:"data"`
	Meta    ResponseMetaDTO            `json:"meta"`
}

type MeetingEnvelope struct {
	Success bool            `json:"success" example:"true"`
	Data    MeetingResponse `json:"data"`
	Meta    ResponseMetaDTO `json:"meta"`
}

type MeetingListResponse struct {
	Items []MeetingResponse `json:"items"`
}

type MeetingListEnvelope struct {
	Success bool                `json:"success" example:"true"`
	Data    MeetingListResponse `json:"data"`
	Meta    ResponseMetaDTO     `json:"meta"`
}

type MessageData struct {
	Message string `json:"message" example:"operation completed successfully"`
}

type MessageEnvelope struct {
	Success bool            `json:"success" example:"true"`
	Data    MessageData     `json:"data"`
	Meta    ResponseMetaDTO `json:"meta"`
}

type ResponseMetaDTO struct {
	RequestID string `json:"request_id" example:"e7cc5d4a-3d85-4938-bff3-d2a6cf8ca2bb"`
}

func (r CreateMeetingRequest) ToCommand(organizerUserID uuid.UUID) (CreateMeetingCommand, error) {
	searchRangeStart, err := time.Parse(time.RFC3339, r.SearchRangeStart)
	if err != nil {
		return CreateMeetingCommand{}, err
	}
	searchRangeEnd, err := time.Parse(time.RFC3339, r.SearchRangeEnd)
	if err != nil {
		return CreateMeetingCommand{}, err
	}
	var status *MeetingStatus
	if r.Status != nil {
		value := MeetingStatus(*r.Status)
		status = &value
	}
	selectedStartAt, selectedEndAt, err := parseOptionalTimeRange(r.SelectedStartAt, r.SelectedEndAt)
	if err != nil {
		return CreateMeetingCommand{}, err
	}
	return CreateMeetingCommand{
		OrganizerUserID:   organizerUserID,
		Title:             r.Title,
		Description:       r.Description,
		DurationMinutes:   r.DurationMinutes,
		Priority:          r.Priority,
		Status:            status,
		SearchRangeStart:  searchRangeStart,
		SearchRangeEnd:    searchRangeEnd,
		EarliestStartTime: r.EarliestStartTime,
		LatestStartTime:   r.LatestStartTime,
		RecurrenceRule:    r.RecurrenceRule,
		SelectedStartAt:   selectedStartAt,
		SelectedEndAt:     selectedEndAt,
	}, nil
}

func (r UpdateMeetingRequest) ToCommand(id, organizerUserID uuid.UUID) (UpdateMeetingCommand, error) {
	command := UpdateMeetingCommand{
		ID:                id,
		OrganizerUserID:   organizerUserID,
		Title:             r.Title,
		Description:       r.Description,
		DurationMinutes:   r.DurationMinutes,
		Priority:          r.Priority,
		EarliestStartTime: r.EarliestStartTime,
		LatestStartTime:   r.LatestStartTime,
		RecurrenceRule:    r.RecurrenceRule,
	}
	if r.Status != nil {
		value := MeetingStatus(*r.Status)
		command.Status = &value
	}
	if r.SearchRangeStart != nil {
		parsed, err := time.Parse(time.RFC3339, *r.SearchRangeStart)
		if err != nil {
			return UpdateMeetingCommand{}, err
		}
		command.SearchRangeStart = &parsed
	}
	if r.SearchRangeEnd != nil {
		parsed, err := time.Parse(time.RFC3339, *r.SearchRangeEnd)
		if err != nil {
			return UpdateMeetingCommand{}, err
		}
		command.SearchRangeEnd = &parsed
	}
	if r.ClearEarliest != nil {
		command.ClearEarliest = *r.ClearEarliest
		command.EarliestSet = command.EarliestSet || *r.ClearEarliest
	}
	if r.EarliestStartTime != nil {
		command.EarliestSet = true
	}
	if r.ClearLatest != nil {
		command.ClearLatest = *r.ClearLatest
		command.LatestSet = command.LatestSet || *r.ClearLatest
	}
	if r.LatestStartTime != nil {
		command.LatestSet = true
	}
	if r.ClearRecurrence != nil {
		command.ClearRecurrence = *r.ClearRecurrence
		command.RecurrenceSet = command.RecurrenceSet || *r.ClearRecurrence
	}
	if r.RecurrenceRule != nil {
		command.RecurrenceSet = true
	}
	selectedStartAt, selectedEndAt, err := parseOptionalTimeRange(r.SelectedStartAt, r.SelectedEndAt)
	if err != nil {
		return UpdateMeetingCommand{}, err
	}
	if r.ClearSelected != nil {
		command.ClearSelected = *r.ClearSelected
		command.SelectedSet = command.SelectedSet || *r.ClearSelected
	}
	if r.SelectedStartAt != nil || r.SelectedEndAt != nil {
		command.SelectedSet = true
		command.SelectedStartAt = selectedStartAt
		command.SelectedEndAt = selectedEndAt
	}
	return command, nil
}

func (r AddParticipantsRequest) ToCommand(meetingID, organizerUserID uuid.UUID) (AddParticipantsCommand, error) {
	userIDs := make([]uuid.UUID, 0, len(r.UserIDs))
	for _, raw := range r.UserIDs {
		parsed, err := uuid.Parse(raw)
		if err != nil {
			return AddParticipantsCommand{}, err
		}
		userIDs = append(userIDs, parsed)
	}
	return AddParticipantsCommand{MeetingID: meetingID, OrganizerUserID: organizerUserID, ParticipantIDs: userIDs}, nil
}

func (r AddResourcesRequest) ToCommand(meetingID, organizerUserID uuid.UUID) (AddResourceCommand, error) {
	resourceIDs := make([]uuid.UUID, 0, len(r.ResourceIDs))
	for _, raw := range r.ResourceIDs {
		parsed, err := uuid.Parse(raw)
		if err != nil {
			return AddResourceCommand{}, err
		}
		resourceIDs = append(resourceIDs, parsed)
	}
	return AddResourceCommand{MeetingID: meetingID, OrganizerUserID: organizerUserID, ResourceIDs: resourceIDs}, nil
}

func (r RespondInvitationRequest) ToCommand(meetingID, userID uuid.UUID) RespondInvitationCommand {
	return RespondInvitationCommand{MeetingID: meetingID, UserID: userID, ResponseStatus: InvitationStatus(r.ResponseStatus)}
}

func (r RequestAlternativeRequest) ToCommand(meetingID, userID uuid.UUID) (RequestAlternativeCommand, error) {
	startAt, err := time.Parse(time.RFC3339, r.ProposedStartAt)
	if err != nil {
		return RequestAlternativeCommand{}, err
	}
	endAt, err := time.Parse(time.RFC3339, r.ProposedEndAt)
	if err != nil {
		return RequestAlternativeCommand{}, err
	}
	return RequestAlternativeCommand{
		MeetingID:       meetingID,
		UserID:          userID,
		ProposedStartAt: startAt,
		ProposedEndAt:   endAt,
		Comment:         r.Comment,
	}, nil
}

func parseOptionalTimeRange(startValue, endValue *string) (*time.Time, *time.Time, error) {
	if startValue == nil && endValue == nil {
		return nil, nil, nil
	}
	if startValue == nil || endValue == nil {
		return nil, nil, errors.New("selected time range is incomplete")
	}
	startAt, err := time.Parse(time.RFC3339, *startValue)
	if err != nil {
		return nil, nil, err
	}
	endAt, err := time.Parse(time.RFC3339, *endValue)
	if err != nil {
		return nil, nil, err
	}
	return &startAt, &endAt, nil
}
