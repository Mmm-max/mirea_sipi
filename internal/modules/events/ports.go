package events

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type RepositoryPort interface {
	Create(ctx context.Context, event *Event) error
	GetByID(ctx context.Context, id uuid.UUID) (*Event, error)
	GetImportedByExternalUID(ctx context.Context, ownerUserID uuid.UUID, externalUID string) (*Event, error)
	ListByOwner(ctx context.Context, ownerUserID uuid.UUID, dateFrom, dateTo *time.Time) ([]Event, error)
	Update(ctx context.Context, event *Event) error
	Delete(ctx context.Context, id uuid.UUID) error
	HasOverlap(ctx context.Context, ownerUserID uuid.UUID, startAt, endAt time.Time, excludeID *uuid.UUID) (bool, error)
}

type ServicePort interface {
	Create(ctx context.Context, command CreateEventCommand) (*Event, error)
	List(ctx context.Context, query ListEventsQuery) ([]Event, error)
	Get(ctx context.Context, query GetEventQuery) (*Event, error)
	Update(ctx context.Context, command UpdateEventCommand) (*Event, error)
	Delete(ctx context.Context, command DeleteEventCommand) error
}
