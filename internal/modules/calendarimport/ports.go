package calendarimport

import (
	"context"

	"github.com/google/uuid"
)

type RepositoryPort interface {
	Create(ctx context.Context, job *ImportJob) error
	Update(ctx context.Context, job *ImportJob) error
	GetByID(ctx context.Context, id uuid.UUID) (*ImportJob, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]ImportJob, error)
}

type ParserPort interface {
	Parse(content []byte) (*ParseResult, error)
}

type EventStorePort interface {
	UpsertImportedEvent(ctx context.Context, event ImportedEvent) (bool, error)
}

type ServicePort interface {
	ImportICS(ctx context.Context, command ImportICSCommand) (*ImportJob, error)
	ListHistory(ctx context.Context, query ListImportJobsQuery) ([]ImportJob, error)
	GetHistoryByID(ctx context.Context, query GetImportJobQuery) (*ImportJob, error)
}
