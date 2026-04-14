package meetings

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repository        RepositoryPort
	userDirectory     UserDirectoryPort
	resourceDirectory ResourceDirectoryPort
	notifications     NotificationPort
}

func NewService(repository RepositoryPort, userDirectory UserDirectoryPort, resourceDirectory ResourceDirectoryPort, notifications NotificationPort) *Service {
	if notifications == nil {
		notifications = NewNoopNotificationHook()
	}
	return &Service{
		repository:        repository,
		userDirectory:     userDirectory,
		resourceDirectory: resourceDirectory,
		notifications:     notifications,
	}
}

func (s *Service) CreateMeeting(ctx context.Context, command CreateMeetingCommand) (*Meeting, error) {
	meeting := &Meeting{
		ID:                uuid.New(),
		OrganizerUserID:   command.OrganizerUserID,
		Title:             strings.TrimSpace(command.Title),
		Description:       strings.TrimSpace(command.Description),
		DurationMinutes:   command.DurationMinutes,
		Priority:          normalizePriority(command.Priority),
		Status:            deriveMeetingStatus(command.Status, command.SelectedStartAt, command.SelectedEndAt),
		SearchRangeStart:  command.SearchRangeStart.UTC(),
		SearchRangeEnd:    command.SearchRangeEnd.UTC(),
		EarliestStartTime: normalizeOptionalString(command.EarliestStartTime),
		LatestStartTime:   normalizeOptionalString(command.LatestStartTime),
		RecurrenceRule:    normalizeOptionalString(command.RecurrenceRule),
		SelectedStartAt:   normalizeTimePointer(command.SelectedStartAt),
		SelectedEndAt:     normalizeTimePointer(command.SelectedEndAt),
	}
	if err := validateMeeting(meeting); err != nil {
		return nil, err
	}
	if err := s.repository.CreateMeeting(ctx, meeting); err != nil {
		return nil, newServiceError(ErrCodeInternal, "failed to create meeting", err)
	}

	organizerParticipant := MeetingParticipant{
		ID:             uuid.New(),
		MeetingID:      meeting.ID,
		UserID:         meeting.OrganizerUserID,
		ResponseStatus: InvitationStatusAccepted,
	}
	if err := s.repository.AddParticipants(ctx, []MeetingParticipant{organizerParticipant}); err != nil {
		return nil, newServiceError(ErrCodeInternal, "failed to create organizer participant", err)
	}

	return s.repository.GetMeetingByID(ctx, meeting.ID)
}

func (s *Service) ListMeetings(ctx context.Context, query ListMeetingsQuery) ([]Meeting, error) {
	items, err := s.repository.ListMeetingsForUser(ctx, query.UserID)
	if err != nil {
		return nil, newServiceError(ErrCodeInternal, "failed to list meetings", err)
	}
	return items, nil
}

func (s *Service) GetMeeting(ctx context.Context, query GetMeetingQuery) (*Meeting, error) {
	meeting, err := s.repository.GetMeetingByID(ctx, query.ID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, newServiceError(ErrCodeNotFound, "meeting not found", err)
		}
		return nil, newServiceError(ErrCodeInternal, "failed to get meeting", err)
	}
	if !canAccessMeeting(meeting, query.UserID) {
		return nil, newServiceError(ErrCodeNotFound, "meeting not found", nil)
	}
	return meeting, nil
}

func (s *Service) UpdateMeeting(ctx context.Context, command UpdateMeetingCommand) (*Meeting, error) {
	meeting, err := s.repository.GetMeetingByID(ctx, command.ID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, newServiceError(ErrCodeNotFound, "meeting not found", err)
		}
		return nil, newServiceError(ErrCodeInternal, "failed to update meeting", err)
	}
	if meeting.OrganizerUserID != command.OrganizerUserID {
		return nil, newServiceError(ErrCodeForbidden, "only organizer can update meeting", nil)
	}
	applyMeetingUpdate(meeting, command)
	if err := validateMeeting(meeting); err != nil {
		return nil, err
	}
	if err := s.validateLinkedResourcesAvailability(ctx, meeting); err != nil {
		return nil, err
	}
	if err := s.repository.UpdateMeeting(ctx, meeting); err != nil {
		return nil, newServiceError(ErrCodeInternal, "failed to update meeting", err)
	}
	s.notifications.NotifyMeetingUpdated(ctx, meeting)
	return meeting, nil
}

