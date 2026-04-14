package scheduling

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type RepositoryPort interface {
	ReplaceSlots(ctx context.Context, meetingID uuid.UUID, slots []MeetingSlot) error
	ListSlotsByMeetingID(ctx context.Context, meetingID uuid.UUID) ([]MeetingSlot, error)
	GetSlotByID(ctx context.Context, id uuid.UUID) (*MeetingSlot, error)
}

type MeetingStorePort interface {
	GetMeeting(ctx context.Context, id uuid.UUID) (*MeetingAggregate, error)
	UpdateSelectedTime(ctx context.Context, meetingID uuid.UUID, startAt, endAt time.Time) error
}

type AvailabilityStorePort interface {
	ListWorkingHours(ctx context.Context, userID uuid.UUID) ([]WorkingHoursWindow, error)
	ListUnavailability(ctx context.Context, userID uuid.UUID) ([]UnavailabilityWindow, error)
	ListEvents(ctx context.Context, userID uuid.UUID, dateFrom, dateTo time.Time) ([]CalendarEvent, error)
	ResourceAvailable(ctx context.Context, resourceID uuid.UUID, startAt, endAt time.Time) (bool, error)
}

type EventSyncPort interface {
	SyncSelectedMeeting(ctx context.Context, meeting *MeetingAggregate, startAt, endAt time.Time) error
}

type NotificationPort interface {
	NotifySlotSelected(ctx context.Context, meeting *MeetingAggregate, startAt, endAt time.Time)
}

type ServicePort interface {
	SearchSlots(ctx context.Context, command SearchSlotsCommand) (*SearchResult, error)
	GetSlots(ctx context.Context, query GetSlotsQuery) (*SearchResult, error)
	SelectSlot(ctx context.Context, command SelectSlotCommand) (*MeetingAggregate, error)
}
