package resources

import (
	"time"

	"github.com/google/uuid"
)

type ResourceType string

const (
	ResourceTypeMeetingRoom ResourceType = "meeting_room"
	ResourceTypeEquipment   ResourceType = "equipment"
	ResourceTypeSharedSpace ResourceType = "shared_space"
	ResourceTypeOther       ResourceType = "other"
)

type Resource struct {
	ID          uuid.UUID
	Name        string
	Type        ResourceType
	Description string
	Capacity    *int
	Location    *string
	OwnerUserID *uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

type ResourceBooking struct {
	ID             uuid.UUID
	ResourceID     uuid.UUID
	EventID        *uuid.UUID
	BookedByUserID uuid.UUID
	StartAt        time.Time
	EndAt          time.Time
	Title          string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
}
