package meetings

import (
	"time"

	"github.com/google/uuid"
)

type MeetingStatus string

const (
	MeetingStatusDraft     MeetingStatus = "draft"
	MeetingStatusScheduled MeetingStatus = "scheduled"
	MeetingStatusCancelled MeetingStatus = "cancelled"
)

type InvitationStatus string

const (
	InvitationStatusPending   InvitationStatus = "pending"
	InvitationStatusAccepted  InvitationStatus = "accepted"
	InvitationStatusDeclined  InvitationStatus = "declined"
	InvitationStatusTentative InvitationStatus = "tentative"
)

type Meeting struct {
	ID                uuid.UUID
	OrganizerUserID   uuid.UUID
	Title             string
	Description       string
	DurationMinutes   int
	Priority          string
	Status            MeetingStatus
	SearchRangeStart  time.Time
	SearchRangeEnd    time.Time
	EarliestStartTime *string
	LatestStartTime   *string
	RecurrenceRule    *string
	SelectedStartAt   *time.Time
	SelectedEndAt     *time.Time
	Participants      []MeetingParticipant
	Resources         []MeetingResource
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
}

type MeetingParticipant struct {
	ID                 uuid.UUID
	MeetingID          uuid.UUID
	UserID             uuid.UUID
	ResponseStatus     InvitationStatus
	VisibilityOverride *string
	AlternativeStartAt *time.Time
	AlternativeEndAt   *time.Time
	AlternativeComment *string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type MeetingResource struct {
	ID         uuid.UUID
	MeetingID  uuid.UUID
	ResourceID uuid.UUID
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
