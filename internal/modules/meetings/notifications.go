package meetings

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"sipi/internal/modules/notifications"

	"github.com/google/uuid"
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

func (h *NotificationHook) NotifyParticipantsInvited(ctx context.Context, meeting *Meeting, participantIDs []uuid.UUID) {
	if h == nil || h.service == nil || meeting == nil || len(participantIDs) == 0 {
		return
	}
	items := make([]notifications.CreateNotificationInput, 0, len(participantIDs))
	for _, userID := range participantIDs {
		metadata := mustMarshalMetadata(map[string]any{
			"meeting_id": meeting.ID.String(),
			"type":       string(notifications.NotificationTypeMeetingInvitation),
		})
		items = append(items, notifications.CreateNotificationInput{
			UserID:       userID,
			Type:         notifications.NotificationTypeMeetingInvitation,
			Title:        "New meeting invitation",
			Body:         fmt.Sprintf("You were invited to meeting \"%s\"", meeting.Title),
			MetadataJSON: metadata,
		})
	}
	_ = h.service.CreateMany(ctx, items)
}

func (h *NotificationHook) NotifyMeetingUpdated(ctx context.Context, meeting *Meeting) {
	if h == nil || h.service == nil || meeting == nil {
		return
	}
	userIDs := participantUserIDs(meeting, true)
	items := make([]notifications.CreateNotificationInput, 0, len(userIDs))
	for _, userID := range userIDs {
		metadata := mustMarshalMetadata(map[string]any{
			"meeting_id": meeting.ID.String(),
			"type":       string(notifications.NotificationTypeMeetingUpdated),
		})
		items = append(items, notifications.CreateNotificationInput{
			UserID:       userID,
			Type:         notifications.NotificationTypeMeetingUpdated,
			Title:        "Meeting updated",
			Body:         fmt.Sprintf("Meeting \"%s\" was updated", meeting.Title),
			MetadataJSON: metadata,
		})
	}
	_ = h.service.CreateMany(ctx, items)
}

func (h *NotificationHook) NotifyAlternativeRequested(ctx context.Context, meeting *Meeting, participant *MeetingParticipant) {
	if h == nil || h.service == nil || meeting == nil || participant == nil {
		return
	}
	metadata := mustMarshalMetadata(map[string]any{
		"meeting_id":        meeting.ID.String(),
		"participant_id":    participant.UserID.String(),
		"proposed_start_at": timePointerString(participant.AlternativeStartAt),
		"proposed_end_at":   timePointerString(participant.AlternativeEndAt),
		"type":              string(notifications.NotificationTypeMeetingAlternativeRequest),
	})
	_ = h.service.CreateMany(ctx, []notifications.CreateNotificationInput{{
		UserID:       meeting.OrganizerUserID,
		Type:         notifications.NotificationTypeMeetingAlternativeRequest,
		Title:        "Alternative time requested",
		Body:         fmt.Sprintf("A participant requested an alternative time for \"%s\"", meeting.Title),
		MetadataJSON: metadata,
	}})
}

func (h *NotificationHook) NotifyResourceConflict(ctx context.Context, meeting *Meeting, resourceID uuid.UUID, startAt, endAt time.Time) {
	if h == nil || h.service == nil || meeting == nil {
		return
	}
	metadata := mustMarshalMetadata(map[string]any{
		"meeting_id":  meeting.ID.String(),
		"resource_id": resourceID.String(),
		"start_at":    startAt.UTC().Format(time.RFC3339),
		"end_at":      endAt.UTC().Format(time.RFC3339),
		"type":        string(notifications.NotificationTypeResourceConflict),
	})
	_ = h.service.CreateMany(ctx, []notifications.CreateNotificationInput{{
		UserID:       meeting.OrganizerUserID,
		Type:         notifications.NotificationTypeResourceConflict,
		Title:        "Resource conflict detected",
		Body:         fmt.Sprintf("Resource conflict detected for meeting \"%s\"", meeting.Title),
		MetadataJSON: metadata,
	}})
}

func (NoopNotificationHook) NotifyParticipantsInvited(context.Context, *Meeting, []uuid.UUID) {}

func (NoopNotificationHook) NotifyMeetingUpdated(context.Context, *Meeting) {}

func (NoopNotificationHook) NotifyAlternativeRequested(context.Context, *Meeting, *MeetingParticipant) {
}

func (NoopNotificationHook) NotifyResourceConflict(context.Context, *Meeting, uuid.UUID, time.Time, time.Time) {
}

func participantUserIDs(meeting *Meeting, excludeOrganizer bool) []uuid.UUID {
	if meeting == nil {
		return nil
	}
	result := make([]uuid.UUID, 0, len(meeting.Participants))
	seen := make(map[uuid.UUID]struct{}, len(meeting.Participants))
	for _, participant := range meeting.Participants {
		if excludeOrganizer && participant.UserID == meeting.OrganizerUserID {
			continue
		}
		if _, ok := seen[participant.UserID]; ok {
			continue
		}
		seen[participant.UserID] = struct{}{}
		result = append(result, participant.UserID)
	}
	return result
}

func mustMarshalMetadata(value map[string]any) json.RawMessage {
	if len(value) == 0 {
		return nil
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	return raw
}

func timePointerString(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC().Format(time.RFC3339)
}
