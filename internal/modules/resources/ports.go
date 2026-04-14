package resources

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type RepositoryPort interface {
	CreateResource(ctx context.Context, resource *Resource) error
	GetResourceByID(ctx context.Context, id uuid.UUID) (*Resource, error)
	ListResources(ctx context.Context) ([]Resource, error)
	UpdateResource(ctx context.Context, resource *Resource) error
	DeleteResource(ctx context.Context, id uuid.UUID) error
	CreateBooking(ctx context.Context, booking *ResourceBooking) error
	GetBookingByID(ctx context.Context, id uuid.UUID) (*ResourceBooking, error)
	DeleteBooking(ctx context.Context, id uuid.UUID) error
	ListBookingsByResource(ctx context.Context, resourceID uuid.UUID, dateFrom, dateTo time.Time) ([]ResourceBooking, error)
	HasBookingOverlap(ctx context.Context, resourceID uuid.UUID, startAt, endAt time.Time, excludeID *uuid.UUID) (bool, error)
}

type ServicePort interface {
	CreateResource(ctx context.Context, command CreateResourceCommand) (*Resource, error)
	ListResources(ctx context.Context, query ListResourcesQuery) ([]Resource, error)
	GetResource(ctx context.Context, query GetResourceQuery) (*Resource, error)
	UpdateResource(ctx context.Context, command UpdateResourceCommand) (*Resource, error)
	DeleteResource(ctx context.Context, id uuid.UUID) error
	GetAvailability(ctx context.Context, query GetResourceAvailabilityQuery) (*ResourceAvailability, error)
	CreateBooking(ctx context.Context, command CreateBookingCommand) (*ResourceBooking, error)
	CancelBooking(ctx context.Context, command CancelBookingCommand) error
}
