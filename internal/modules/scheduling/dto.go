package scheduling

import (
	"github.com/google/uuid"
)

type SearchSlotsRequest struct {
	TopN int `json:"top_n,omitempty" binding:"omitempty,min=1,max=50" example:"10"`
}

type SelectSlotRequest struct {
	MeetingSlotID string `json:"meeting_slot_id" binding:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
}

type SlotConflictResponse struct {
	UserID          *string `json:"user_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	EventID         *string `json:"event_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	ResourceID      *string `json:"resource_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	ConflictType    string  `json:"conflict_type" example:"event_conflict"`
	VisibleTitle    string  `json:"visible_title" example:"Busy"`
	VisiblePriority *string `json:"visible_priority,omitempty" example:"medium"`
}

type MeetingSlotResponse struct {
	ID        string                 `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	StartAt   string                 `json:"start_at" example:"2026-03-27T10:00:00Z"`
	EndAt     string                 `json:"end_at" example:"2026-03-27T11:00:00Z"`
	Score     int                    `json:"score" example:"85"`
	Rank      int                    `json:"rank" example:"1"`
	Conflicts []SlotConflictResponse `json:"conflicts"`
}

type SearchSlotsResponse struct {
	MeetingID string                `json:"meeting_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Slots     []MeetingSlotResponse `json:"slots"`
}

type SearchSlotsEnvelope struct {
	Success bool                `json:"success" example:"true"`
	Data    SearchSlotsResponse `json:"data"`
	Meta    ResponseMetaDTO     `json:"meta"`
}

type MessageData struct {
	Message string `json:"message" example:"slot selected successfully"`
}

type MessageEnvelope struct {
	Success bool            `json:"success" example:"true"`
	Data    MessageData     `json:"data"`
	Meta    ResponseMetaDTO `json:"meta"`
}

type ResponseMetaDTO struct {
	RequestID string `json:"request_id" example:"e7cc5d4a-3d85-4938-bff3-d2a6cf8ca2bb"`
}

func (r SearchSlotsRequest) ToCommand(meetingID, organizerUserID uuid.UUID) SearchSlotsCommand {
	return SearchSlotsCommand{
		MeetingID:       meetingID,
		OrganizerUserID: organizerUserID,
		TopN:            r.TopN,
	}
}

func (r SelectSlotRequest) ToCommand(meetingID, organizerUserID uuid.UUID) (SelectSlotCommand, error) {
	slotID, err := uuid.Parse(r.MeetingSlotID)
	if err != nil {
		return SelectSlotCommand{}, err
	}
	return SelectSlotCommand{
		MeetingID:       meetingID,
		OrganizerUserID: organizerUserID,
		MeetingSlotID:   slotID,
	}, nil
}
