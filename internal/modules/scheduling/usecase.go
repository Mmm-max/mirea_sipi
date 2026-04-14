package scheduling

import "github.com/google/uuid"

type SearchSlotsCommand struct {
	MeetingID       uuid.UUID
	OrganizerUserID uuid.UUID
	TopN            int
}

type GetSlotsQuery struct {
	MeetingID       uuid.UUID
	RequesterUserID uuid.UUID
}

type SelectSlotCommand struct {
	MeetingID       uuid.UUID
	OrganizerUserID uuid.UUID
	MeetingSlotID   uuid.UUID
}
