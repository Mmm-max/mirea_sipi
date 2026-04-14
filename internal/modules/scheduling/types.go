package scheduling

import (
	"time"

	"github.com/google/uuid"
)

type MeetingAggregate struct {
	ID                uuid.UUID
	OrganizerUserID   uuid.UUID
	Title             string
	DurationMinutes   int
	SearchRangeStart  time.Time
	SearchRangeEnd    time.Time
	EarliestStartTime *string
	LatestStartTime   *string
	SelectedStartAt   *time.Time
	SelectedEndAt     *time.Time
	Participants      []MeetingParticipantRef
	Resources         []MeetingResourceRef
}

type MeetingParticipantRef struct {
	UserID             uuid.UUID
	VisibilityOverride *string
}

type MeetingResourceRef struct {
	ResourceID uuid.UUID
}

type WorkingHoursWindow struct {
	Weekday      int
	StartTime    string
	EndTime      string
	IsWorkingDay bool
}

type UnavailabilityWindow struct {
	Title   string
	StartAt time.Time
	EndAt   time.Time
}

type CalendarEvent struct {
	ID              uuid.UUID
	OwnerUserID     uuid.UUID
	Title           string
	StartAt         time.Time
	EndAt           time.Time
	Priority        string
	IsReschedulable bool
	VisibilityHint  string
}
