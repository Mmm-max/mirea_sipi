package notifications

import (
	"encoding/json"

	"github.com/google/uuid"
)

type CreateNotificationInput struct {
	UserID       uuid.UUID
	Type         NotificationType
	Title        string
	Body         string
	MetadataJSON json.RawMessage
}

type ListNotificationsQuery struct {
	UserID uuid.UUID
}

type MarkReadCommand struct {
	UserID         uuid.UUID
	NotificationID uuid.UUID
}

type MarkAllReadCommand struct {
	UserID uuid.UUID
}
