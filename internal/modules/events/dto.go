package events

import (
	"time"

	"github.com/google/uuid"
)

type CreateEventRequest struct {
	SourceType      string  `json:"source_type" binding:"required,oneof=personal_task meeting imported blocked_time" example:"personal_task"`
	ExternalUID     *string `json:"external_uid" example:"ext-123"`
	Title           string  `json:"title" binding:"required,min=1,max=255" example:"Plan sprint backlog"`
	Description     string  `json:"description" binding:"omitempty,max=5000" example:"Prepare tasks for next sprint"`
	StartAt         string  `json:"start_at" binding:"required" example:"2026-03-27T09:00:00Z"`
	EndAt           string  `json:"end_at" binding:"required" example:"2026-03-27T10:00:00Z"`
	Priority        string  `json:"priority" binding:"required,oneof=low medium high critical" example:"high"`
	IsReschedulable bool    `json:"is_reschedulable" example:"true"`
	VisibilityHint  string  `json:"visibility_hint" binding:"omitempty,max=64" example:"busy"`
}

type UpdateEventRequest struct {
	SourceType       *string `json:"source_type,omitempty" binding:"omitempty,oneof=personal_task meeting imported blocked_time" example:"meeting"`
	ExternalUID      *string `json:"external_uid,omitempty" example:"ext-456"`
	ClearExternalUID *bool   `json:"clear_external_uid,omitempty" example:"false"`
	Title            *string `json:"title,omitempty" binding:"omitempty,min=1,max=255" example:"Updated title"`
	Description      *string `json:"description,omitempty" binding:"omitempty,max=5000" example:"Updated description"`
	StartAt          *string `json:"start_at,omitempty" example:"2026-03-27T11:00:00Z"`
	EndAt            *string `json:"end_at,omitempty" example:"2026-03-27T12:00:00Z"`
	Priority         *string `json:"priority,omitempty" binding:"omitempty,oneof=low medium high critical" example:"critical"`
	IsReschedulable  *bool   `json:"is_reschedulable,omitempty" example:"false"`
	VisibilityHint   *string `json:"visibility_hint,omitempty" binding:"omitempty,max=64" example:"focus"`
}

type EventResponse struct {
	ID              string  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	OwnerUserID     string  `json:"owner_user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	SourceType      string  `json:"source_type" example:"personal_task"`
	ExternalUID     *string `json:"external_uid,omitempty" example:"ext-123"`
	Title           string  `json:"title" example:"Plan sprint backlog"`
	Description     string  `json:"description" example:"Prepare tasks for next sprint"`
	StartAt         string  `json:"start_at" example:"2026-03-27T09:00:00Z"`
	EndAt           string  `json:"end_at" example:"2026-03-27T10:00:00Z"`
	Priority        string  `json:"priority" example:"high"`
	IsReschedulable bool    `json:"is_reschedulable" example:"true"`
	VisibilityHint  string  `json:"visibility_hint" example:"busy"`
	CreatedAt       string  `json:"created_at" example:"2026-03-26T12:00:00Z"`
	UpdatedAt       string  `json:"updated_at" example:"2026-03-26T12:00:00Z"`
}

type EventEnvelope struct {
	Success bool            `json:"success" example:"true"`
	Data    EventResponse   `json:"data"`
	Meta    ResponseMetaDTO `json:"meta"`
}

type EventListResponse struct {
	Items []EventResponse `json:"items"`
}

type EventListEnvelope struct {
	Success bool              `json:"success" example:"true"`
	Data    EventListResponse `json:"data"`
	Meta    ResponseMetaDTO   `json:"meta"`
}

type MessageData struct {
	Message string `json:"message" example:"event deleted successfully"`
}

type MessageEnvelope struct {
	Success bool            `json:"success" example:"true"`
	Data    MessageData     `json:"data"`
	Meta    ResponseMetaDTO `json:"meta"`
}

type ResponseMetaDTO struct {
	RequestID string `json:"request_id" example:"e7cc5d4a-3d85-4938-bff3-d2a6cf8ca2bb"`
}

func (r CreateEventRequest) ToCommand(ownerUserID uuid.UUID) (CreateEventCommand, error) {
	startAt, err := time.Parse(time.RFC3339, r.StartAt)
	if err != nil {
		return CreateEventCommand{}, err
	}
	endAt, err := time.Parse(time.RFC3339, r.EndAt)
	if err != nil {
		return CreateEventCommand{}, err
	}

	return CreateEventCommand{
		OwnerUserID:     ownerUserID,
		SourceType:      SourceType(r.SourceType),
		ExternalUID:     r.ExternalUID,
		Title:           r.Title,
		Description:     r.Description,
		StartAt:         startAt,
		EndAt:           endAt,
		Priority:        Priority(r.Priority),
		IsReschedulable: r.IsReschedulable,
		VisibilityHint:  r.VisibilityHint,
	}, nil
}

func (r UpdateEventRequest) ToCommand(id, ownerUserID uuid.UUID) (UpdateEventCommand, error) {
	command := UpdateEventCommand{
		ID:              id,
		OwnerUserID:     ownerUserID,
		Title:           r.Title,
		Description:     r.Description,
		IsReschedulable: r.IsReschedulable,
		VisibilityHint:  r.VisibilityHint,
	}

	if r.SourceType != nil {
		value := SourceType(*r.SourceType)
		command.SourceType = &value
	}
	if r.Priority != nil {
		value := Priority(*r.Priority)
		command.Priority = &value
	}
	if r.ExternalUID != nil {
		command.ExternalUIDSet = true
		command.ExternalUID = r.ExternalUID
	}
	if r.ClearExternalUID != nil {
		command.ClearExternalUID = *r.ClearExternalUID
		command.ExternalUIDSet = command.ExternalUIDSet || *r.ClearExternalUID
	}
	if r.StartAt != nil {
		parsed, err := time.Parse(time.RFC3339, *r.StartAt)
		if err != nil {
			return UpdateEventCommand{}, err
		}
		command.StartAt = &parsed
	}
	if r.EndAt != nil {
		parsed, err := time.Parse(time.RFC3339, *r.EndAt)
		if err != nil {
			return UpdateEventCommand{}, err
		}
		command.EndAt = &parsed
	}

	return command, nil
}
