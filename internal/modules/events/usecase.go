package events

import (
	"time"

	"github.com/google/uuid"
)

type CreateEventCommand struct {
	OwnerUserID     uuid.UUID
	SourceType      SourceType
	ExternalUID     *string
	Title           string
	Description     string
	StartAt         time.Time
	EndAt           time.Time
	Priority        Priority
	IsReschedulable bool
	VisibilityHint  string
}

type UpdateEventCommand struct {
	ID               uuid.UUID
	OwnerUserID      uuid.UUID
	SourceType       *SourceType
	ExternalUID      *string
	ExternalUIDSet   bool
	ClearExternalUID bool
	Title            *string
	Description      *string
	StartAt          *time.Time
	EndAt            *time.Time
	Priority         *Priority
	IsReschedulable  *bool
	VisibilityHint   *string
}

type DeleteEventCommand struct {
	ID          uuid.UUID
	OwnerUserID uuid.UUID
}

type GetEventQuery struct {
	ID          uuid.UUID
	OwnerUserID uuid.UUID
}

type ListEventsQuery struct {
	OwnerUserID uuid.UUID
	DateFrom    *time.Time
	DateTo      *time.Time
}
