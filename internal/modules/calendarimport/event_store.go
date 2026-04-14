package calendarimport

import (
	"context"
	"errors"

	"sipi/internal/modules/events"
)

type EventStore struct {
	repository events.RepositoryPort
}

func NewEventStore(repository events.RepositoryPort) *EventStore {
	return &EventStore{repository: repository}
}

func (s *EventStore) UpsertImportedEvent(ctx context.Context, event ImportedEvent) (bool, error) {
	existing, err := s.repository.GetImportedByExternalUID(ctx, event.OwnerUserID, event.ExternalUID)
	if err != nil && !errors.Is(err, events.ErrNotFound) {
		return false, err
	}

	if errors.Is(err, events.ErrNotFound) {
		model := &events.Event{
			OwnerUserID:     event.OwnerUserID,
			SourceType:      events.SourceTypeImported,
			ExternalUID:     stringPointer(event.ExternalUID),
			Title:           event.Title,
			Description:     event.Description,
			StartAt:         event.StartAt,
			EndAt:           event.EndAt,
			Priority:        events.Priority(event.Priority),
			IsReschedulable: event.IsReschedulable,
			VisibilityHint:  event.VisibilityHint,
		}
		if createErr := s.repository.Create(ctx, model); createErr != nil {
			return false, createErr
		}
		return true, nil
	}

	existing.Title = event.Title
	existing.Description = event.Description
	existing.StartAt = event.StartAt
	existing.EndAt = event.EndAt
	existing.Priority = events.Priority(event.Priority)
	existing.IsReschedulable = event.IsReschedulable
	existing.VisibilityHint = event.VisibilityHint
	existing.ExternalUID = stringPointer(event.ExternalUID)
	existing.SourceType = events.SourceTypeImported

	if updateErr := s.repository.Update(ctx, existing); updateErr != nil {
		return false, updateErr
	}

	return false, nil
}

func stringPointer(value string) *string {
	if value == "" {
		return nil
	}

	copyValue := value
	return &copyValue
}
