package meetings

import (
	"time"

	"github.com/google/uuid"
)

type CreateMeetingCommand struct {
	OrganizerUserID   uuid.UUID
	Title             string
	Description       string
	DurationMinutes   int
	Priority          string
	Status            *MeetingStatus
	SearchRangeStart  time.Time
	SearchRangeEnd    time.Time
	EarliestStartTime *string
	LatestStartTime   *string
	RecurrenceRule    *string
	SelectedStartAt   *time.Time
	SelectedEndAt     *time.Time
}

type UpdateMeetingCommand struct {
	ID                uuid.UUID
	OrganizerUserID   uuid.UUID
	Title             *string
	Description       *string
	DurationMinutes   *int
	Priority          *string
	Status            *MeetingStatus
	SearchRangeStart  *time.Time
	SearchRangeEnd    *time.Time
	EarliestStartTime *string
	EarliestSet       bool
	ClearEarliest     bool
	LatestStartTime   *string
	LatestSet         bool
	ClearLatest       bool
	RecurrenceRule    *string
	RecurrenceSet     bool
	ClearRecurrence   bool
	SelectedStartAt   *time.Time
	SelectedEndAt     *time.Time
	SelectedSet       bool
	ClearSelected     bool
}

type GetMeetingQuery struct {
	ID     uuid.UUID
	UserID uuid.UUID
}

type ListMeetingsQuery struct {
	UserID uuid.UUID
}

type AddParticipantsCommand struct {
	MeetingID       uuid.UUID
	OrganizerUserID uuid.UUID
	ParticipantIDs  []uuid.UUID
}

type RemoveParticipantCommand struct {
	MeetingID         uuid.UUID
	ActingUserID      uuid.UUID
	ParticipantUserID uuid.UUID
}

type AddResourceCommand struct {
	MeetingID       uuid.UUID
	OrganizerUserID uuid.UUID
	ResourceIDs     []uuid.UUID
}

type RemoveResourceCommand struct {
	MeetingID       uuid.UUID
	OrganizerUserID uuid.UUID
	ResourceID      uuid.UUID
}

type RespondInvitationCommand struct {
	MeetingID      uuid.UUID
	UserID         uuid.UUID
	ResponseStatus InvitationStatus
}

type RequestAlternativeCommand struct {
	MeetingID       uuid.UUID
	UserID          uuid.UUID
	ProposedStartAt time.Time
	ProposedEndAt   time.Time
	Comment         *string
}
