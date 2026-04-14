package resources

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

func (s *Service) CreateResource(ctx context.Context, command CreateResourceCommand) (*Resource, error) {
	resource := &Resource{
		ID:          uuid.New(),
		Name:        strings.TrimSpace(command.Name),
		Type:        command.Type,
		Description: strings.TrimSpace(command.Description),
		Capacity:    normalizeCapacity(command.Capacity),
		Location:    normalizeOptionalString(command.Location),
		OwnerUserID: command.OwnerUserID,
	}
	if err := validateResource(resource); err != nil {
		return nil, err
	}

	if err := s.repository.CreateResource(ctx, resource); err != nil {
		return nil, newServiceError(ErrCodeInternal, "failed to create resource", err)
	}

	return resource, nil
}

func (s *Service) ListResources(ctx context.Context, _ ListResourcesQuery) ([]Resource, error) {
	items, err := s.repository.ListResources(ctx)
	if err != nil {
		return nil, newServiceError(ErrCodeInternal, "failed to list resources", err)
	}

	return items, nil
}

func (s *Service) GetResource(ctx context.Context, query GetResourceQuery) (*Resource, error) {
	resource, err := s.repository.GetResourceByID(ctx, query.ID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, newServiceError(ErrCodeNotFound, "resource not found", err)
		}
		return nil, newServiceError(ErrCodeInternal, "failed to get resource", err)
	}

	return resource, nil
}

func (s *Service) UpdateResource(ctx context.Context, command UpdateResourceCommand) (*Resource, error) {
	resource, err := s.repository.GetResourceByID(ctx, command.ID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, newServiceError(ErrCodeNotFound, "resource not found", err)
		}
		return nil, newServiceError(ErrCodeInternal, "failed to update resource", err)
	}

	applyUpdate(resource, command)
	if err := validateResource(resource); err != nil {
		return nil, err
	}

	if err := s.repository.UpdateResource(ctx, resource); err != nil {
		return nil, newServiceError(ErrCodeInternal, "failed to update resource", err)
	}

	return resource, nil
}

func (s *Service) DeleteResource(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repository.GetResourceByID(ctx, id); err != nil {
		if errors.Is(err, ErrNotFound) {
			return newServiceError(ErrCodeNotFound, "resource not found", err)
		}
		return newServiceError(ErrCodeInternal, "failed to delete resource", err)
	}

	if err := s.repository.DeleteResource(ctx, id); err != nil {
		return newServiceError(ErrCodeInternal, "failed to delete resource", err)
	}

	return nil
}

func (s *Service) GetAvailability(ctx context.Context, query GetResourceAvailabilityQuery) (*ResourceAvailability, error) {
	if query.DateFrom.IsZero() || query.DateTo.IsZero() || !query.DateFrom.Before(query.DateTo) {
		return nil, newServiceError(ErrCodeValidation, "date_from must be before date_to", nil)
	}
	if _, err := s.repository.GetResourceByID(ctx, query.ResourceID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, newServiceError(ErrCodeNotFound, "resource not found", err)
		}
		return nil, newServiceError(ErrCodeInternal, "failed to get resource availability", err)
	}

	bookings, err := s.repository.ListBookingsByResource(ctx, query.ResourceID, query.DateFrom, query.DateTo)
	if err != nil {
		return nil, newServiceError(ErrCodeInternal, "failed to list resource bookings", err)
	}

	return &ResourceAvailability{
		ResourceID:  query.ResourceID,
		DateFrom:    query.DateFrom.UTC(),
		DateTo:      query.DateTo.UTC(),
		IsAvailable: len(bookings) == 0,
		Bookings:    bookings,
	}, nil
}

