package resources

import (
	"time"

	"github.com/google/uuid"
)

type CreateResourceCommand struct {
	Name        string
	Type        ResourceType
	Description string
	Capacity    *int
	Location    *string
	OwnerUserID *uuid.UUID
}

type UpdateResourceCommand struct {
	ID            uuid.UUID
	Name          *string
	Type          *ResourceType
	Description   *string
	Capacity      *int
	CapacitySet   bool
	ClearCapacity bool
	Location      *string
	LocationSet   bool
	ClearLocation bool
}

type GetResourceQuery struct {
	ID uuid.UUID
}

type ListResourcesQuery struct{}

type GetResourceAvailabilityQuery struct {
	ResourceID uuid.UUID
	DateFrom   time.Time
	DateTo     time.Time
}

type CreateBookingCommand struct {
	ResourceID     uuid.UUID
	BookedByUserID uuid.UUID
	EventID        *uuid.UUID
	StartAt        time.Time
	EndAt          time.Time
	Title          string
}

type CancelBookingCommand struct {
	ResourceID uuid.UUID
	BookingID  uuid.UUID
	UserID     uuid.UUID
}

type ResourceAvailability struct {
	ResourceID  uuid.UUID
	DateFrom    time.Time
	DateTo      time.Time
	IsAvailable bool
	Bookings    []ResourceBooking
}
