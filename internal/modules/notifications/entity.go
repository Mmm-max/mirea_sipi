package notifications

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type NotificationType string

const (
	NotificationTypeMeetingInvitation         NotificationType = "meeting_invitation"
	NotificationTypeMeetingSlotSelected       NotificationType = "meeting_slot_selected"
	NotificationTypeMeetingAlternativeRequest NotificationType = "meeting_alternative_request"
	NotificationTypeMeetingUpdated            NotificationType = "meeting_updated"
	NotificationTypeResourceConflict          NotificationType = "resource_conflict"
)

type Notification struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	Type         NotificationType
	Title        string
	Body         string
	IsRead       bool
	MetadataJSON json.RawMessage
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