func (s *Service) DeleteMeeting(ctx context.Context, meetingID, organizerUserID uuid.UUID) error {
	meeting, err := s.repository.GetMeetingByID(ctx, meetingID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return newServiceError(ErrCodeNotFound, "meeting not found", err)
		}
		return newServiceError(ErrCodeInternal, "failed to delete meeting", err)
	}
	if meeting.OrganizerUserID != organizerUserID {
		return newServiceError(ErrCodeForbidden, "only organizer can delete meeting", nil)
	}
	if err := s.repository.DeleteMeeting(ctx, meetingID); err != nil {
		return newServiceError(ErrCodeInternal, "failed to delete meeting", err)
	}
	return nil
}

func (s *Service) AddParticipants(ctx context.Context, command AddParticipantsCommand) (*Meeting, error) {
	meeting, err := s.requireOrganizerMeeting(ctx, command.MeetingID, command.OrganizerUserID)
	if err != nil {
		return nil, err
	}
	toAdd := make([]MeetingParticipant, 0, len(command.ParticipantIDs))
	for _, participantID := range uniqueUUIDs(command.ParticipantIDs) {
		if participantID == meeting.OrganizerUserID {
			continue
		}
		exists, err := s.userDirectory.Exists(ctx, participantID)
		if err != nil {
			return nil, newServiceError(ErrCodeInternal, "failed to validate participant", err)
		}
		if !exists {
			return nil, newServiceError(ErrCodeValidation, fmt.Sprintf("participant %s not found", participantID), nil)
		}
		_, err = s.repository.GetParticipant(ctx, meeting.ID, participantID)
		if err == nil {
			continue
		}
		if !errors.Is(err, ErrNotFound) {
			return nil, newServiceError(ErrCodeInternal, "failed to validate participant", err)
		}
		toAdd = append(toAdd, MeetingParticipant{
			ID:             uuid.New(),
			MeetingID:      meeting.ID,
			UserID:         participantID,
			ResponseStatus: InvitationStatusPending,
		})
	}
	if err := s.repository.AddParticipants(ctx, toAdd); err != nil {
		if errors.Is(err, ErrAlreadyExists) {
			return nil, newServiceError(ErrCodeConflict, "participant already exists", err)
		}
		return nil, newServiceError(ErrCodeInternal, "failed to add participants", err)
	}
	fresh, err := s.repository.GetMeetingByID(ctx, meeting.ID)
	if err != nil {
		return nil, newServiceError(ErrCodeInternal, "failed to load meeting", err)
	}
	s.notifications.NotifyParticipantsInvited(ctx, fresh, participantIDsFromItems(toAdd))
	return fresh, nil
}

func (s *Service) RemoveParticipant(ctx context.Context, command RemoveParticipantCommand) error {
	meeting, err := s.repository.GetMeetingByID(ctx, command.MeetingID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return newServiceError(ErrCodeNotFound, "meeting not found", err)
		}
		return newServiceError(ErrCodeInternal, "failed to remove participant", err)
	}
	if command.ParticipantUserID == meeting.OrganizerUserID {
		return newServiceError(ErrCodeForbidden, "organizer cannot be removed from own meeting", nil)
	}
	if command.ActingUserID != meeting.OrganizerUserID && command.ActingUserID != command.ParticipantUserID {
		return newServiceError(ErrCodeForbidden, "not allowed to remove participant", nil)
	}
	if _, err := s.repository.GetParticipant(ctx, meeting.ID, command.ParticipantUserID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return newServiceError(ErrCodeNotFound, "participant not found", err)
		}
		return newServiceError(ErrCodeInternal, "failed to remove participant", err)
	}
	if err := s.repository.DeleteParticipant(ctx, meeting.ID, command.ParticipantUserID); err != nil {
		return newServiceError(ErrCodeInternal, "failed to remove participant", err)
	}
	return nil
}

