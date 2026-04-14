package events

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
)

type Service struct {
	repository RepositoryPort
}

func NewService(repository RepositoryPort) *Service {
	return &Service{repository: repository}
}

func (s *Service) Create(ctx context.Context, command CreateEventCommand) (*Event, error) {
	if err := validateCreateCommand(command); err != nil {
		return nil, err
	}

	hasOverlap, err := s.repository.HasOverlap(ctx, command.OwnerUserID, command.StartAt, command.EndAt, nil)
	if err != nil {
		return nil, newServiceError(ErrCodeInternal, "failed to validate event overlap", err)
	}

	if hasOverlap {
		return nil, newServiceError(ErrCodeConflict, "event overlaps with an existing event", nil)
	}

	event := &Event{
		ID:              uuid.New(),
		OwnerUserID:     command.OwnerUserID,
		SourceType:      command.SourceType,
		ExternalUID:     normalizeOptionalString(command.ExternalUID),
		Title:           strings.TrimSpace(command.Title),
		Description:     strings.TrimSpace(command.Description),
		StartAt:         command.StartAt.UTC(),
		EndAt:           command.EndAt.UTC(),
		Priority:        command.Priority,
		IsReschedulable: command.IsReschedulable,
		VisibilityHint:  normalizeVisibilityHint(command.VisibilityHint),
	}

	if err := s.repository.Create(ctx, event); err != nil {
		return nil, newServiceError(ErrCodeInternal, "failed to create event", err)
	}

	return event, nil
}

func (s *Service) List(ctx context.Context, query ListEventsQuery) ([]Event, error) {
	if query.DateFrom != nil && query.DateTo != nil && query.DateFrom.After(*query.DateTo) {
		return nil, newServiceError(ErrCodeValidation, "date_from must be before or equal to date_to", nil)
	}

	events, err := s.repository.ListByOwner(ctx, query.OwnerUserID, query.DateFrom, query.DateTo)
	if err != nil {
		return nil, newServiceError(ErrCodeInternal, "failed to list events", err)
	}

	return events, nil
}

func (s *Service) Get(ctx context.Context, query GetEventQuery) (*Event, error) {
	event, err := s.repository.GetByID(ctx, query.ID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, newServiceError(ErrCodeNotFound, "event not found", err)
		}
		return nil, newServiceError(ErrCodeInternal, "failed to get event", err)
	}

	if event.OwnerUserID != query.OwnerUserID {
		return nil, newServiceError(ErrCodeNotFound, "event not found", nil)
	}

	return event, nil
}

func (s *Service) Update(ctx context.Context, command UpdateEventCommand) (*Event, error) {
	event, err := s.repository.GetByID(ctx, command.ID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, newServiceError(ErrCodeNotFound, "event not found", err)
		}
		return nil, newServiceError(ErrCodeInternal, "failed to update event", err)
	}

	if event.OwnerUserID != command.OwnerUserID {
		return nil, newServiceError(ErrCodeNotFound, "event not found", nil)
	}

	applyUpdate(event, command)
	if err := validateEvent(event); err != nil {
		return nil, err
	}

	hasOverlap, err := s.repository.HasOverlap(ctx, event.OwnerUserID, event.StartAt, event.EndAt, &event.ID)
	if err != nil {
		return nil, newServiceError(ErrCodeInternal, "failed to validate event overlap", err)
	}

	if hasOverlap {
		return nil, newServiceError(ErrCodeConflict, "event overlaps with an existing event", nil)
	}

	if err := s.repository.Update(ctx, event); err != nil {
		return nil, newServiceError(ErrCodeInternal, "failed to update event", err)
	}

	return event, nil
}

func (s *Service) Delete(ctx context.Context, command DeleteEventCommand) error {
	event, err := s.repository.GetByID(ctx, command.ID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return newServiceError(ErrCodeNotFound, "event not found", err)
		}
		return newServiceError(ErrCodeInternal, "failed to delete event", err)
	}

	if event.OwnerUserID != command.OwnerUserID {
		return newServiceError(ErrCodeNotFound, "event not found", nil)
	}

	if err := s.repository.Delete(ctx, command.ID); err != nil {
		return newServiceError(ErrCodeInternal, "failed to delete event", err)
	}

	return nil
}

func validateCreateCommand(command CreateEventCommand) error {
	event := &Event{
		SourceType:      command.SourceType,
		ExternalUID:     command.ExternalUID,
		Title:           command.Title,
		Description:     command.Description,
		StartAt:         command.StartAt,
		EndAt:           command.EndAt,
		Priority:        command.Priority,
		IsReschedulable: command.IsReschedulable,
		VisibilityHint:  command.VisibilityHint,
	}

	return validateEvent(event)
}

func validateEvent(event *Event) error {
	if event == nil {
		return newServiceError(ErrCodeValidation, "event is required", nil)
	}
	if !isValidSourceType(event.SourceType) {
		return newServiceError(ErrCodeValidation, "invalid source_type", nil)
	}
	if !isValidPriority(event.Priority) {
		return newServiceError(ErrCodeValidation, "invalid priority", nil)
	}
	if strings.TrimSpace(event.Title) == "" {
		return newServiceError(ErrCodeValidation, "title is required", nil)
	}
	if event.StartAt.IsZero() || event.EndAt.IsZero() || !event.StartAt.Before(event.EndAt) {
		return newServiceError(ErrCodeValidation, "start_at must be before end_at", nil)
	}

	return nil
}

func applyUpdate(event *Event, command UpdateEventCommand) {
	if command.SourceType != nil {
		event.SourceType = *command.SourceType
	}
	if command.ExternalUIDSet {
		if command.ClearExternalUID {
			event.ExternalUID = nil
		} else {
			event.ExternalUID = normalizeOptionalString(command.ExternalUID)
		}
	}
	if command.Title != nil {
		event.Title = strings.TrimSpace(*command.Title)
	}
	if command.Description != nil {
		event.Description = strings.TrimSpace(*command.Description)
	}
	if command.StartAt != nil {
		event.StartAt = command.StartAt.UTC()
	}
	if command.EndAt != nil {
		event.EndAt = command.EndAt.UTC()
	}
	if command.Priority != nil {
		event.Priority = *command.Priority
	}
	if command.IsReschedulable != nil {
		event.IsReschedulable = *command.IsReschedulable
	}
	if command.VisibilityHint != nil {
		event.VisibilityHint = normalizeVisibilityHint(*command.VisibilityHint)
	}
}

func isValidSourceType(value SourceType) bool {
	switch value {
	case SourceTypePersonalTask, SourceTypeMeeting, SourceTypeImported, SourceTypeBlockedTime:
		return true
	default:
		return false
	}
}

func isValidPriority(value Priority) bool {
	switch value {
	case PriorityLow, PriorityMedium, PriorityHigh, PriorityCritical:
		return true
	default:
		return false
	}
}

func normalizeOptionalString(value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}

func normalizeVisibilityHint(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "default"
	}

	return trimmed
}
