package scheduling

import (
	"context"
	"errors"
	"time"

	"sipi/internal/modules/events"
	"sipi/internal/modules/resources"
	"sipi/internal/modules/users"

	"github.com/google/uuid"
)

type workingHoursReader interface {
	ListWorkingHours(ctx context.Context, userID uuid.UUID) ([]users.WorkingHours, error)
	ListUnavailabilityPeriods(ctx context.Context, userID uuid.UUID) ([]users.UnavailabilityPeriod, error)
}

type eventReader interface {
	ListByOwner(ctx context.Context, ownerUserID uuid.UUID, dateFrom, dateTo *time.Time) ([]events.Event, error)
}

type resourceAvailabilityReader interface {
	GetResourceByID(ctx context.Context, id uuid.UUID) (*resources.Resource, error)
	HasBookingOverlap(ctx context.Context, resourceID uuid.UUID, startAt, endAt time.Time, excludeID *uuid.UUID) (bool, error)
}

type AvailabilityStore struct {
	users     workingHoursReader
	events    eventReader
	resources resourceAvailabilityReader
}

func NewAvailabilityStore(usersRepository workingHoursReader, eventsRepository eventReader, resourcesRepository resourceAvailabilityReader) *AvailabilityStore {
	return &AvailabilityStore{
		users:     usersRepository,
		events:    eventsRepository,
		resources: resourcesRepository,
	}
}

func (s *AvailabilityStore) ListWorkingHours(ctx context.Context, userID uuid.UUID) ([]WorkingHoursWindow, error) {
	items, err := s.users.ListWorkingHours(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := make([]WorkingHoursWindow, 0, len(items))
	for _, item := range items {
		result = append(result, WorkingHoursWindow{
			Weekday:      item.Weekday,
			StartTime:    item.StartTime,
			EndTime:      item.EndTime,
			IsWorkingDay: item.IsWorkingDay,
		})
	}
	return result, nil
}

func (s *AvailabilityStore) ListUnavailability(ctx context.Context, userID uuid.UUID) ([]UnavailabilityWindow, error) {
	items, err := s.users.ListUnavailabilityPeriods(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := make([]UnavailabilityWindow, 0, len(items))
	for _, item := range items {
		result = append(result, UnavailabilityWindow{
			Title:   item.Title,
			StartAt: item.StartAt,
			EndAt:   item.EndAt,
		})
	}
	return result, nil
}

func (s *AvailabilityStore) ListEvents(ctx context.Context, userID uuid.UUID, dateFrom, dateTo time.Time) ([]CalendarEvent, error) {
	items, err := s.events.ListByOwner(ctx, userID, &dateFrom, &dateTo)
	if err != nil {
		return nil, err
	}
	result := make([]CalendarEvent, 0, len(items))
	for _, item := range items {
		result = append(result, CalendarEvent{
			ID:              item.ID,
			OwnerUserID:     item.OwnerUserID,
			Title:           item.Title,
			StartAt:         item.StartAt,
			EndAt:           item.EndAt,
			Priority:        string(item.Priority),
			IsReschedulable: item.IsReschedulable,
			VisibilityHint:  item.VisibilityHint,
		})
	}
	return result, nil
}

func (s *AvailabilityStore) ResourceAvailable(ctx context.Context, resourceID uuid.UUID, startAt, endAt time.Time) (bool, error) {
	_, err := s.resources.GetResourceByID(ctx, resourceID)
	if err != nil {
		if errors.Is(err, resources.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	hasOverlap, err := s.resources.HasBookingOverlap(ctx, resourceID, startAt, endAt, nil)
	if err != nil {
		return false, err
	}
	return !hasOverlap, nil
}
