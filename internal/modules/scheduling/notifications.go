package scheduling

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"sipi/internal/modules/notifications"
)

type notificationCreator interface {
	CreateMany(ctx context.Context, items []notifications.CreateNotificationInput) error
}

type NotificationHook struct {
	service notificationCreator
}

func NewNotificationHook(service notificationCreator) *NotificationHook {
	return &NotificationHook{service: service}
}

type NoopNotificationHook struct{}

func NewNoopNotificationHook() *NoopNotificationHook {
	return &NoopNotificationHook{}
}

func (h *NotificationHook) NotifySlotSelected(ctx context.Context, meeting *MeetingAggregate, startAt, endAt time.Time) {
	if h == nil || h.service == nil || meeting == nil {
		return
	}
	items := make([]notifications.CreateNotificationInput, 0, len(meetingParticipants(meeting)))
	for _, participant := range meetingParticipants(meeting) {
		if participant.UserID == meeting.OrganizerUserID {
			continue
		}
		metadata := marshalSchedulingMetadata(map[string]any{
			"meeting_id": meeting.ID.String(),
			"start_at":   startAt.UTC().Format(time.RFC3339),
			"end_at":     endAt.UTC().Format(time.RFC3339),
			"type":       string(notifications.NotificationTypeMeetingSlotSelected),
		})
		items = append(items, notifications.CreateNotificationInput{
			UserID:       participant.UserID,
			Type:         notifications.NotificationTypeMeetingSlotSelected,
			Title:        "Meeting slot selected",
			Body:         fmt.Sprintf("A time slot was selected for meeting \"%s\"", meeting.Title),
			MetadataJSON: metadata,
		})
	}
	_ = h.service.CreateMany(ctx, items)
}

func marshalSchedulingMetadata(value map[string]any) json.RawMessage {
	if len(value) == 0 {
		return nil
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	return raw
}

func (NoopNotificationHook) NotifySlotSelected(context.Context, *MeetingAggregate, time.Time, time.Time) {
}
