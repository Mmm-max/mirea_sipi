package users

import (
	"time"

	"github.com/google/uuid"
)

type GetProfileQuery struct {
	UserID uuid.UUID
}

type UpdateProfileCommand struct {
	UserID   uuid.UUID
	FullName string
	Timezone string
}

type WorkingHoursInput struct {
	Weekday      int
	StartTime    string
	EndTime      string
	IsWorkingDay bool
}

type ReplaceWorkingHoursCommand struct {
	UserID uuid.UUID
	Items  []WorkingHoursInput
}

type ListWorkingHoursQuery struct {
	UserID uuid.UUID
}

type CreateUnavailabilityCommand struct {
	UserID  uuid.UUID
	Type    UnavailabilityType
	Title   string
	StartAt time.Time
	EndAt   time.Time
	Comment string
}

type ListUnavailabilityQuery struct {
	UserID uuid.UUID
}

type DeleteUnavailabilityCommand struct {
	UserID uuid.UUID
	ID     uuid.UUID
}

type ProfileResult struct {
	ID        uuid.UUID
	Email     string
	FullName  string
	Timezone  string
	CreatedAt time.Time
	UpdatedAt time.Time
}