func (s *Service) AddResources(ctx context.Context, command AddResourceCommand) (*Meeting, error) {
	meeting, err := s.requireOrganizerMeeting(ctx, command.MeetingID, command.OrganizerUserID)
	if err != nil {
		return nil, err
	}
	for _, resourceID := range uniqueUUIDs(command.ResourceIDs) {
		exists, err := s.resourceDirectory.Exists(ctx, resourceID)
		if err != nil {
			return nil, newServiceError(ErrCodeInternal, "failed to validate resource", err)
		}
		if !exists {
			return nil, newServiceError(ErrCodeValidation, fmt.Sprintf("resource %s not found", resourceID), nil)
		}
		if meeting.SelectedStartAt != nil && meeting.SelectedEndAt != nil {
			available, err := s.resourceDirectory.IsAvailable(ctx, resourceID, *meeting.SelectedStartAt, *meeting.SelectedEndAt)
			if err != nil {
				return nil, newServiceError(ErrCodeInternal, "failed to validate resource availability", err)
			}
			if !available {
				s.notifications.NotifyResourceConflict(ctx, meeting, resourceID, *meeting.SelectedStartAt, *meeting.SelectedEndAt)
				return nil, newServiceError(ErrCodeConflict, "resource is not available for selected meeting time", nil)
			}
		}
		if _, err := s.repository.GetMeetingResource(ctx, meeting.ID, resourceID); err == nil {
			continue
		} else if !errors.Is(err, ErrNotFound) {
			return nil, newServiceError(ErrCodeInternal, "failed to validate meeting resource", err)
		}
		if err := s.repository.AddMeetingResource(ctx, &MeetingResource{
			ID:         uuid.New(),
			MeetingID:  meeting.ID,
			ResourceID: resourceID,
		}); err != nil {
			if errors.Is(err, ErrAlreadyExists) {
				return nil, newServiceError(ErrCodeConflict, "resource already added", err)
			}
			return nil, newServiceError(ErrCodeInternal, "failed to add resource", err)
		}
	}
	return s.repository.GetMeetingByID(ctx, meeting.ID)
}

func (s *Service) RemoveResource(ctx context.Context, command RemoveResourceCommand) error {
	meeting, err := s.requireOrganizerMeeting(ctx, command.MeetingID, command.OrganizerUserID)
	if err != nil {
		return err
	}
	if _, err := s.repository.GetMeetingResource(ctx, meeting.ID, command.ResourceID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return newServiceError(ErrCodeNotFound, "meeting resource not found", err)
		}
		return newServiceError(ErrCodeInternal, "failed to remove resource", err)
	}
	if err := s.repository.DeleteMeetingResource(ctx, meeting.ID, command.ResourceID); err != nil {
		return newServiceError(ErrCodeInternal, "failed to remove resource", err)
	}
	return nil
}

func (s *Service) RespondInvitation(ctx context.Context, command RespondInvitationCommand) (*MeetingParticipant, error) {
	participant, meeting, err := s.requireMeetingParticipant(ctx, command.MeetingID, command.UserID)
	if err != nil {
		return nil, err
	}
	_ = meeting
	participant.ResponseStatus = command.ResponseStatus
	participant.AlternativeStartAt = nil
	participant.AlternativeEndAt = nil
	participant.AlternativeComment = nil
	if err := s.repository.UpdateParticipant(ctx, participant); err != nil {
		return nil, newServiceError(ErrCodeInternal, "failed to respond to invitation", err)
	}
	return participant, nil
}

func (s *Service) RequestAlternative(ctx context.Context, command RequestAlternativeCommand) (*MeetingParticipant, error) {
	participant, meeting, err := s.requireMeetingParticipant(ctx, command.MeetingID, command.UserID)
	if err != nil {
		return nil, err
	}
	if command.ProposedStartAt.IsZero() || command.ProposedEndAt.IsZero() || !command.ProposedStartAt.Before(command.ProposedEndAt) {
		return nil, newServiceError(ErrCodeValidation, "proposed_start_at must be before proposed_end_at", nil)
	}
	if command.ProposedStartAt.Before(meeting.SearchRangeStart) || command.ProposedEndAt.After(meeting.SearchRangeEnd) {
		return nil, newServiceError(ErrCodeValidation, "alternative time must be within search range", nil)
	}
	participant.ResponseStatus = InvitationStatusTentative
	participant.AlternativeStartAt = normalizeTimePointer(&command.ProposedStartAt)
	participant.AlternativeEndAt = normalizeTimePointer(&command.ProposedEndAt)
	participant.AlternativeComment = normalizeOptionalString(command.Comment)
	if err := s.repository.UpdateParticipant(ctx, participant); err != nil {
		return nil, newServiceError(ErrCodeInternal, "failed to request alternative time", err)
	}
	s.notifications.NotifyAlternativeRequested(ctx, meeting, participant)
	return participant, nil
}

