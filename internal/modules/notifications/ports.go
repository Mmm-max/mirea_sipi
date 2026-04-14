package notifications

import (
	"context"

	"github.com/google/uuid"
)

type RepositoryPort interface {
	CreateMany(ctx context.Context, items []Notification) error
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]Notification, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Notification, error)
	MarkRead(ctx context.Context, id uuid.UUID) error
	MarkAllReadByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
}

type ServicePort interface {
	CreateMany(ctx context.Context, items []CreateNotificationInput) error
	ListNotifications(ctx context.Context, query ListNotificationsQuery) ([]Notification, error)
	MarkRead(ctx context.Context, command MarkReadCommand) error
	MarkAllRead(ctx context.Context, command MarkAllReadCommand) (int64, error)
}
