package users

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repository RepositoryPort
}

var allowedUnavailabilityTypes = []UnavailabilityType{
	UnavailabilityVacation,
	UnavailabilitySickLeave,
	UnavailabilityBusinessTrip,
	UnavailabilityCustom,
}

func NewService(repository RepositoryPort) *Service {
	return &Service{repository: repository}
}

func (s *Service) GetProfile(ctx context.Context, query GetProfileQuery) (*ProfileResult, error) {
	user, err := s.repository.GetUserByID(ctx, query.UserID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, newServiceError(ErrCodeNotFound, "user not found", err)
		}

		return nil, newServiceError(ErrCodeInternal, "failed to fetch profile", err)
	}

	return mapProfileResult(user), nil
}

func (s *Service) UpdateProfile(ctx context.Context, command UpdateProfileCommand) (*ProfileResult, error) {
	user, err := s.repository.GetUserByID(ctx, command.UserID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, newServiceError(ErrCodeNotFound, "user not found", err)
		}

		return nil, newServiceError(ErrCodeInternal, "failed to update profile", err)
	}

	if command.FullName != "" {
		user.FullName = strings.TrimSpace(command.FullName)
	}

	if command.Timezone != "" {
		if err := validateTimezone(command.Timezone); err != nil {
			return nil, err
		}
		user.Timezone = strings.TrimSpace(command.Timezone)
	}

	if err := s.repository.UpdateUserProfile(ctx, user); err != nil {
		return nil, newServiceError(ErrCodeInternal, "failed to update profile", err)
	}

	return mapProfileResult(user), nil
}

func (s *Service) ReplaceWorkingHours(ctx context.Context, command ReplaceWorkingHoursCommand) ([]WorkingHours, error) {
	if len(command.Items) == 0 {
		return nil, newServiceError(ErrCodeValidation, "working hours list cannot be empty", nil)
	}

	seenWeekdays := make(map[int]struct{}, len(command.Items))
	items := make([]WorkingHours, 0, len(command.Items))
	for _, item := range command.Items {
		if _, exists := seenWeekdays[item.Weekday]; exists {
			return nil, newServiceError(ErrCodeValidation, fmt.Sprintf("duplicate weekday: %d", item.Weekday), nil)
		}

		seenWeekdays[item.Weekday] = struct{}{}
		if err := validateWorkingHoursItem(item); err != nil {
			return nil, err
		}

		items = append(items, WorkingHours{
			ID:           uuid.New(),
			UserID:       command.UserID,
			Weekday:      item.Weekday,
			StartTime:    normalizeClock(item.StartTime),
			EndTime:      normalizeClock(item.EndTime),
			IsWorkingDay: item.IsWorkingDay,
		})
	}

	if err := s.repository.ReplaceWorkingHours(ctx, command.UserID, items); err != nil {
		return nil, newServiceError(ErrCodeInternal, "failed to replace working hours", err)
	}

	return s.repository.ListWorkingHours(ctx, command.UserID)
}

func (s *Service) ListWorkingHours(ctx context.Context, query ListWorkingHoursQuery) ([]WorkingHours, error) {
	items, err := s.repository.ListWorkingHours(ctx, query.UserID)
	if err != nil {
		return nil, newServiceError(ErrCodeInternal, "failed to list working hours", err)
	}

	return items, nil
}

func (s *Service) CreateUnavailabilityPeriod(ctx context.Context, command CreateUnavailabilityCommand) (*UnavailabilityPeriod, error) {
	if !slices.Contains(allowedUnavailabilityTypes, command.Type) {
		return nil, newServiceError(ErrCodeValidation, "invalid unavailability type", nil)
	}

	if command.StartAt.IsZero() || command.EndAt.IsZero() || !command.StartAt.Before(command.EndAt) {
		return nil, newServiceError(ErrCodeValidation, "start_at must be before end_at", nil)
	}

	hasOverlap, err := s.repository.HasUnavailabilityOverlap(ctx, command.UserID, command.StartAt, command.EndAt)
	if err != nil {
		return nil, newServiceError(ErrCodeInternal, "failed to validate unavailability overlap", err)
	}

	if hasOverlap {
		return nil, newServiceError(ErrCodeConflict, "unavailability period overlaps an existing period", nil)
	}

	item := &UnavailabilityPeriod{
		ID:      uuid.New(),
		UserID:  command.UserID,
		Type:    command.Type,
		Title:   strings.TrimSpace(command.Title),
		StartAt: command.StartAt.UTC(),
		EndAt:   command.EndAt.UTC(),
		Comment: strings.TrimSpace(command.Comment),
	}

	if err := s.repository.CreateUnavailabilityPeriod(ctx, item); err != nil {
		return nil, newServiceError(ErrCodeInternal, "failed to create unavailability period", err)
	}

	return item, nil
}

func (s *Service) ListUnavailabilityPeriods(ctx context.Context, query ListUnavailabilityQuery) ([]UnavailabilityPeriod, error) {
	items, err := s.repository.ListUnavailabilityPeriods(ctx, query.UserID)
	if err != nil {
		return nil, newServiceError(ErrCodeInternal, "failed to list unavailability periods", err)
	}

	return items, nil
}

func (s *Service) DeleteUnavailabilityPeriod(ctx context.Context, command DeleteUnavailabilityCommand) error {
	item, err := s.repository.GetUnavailabilityPeriodByID(ctx, command.ID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return newServiceError(ErrCodeNotFound, "unavailability period not found", err)
		}

		return newServiceError(ErrCodeInternal, "failed to delete unavailability period", err)
	}

	if item.UserID != command.UserID {
		return newServiceError(ErrCodeForbidden, "unavailability period does not belong to current user", nil)
	}

	if err := s.repository.DeleteUnavailabilityPeriod(ctx, command.ID); err != nil {
		return newServiceError(ErrCodeInternal, "failed to delete unavailability period", err)
	}

	return nil
}

func validateTimezone(value string) error {
	if _, err := time.LoadLocation(strings.TrimSpace(value)); err != nil {
		return newServiceError(ErrCodeValidation, "invalid timezone", err)
	}

	return nil
}

func validateWorkingHoursItem(item WorkingHoursInput) error {
	if item.Weekday < 0 || item.Weekday > 6 {
		return newServiceError(ErrCodeValidation, "weekday must be between 0 and 6", nil)
	}

	if !item.IsWorkingDay {
		return nil
	}

	start, err := parseClock(item.StartTime)
	if err != nil {
		return newServiceError(ErrCodeValidation, "invalid start_time format, expected HH:MM", err)
	}

	end, err := parseClock(item.EndTime)
	if err != nil {
		return newServiceError(ErrCodeValidation, "invalid end_time format, expected HH:MM", err)
	}

	if !start.Before(end) {
		return newServiceError(ErrCodeValidation, "start_time must be before end_time", nil)
	}

	return nil
}

func parseClock(value string) (time.Time, error) {
	return time.Parse("15:04", strings.TrimSpace(value))
}

func normalizeClock(value string) string {
	parsed, err := parseClock(value)
	if err != nil {
		return strings.TrimSpace(value)
	}

	return parsed.Format("15:04")
}

func mapProfileResult(user *User) *ProfileResult {
	return &ProfileResult{
		ID:        user.ID,
		Email:     user.Email,
		FullName:  user.FullName,
		Timezone:  user.Timezone,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