func (s *Service) requireOrganizerMeeting(ctx context.Context, meetingID, organizerUserID uuid.UUID) (*Meeting, error) {
	meeting, err := s.repository.GetMeetingByID(ctx, meetingID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, newServiceError(ErrCodeNotFound, "meeting not found", err)
		}
		return nil, newServiceError(ErrCodeInternal, "failed to load meeting", err)
	}
	if meeting.OrganizerUserID != organizerUserID {
		return nil, newServiceError(ErrCodeForbidden, "only organizer can modify meeting", nil)
	}
	return meeting, nil
}

func (s *Service) requireMeetingParticipant(ctx context.Context, meetingID, userID uuid.UUID) (*MeetingParticipant, *Meeting, error) {
	meeting, err := s.repository.GetMeetingByID(ctx, meetingID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, nil, newServiceError(ErrCodeNotFound, "meeting not found", err)
		}
		return nil, nil, newServiceError(ErrCodeInternal, "failed to load meeting", err)
	}
	participant, err := s.repository.GetParticipant(ctx, meetingID, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, nil, newServiceError(ErrCodeForbidden, "only participant can perform this action", err)
		}
		return nil, nil, newServiceError(ErrCodeInternal, "failed to load participant", err)
	}
	return participant, meeting, nil
}

func (s *Service) validateLinkedResourcesAvailability(ctx context.Context, meeting *Meeting) error {
	if meeting.SelectedStartAt == nil || meeting.SelectedEndAt == nil {
		return nil
	}
	for i := range meeting.Resources {
		available, err := s.resourceDirectory.IsAvailable(ctx, meeting.Resources[i].ResourceID, *meeting.SelectedStartAt, *meeting.SelectedEndAt)
		if err != nil {
			return newServiceError(ErrCodeInternal, "failed to validate resource availability", err)
		}
		if !available {
			return newServiceError(ErrCodeConflict, "one or more linked resources are not available for selected meeting time", nil)
		}
	}
	return nil
}

func validateMeeting(meeting *Meeting) error {
	if meeting == nil {
		return newServiceError(ErrCodeValidation, "meeting is required", nil)
	}
	if strings.TrimSpace(meeting.Title) == "" {
		return newServiceError(ErrCodeValidation, "title is required", nil)
	}
	if meeting.DurationMinutes <= 0 {
		return newServiceError(ErrCodeValidation, "duration_minutes must be greater than zero", nil)
	}
	if meeting.SearchRangeStart.IsZero() || meeting.SearchRangeEnd.IsZero() || !meeting.SearchRangeStart.Before(meeting.SearchRangeEnd) {
		return newServiceError(ErrCodeValidation, "search_range_start must be before search_range_end", nil)
	}
	if meeting.EarliestStartTime != nil && !isValidClockValue(*meeting.EarliestStartTime) {
		return newServiceError(ErrCodeValidation, "invalid earliest_start_time", nil)
	}
	if meeting.LatestStartTime != nil && !isValidClockValue(*meeting.LatestStartTime) {
		return newServiceError(ErrCodeValidation, "invalid latest_start_time", nil)
	}
	if !isValidMeetingStatus(meeting.Status) {
		return newServiceError(ErrCodeValidation, "invalid status", nil)
	}
	if !isValidPriority(meeting.Priority) {
		return newServiceError(ErrCodeValidation, "invalid priority", nil)
	}
	if (meeting.SelectedStartAt == nil) != (meeting.SelectedEndAt == nil) {
		return newServiceError(ErrCodeValidation, "selected_start_at and selected_end_at must be provided together", nil)
	}
	if meeting.SelectedStartAt != nil && meeting.SelectedEndAt != nil {
		if !meeting.SelectedStartAt.Before(*meeting.SelectedEndAt) {
			return newServiceError(ErrCodeValidation, "selected_start_at must be before selected_end_at", nil)
		}
		if meeting.SelectedStartAt.Before(meeting.SearchRangeStart) || meeting.SelectedEndAt.After(meeting.SearchRangeEnd) {
			return newServiceError(ErrCodeValidation, "selected time must be within search range", nil)
		}
		duration := int(meeting.SelectedEndAt.Sub(*meeting.SelectedStartAt).Minutes())
		if duration != meeting.DurationMinutes {
			return newServiceError(ErrCodeValidation, "selected time must match duration_minutes", nil)
		}
	}
	return nil
}