func (s *Service) CreateBooking(ctx context.Context, command CreateBookingCommand) (*ResourceBooking, error) {
	if err := validateCreateBookingCommand(command); err != nil {
		return nil, err
	}
	if _, err := s.repository.GetResourceByID(ctx, command.ResourceID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, newServiceError(ErrCodeNotFound, "resource not found", err)
		}
		return nil, newServiceError(ErrCodeInternal, "failed to create booking", err)
	}

	hasOverlap, err := s.repository.HasBookingOverlap(ctx, command.ResourceID, command.StartAt, command.EndAt, nil)
	if err != nil {
		return nil, newServiceError(ErrCodeInternal, "failed to validate booking overlap", err)
	}
	if hasOverlap {
		return nil, newServiceError(ErrCodeConflict, "resource is already booked for the selected time", nil)
	}

	booking := &ResourceBooking{
		ID:             uuid.New(),
		ResourceID:     command.ResourceID,
		EventID:        command.EventID,
		BookedByUserID: command.BookedByUserID,
		StartAt:        command.StartAt.UTC(),
		EndAt:          command.EndAt.UTC(),
		Title:          strings.TrimSpace(command.Title),
	}
	if err := s.repository.CreateBooking(ctx, booking); err != nil {
		return nil, newServiceError(ErrCodeInternal, "failed to create booking", err)
	}

	return booking, nil
}

func (s *Service) CancelBooking(ctx context.Context, command CancelBookingCommand) error {
	resource, err := s.repository.GetResourceByID(ctx, command.ResourceID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return newServiceError(ErrCodeNotFound, "resource not found", err)
		}
		return newServiceError(ErrCodeInternal, "failed to cancel booking", err)
	}

	booking, err := s.repository.GetBookingByID(ctx, command.BookingID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return newServiceError(ErrCodeNotFound, "booking not found", err)
		}
		return newServiceError(ErrCodeInternal, "failed to cancel booking", err)
	}

	if booking.ResourceID != command.ResourceID {
		return newServiceError(ErrCodeNotFound, "booking not found", nil)
	}

	if booking.BookedByUserID != command.UserID && !isResourceOwner(resource, command.UserID) {
		return newServiceError(ErrCodeForbidden, "you are not allowed to cancel this booking", nil)
	}

	if err := s.repository.DeleteBooking(ctx, command.BookingID); err != nil {
		return newServiceError(ErrCodeInternal, "failed to cancel booking", err)
	}

	return nil
}

func validateResource(resource *Resource) error {
	if resource == nil {
		return newServiceError(ErrCodeValidation, "resource is required", nil)
	}
	if strings.TrimSpace(resource.Name) == "" {
		return newServiceError(ErrCodeValidation, "name is required", nil)
	}
	if !isValidResourceType(resource.Type) {
		return newServiceError(ErrCodeValidation, "invalid type", nil)
	}
	if resource.Capacity != nil && *resource.Capacity <= 0 {
		return newServiceError(ErrCodeValidation, "capacity must be greater than zero", nil)
	}

	return nil
}

func validateCreateBookingCommand(command CreateBookingCommand) error {
	if strings.TrimSpace(command.Title) == "" {
		return newServiceError(ErrCodeValidation, "title is required", nil)
	}
	if command.StartAt.IsZero() || command.EndAt.IsZero() || !command.StartAt.Before(command.EndAt) {
		return newServiceError(ErrCodeValidation, "start_at must be before end_at", nil)
	}

	return nil
}

func applyUpdate(resource *Resource, command UpdateResourceCommand) {
	if command.Name != nil {
		resource.Name = strings.TrimSpace(*command.Name)
	}
	if command.Type != nil {
		resource.Type = *command.Type
	}
	if command.Description != nil {
		resource.Description = strings.TrimSpace(*command.Description)
	}
	if command.CapacitySet {
		if command.ClearCapacity {
			resource.Capacity = nil
		} else {
			resource.Capacity = normalizeCapacity(command.Capacity)
		}
	}
	if command.LocationSet {
		if command.ClearLocation {
			resource.Location = nil
		} else {
			resource.Location = normalizeOptionalString(command.Location)
		}
	}
}

func isValidResourceType(resourceType ResourceType) bool {
	switch resourceType {
	case ResourceTypeMeetingRoom, ResourceTypeEquipment, ResourceTypeSharedSpace, ResourceTypeOther:
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

func normalizeCapacity(value *int) *int {
	if value == nil {
		return nil
	}

	capacity := *value
	return &capacity
}

func isResourceOwner(resource *Resource, userID uuid.UUID) bool {
	return resource != nil && resource.OwnerUserID != nil && *resource.OwnerUserID == userID
}
