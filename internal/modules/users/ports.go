package users

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type RepositoryPort interface {
	GetUserByID(ctx context.Context, userID uuid.UUID) (*User, error)
	UpdateUserProfile(ctx context.Context, user *User) error
	ReplaceWorkingHours(ctx context.Context, userID uuid.UUID, items []WorkingHours) error
	ListWorkingHours(ctx context.Context, userID uuid.UUID) ([]WorkingHours, error)
	CreateUnavailabilityPeriod(ctx context.Context, period *UnavailabilityPeriod) error
	ListUnavailabilityPeriods(ctx context.Context, userID uuid.UUID) ([]UnavailabilityPeriod, error)
	GetUnavailabilityPeriodByID(ctx context.Context, id uuid.UUID) (*UnavailabilityPeriod, error)
	DeleteUnavailabilityPeriod(ctx context.Context, id uuid.UUID) error
	HasUnavailabilityOverlap(ctx context.Context, userID uuid.UUID, startAt, endAt time.Time) (bool, error)
}

type ServicePort interface {
	GetProfile(ctx context.Context, query GetProfileQuery) (*ProfileResult, error)
	UpdateProfile(ctx context.Context, command UpdateProfileCommand) (*ProfileResult, error)
	ReplaceWorkingHours(ctx context.Context, command ReplaceWorkingHoursCommand) ([]WorkingHours, error)
	ListWorkingHours(ctx context.Context, query ListWorkingHoursQuery) ([]WorkingHours, error)
	CreateUnavailabilityPeriod(ctx context.Context, command CreateUnavailabilityCommand) (*UnavailabilityPeriod, error)
	ListUnavailabilityPeriods(ctx context.Context, query ListUnavailabilityQuery) ([]UnavailabilityPeriod, error)
	DeleteUnavailabilityPeriod(ctx context.Context, command DeleteUnavailabilityCommand) error
}