func applyMeetingUpdate(meeting *Meeting, command UpdateMeetingCommand) {
	if command.Title != nil {
		meeting.Title = strings.TrimSpace(*command.Title)
	}
	if command.Description != nil {
		meeting.Description = strings.TrimSpace(*command.Description)
	}
	if command.DurationMinutes != nil {
		meeting.DurationMinutes = *command.DurationMinutes
	}
	if command.Priority != nil {
		meeting.Priority = normalizePriority(*command.Priority)
	}
	if command.Status != nil {
		meeting.Status = *command.Status
	}
	if command.SearchRangeStart != nil {
		meeting.SearchRangeStart = command.SearchRangeStart.UTC()
	}
	if command.SearchRangeEnd != nil {
		meeting.SearchRangeEnd = command.SearchRangeEnd.UTC()
	}
	if command.EarliestSet {
		if command.ClearEarliest {
			meeting.EarliestStartTime = nil
		} else {
			meeting.EarliestStartTime = normalizeOptionalString(command.EarliestStartTime)
		}
	}
	if command.LatestSet {
		if command.ClearLatest {
			meeting.LatestStartTime = nil
		} else {
			meeting.LatestStartTime = normalizeOptionalString(command.LatestStartTime)
		}
	}
	if command.RecurrenceSet {
		if command.ClearRecurrence {
			meeting.RecurrenceRule = nil
		} else {
			meeting.RecurrenceRule = normalizeOptionalString(command.RecurrenceRule)
		}
	}
	if command.SelectedSet {
		if command.ClearSelected {
			meeting.SelectedStartAt = nil
			meeting.SelectedEndAt = nil
		} else {
			meeting.SelectedStartAt = normalizeTimePointer(command.SelectedStartAt)
			meeting.SelectedEndAt = normalizeTimePointer(command.SelectedEndAt)
		}
	}
	if meeting.SelectedStartAt != nil && meeting.SelectedEndAt != nil && meeting.Status == MeetingStatusDraft {
		meeting.Status = MeetingStatusScheduled
	}
	if meeting.SelectedStartAt == nil && meeting.SelectedEndAt == nil && meeting.Status == MeetingStatusScheduled {
		meeting.Status = MeetingStatusDraft
	}
}

func canAccessMeeting(meeting *Meeting, userID uuid.UUID) bool {
	if meeting == nil {
		return false
	}
	if meeting.OrganizerUserID == userID {
		return true
	}
	for i := range meeting.Participants {
		if meeting.Participants[i].UserID == userID {
			return true
		}
	}
	return false
}

func uniqueUUIDs(values []uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(values))
	result := make([]uuid.UUID, 0, len(values))
	for _, value := range values {
		if value == uuid.Nil {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func participantIDsFromItems(items []MeetingParticipant) []uuid.UUID {
	result := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		if item.UserID == uuid.Nil {
			continue
		}
		result = append(result, item.UserID)
	}
	return result
}

func deriveMeetingStatus(status *MeetingStatus, selectedStartAt, selectedEndAt *time.Time) MeetingStatus {
	if status != nil {
		return *status
	}
	if selectedStartAt != nil && selectedEndAt != nil {
		return MeetingStatusScheduled
	}
	return MeetingStatusDraft
}

func normalizePriority(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "medium"
	}
	return trimmed
}

func isValidPriority(value string) bool {
	switch value {
	case "low", "medium", "high", "critical":
		return true
	default:
		return false
	}
}

func isValidMeetingStatus(value MeetingStatus) bool {
	switch value {
	case MeetingStatusDraft, MeetingStatusScheduled, MeetingStatusCancelled:
		return true
	default:
		return false
	}
}

func isValidInvitationStatus(value InvitationStatus) bool {
	switch value {
	case InvitationStatusPending, InvitationStatusAccepted, InvitationStatusDeclined, InvitationStatusTentative:
		return true
	default:
		return false
	}
}

func isValidClockValue(value string) bool {
	_, err := time.Parse("15:04", value)
	return err == nil
}

func normalizeTimePointer(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	normalized := value.UTC()
	return &normalized
}
