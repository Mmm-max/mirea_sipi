package users

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Email        string
	FullName     string
	Timezone     string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type WorkingHours struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	Weekday      int
	StartTime    string
	EndTime      string
	IsWorkingDay bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type UnavailabilityType string

const (
	UnavailabilityVacation     UnavailabilityType = "vacation"
	UnavailabilitySickLeave    UnavailabilityType = "sick_leave"
	UnavailabilityBusinessTrip UnavailabilityType = "business_trip"
	UnavailabilityCustom       UnavailabilityType = "custom"
)

type UnavailabilityPeriod struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Type      UnavailabilityType
	Title     string
	StartAt   time.Time
	EndAt     time.Time
	Comment   string
	CreatedAt time.Time
	UpdatedAt time.Time
}
