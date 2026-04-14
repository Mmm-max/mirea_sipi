package scheduling

import (
	"time"

	"github.com/google/uuid"
)

type MeetingSlot struct {
	ID        uuid.UUID
	MeetingID uuid.UUID
	StartAt   time.Time
	EndAt     time.Time
	Score     int
	Rank      int
	Conflicts []SlotConflict
	CreatedAt time.Time
	UpdatedAt time.Time
}

type SlotConflict struct {
	ID              uuid.UUID
	MeetingSlotID   uuid.UUID
	UserID          *uuid.UUID
	EventID         *uuid.UUID
	ResourceID      *uuid.UUID
	ConflictType    string
	VisibleTitle    string
	VisiblePriority *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type SearchResult struct {
	MeetingID uuid.UUID
	Slots     []MeetingSlot
}
