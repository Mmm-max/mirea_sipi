package scheduling

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	defaultTopSlots   = 10
	slotStep          = 15 * time.Minute
	softPenaltyLow    = 10
	softPenaltyMedium = 25
	softPenaltyHigh   = 45
)

type Service struct {
	repository    RepositoryPort
	meetings      MeetingStorePort
	availability  AvailabilityStorePort
	eventSync     EventSyncPort
	notifications NotificationPort
}

// NewService создаёт новый экземпляр сервиса планирования.
// Если notifications равен nil, используется заглушка (noop).
func NewService(repository RepositoryPort, meetings MeetingStorePort, availability AvailabilityStorePort, eventSync EventSyncPort, notifications NotificationPort) *Service {
	if notifications == nil {
		notifications = NewNoopNotificationHook()
	}
	return &Service{
		repository:    repository,
		meetings:      meetings,
		availability:  availability,
		eventSync:     eventSync,
		notifications: notifications,
	}
}

// SearchSlots выполняет поиск доступных временных слотов для встречи.
// Перебирает кандидатные временны́е окна с шагом 15 минут в пределах
// SearchRangeStart..SearchRangeEnd, оценивает каждое окно через evaluateSlot,
// сортирует по убыванию Score и возвращает топ-N слотов (по умолчанию 10).
// Результаты сохраняются в БД и перезаписывают предыдущие слоты встречи.
// Вызывать может только организатор встречи.
func (s *Service) SearchSlots(ctx context.Context, command SearchSlotsCommand) (*SearchResult, error) {
	meeting, err := s.meetings.GetMeeting(ctx, command.MeetingID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, newServiceError(ErrCodeNotFound, "meeting not found", err)
		}
		return nil, newServiceError(ErrCodeInternal, "failed to load meeting", err)
	}
	if meeting.OrganizerUserID != command.OrganizerUserID {
		return nil, newServiceError(ErrCodeForbidden, "only organizer can search slots", nil)
	}
	if err := validateMeetingAggregate(meeting); err != nil {
		return nil, err
	}

	topN := command.TopN
	if topN <= 0 || topN > 50 {
		topN = defaultTopSlots
	}

	starts := buildCandidateStarts(meeting)
	slots := make([]MeetingSlot, 0, len(starts))
	for _, startAt := range starts {
		slot := s.evaluateSlot(ctx, meeting, startAt, startAt.Add(time.Duration(meeting.DurationMinutes)*time.Minute))
		slots = append(slots, slot)
	}

	sort.SliceStable(slots, func(i, j int) bool {
		if slots[i].Score == slots[j].Score {
			return slots[i].StartAt.Before(slots[j].StartAt)
		}
		return slots[i].Score > slots[j].Score
	})

	if len(slots) > topN {
		slots = slots[:topN]
	}
	for i := range slots {
		slots[i].Rank = i + 1
	}

	if err := s.repository.ReplaceSlots(ctx, meeting.ID, slots); err != nil {
		return nil, newServiceError(ErrCodeInternal, "failed to store meeting slots", err)
	}

	return &SearchResult{MeetingID: meeting.ID, Slots: slots}, nil
}

func (s *Service) GetSlots(ctx context.Context, query GetSlotsQuery) (*SearchResult, error) {
	meeting, err := s.meetings.GetMeeting(ctx, query.MeetingID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, newServiceError(ErrCodeNotFound, "meeting not found", err)
		}
		return nil, newServiceError(ErrCodeInternal, "failed to load meeting", err)
	}
	if !canAccessMeeting(meeting, query.RequesterUserID) {
		return nil, newServiceError(ErrCodeNotFound, "meeting not found", nil)
	}
	slots, err := s.repository.ListSlotsByMeetingID(ctx, query.MeetingID)
	if err != nil {
		return nil, newServiceError(ErrCodeInternal, "failed to list meeting slots", err)
	}
	return &SearchResult{MeetingID: query.MeetingID, Slots: slots}, nil
}

// SelectSlot подтверждает выбор временного слота для встречи.
// Запрещает выбор слота, имеющего жёсткие конфликты (hard conflicts).
// После выбора обновляет время встречи, синхронизирует события участников
// через EventSyncPort и рассылает уведомления через NotificationPort.
// Вызывать может только организатор встречи.
func (s *Service) SelectSlot(ctx context.Context, command SelectSlotCommand) (*MeetingAggregate, error) {
	meeting, err := s.meetings.GetMeeting(ctx, command.MeetingID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, newServiceError(ErrCodeNotFound, "meeting not found", err)
		}
		return nil, newServiceError(ErrCodeInternal, "failed to load meeting", err)
	}
	if meeting.OrganizerUserID != command.OrganizerUserID {
		return nil, newServiceError(ErrCodeForbidden, "only organizer can select slot", nil)
	}
	slot, err := s.repository.GetSlotByID(ctx, command.MeetingSlotID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, newServiceError(ErrCodeNotFound, "meeting slot not found", err)
		}
		return nil, newServiceError(ErrCodeInternal, "failed to load meeting slot", err)
	}
	if slot.MeetingID != meeting.ID {
		return nil, newServiceError(ErrCodeNotFound, "meeting slot not found", nil)
	}
	if hasHardConflict(slot.Conflicts) {
		return nil, newServiceError(ErrCodeConflict, "cannot select slot with hard conflicts", nil)
	}
	if err := s.meetings.UpdateSelectedTime(ctx, meeting.ID, slot.StartAt, slot.EndAt); err != nil {
		return nil, newServiceError(ErrCodeInternal, "failed to select slot", err)
	}
	meeting.SelectedStartAt = normalizeTimePointer(&slot.StartAt)
	meeting.SelectedEndAt = normalizeTimePointer(&slot.EndAt)
	if err := s.eventSync.SyncSelectedMeeting(ctx, meeting, slot.StartAt, slot.EndAt); err != nil {
		return nil, newServiceError(ErrCodeInternal, "failed to sync selected meeting events", err)
	}
	s.notifications.NotifySlotSelected(ctx, meeting, slot.StartAt, slot.EndAt)
	return s.meetings.GetMeeting(ctx, meeting.ID)
}

// evaluateSlot оценивает кандидатный временной интервал [startAt, endAt].
// Начальная оценка — 100 баллов. За каждый мягкий конфликт (событие с
// приоритетом low/medium/high) штраф вычитается из счёта. Жёсткие конфликты
// (критические события, недоступность ресурсов, выход за рабочие часы)
// обнуляют счёт. Возвращает MeetingSlot с заполненным списком конфликтов.
func (s *Service) evaluateSlot(ctx context.Context, meeting *MeetingAggregate, startAt, endAt time.Time) MeetingSlot {
	slot := MeetingSlot{
		ID:        uuid.New(),
		MeetingID: meeting.ID,
		StartAt:   startAt.UTC(),
		EndAt:     endAt.UTC(),
		Score:     100,
		Conflicts: make([]SlotConflict, 0),
	}

	if !withinMeetingRange(meeting, startAt, endAt) {
		slot.Conflicts = append(slot.Conflicts, newGenericConflict("outside_search_range", "Outside search range", nil, nil, nil))
		slot.Score = 0
		return slot
	}
	if !withinMeetingDailyBounds(meeting, startAt, endAt) {
		slot.Conflicts = append(slot.Conflicts, newGenericConflict("outside_meeting_hours", "Outside meeting preferred hours", nil, nil, nil))
		slot.Score = 0
		return slot
	}

	for _, participant := range meetingParticipants(meeting) {
		workingHours, err := s.availability.ListWorkingHours(ctx, participant.UserID)
		if err == nil && len(workingHours) > 0 && !withinWorkingHours(workingHours, startAt, endAt) {
			userID := participant.UserID
			slot.Conflicts = append(slot.Conflicts, newGenericConflict("outside_working_hours", "Outside working hours", &userID, nil, nil))
			slot.Score = 0
		}

		unavailability, err := s.availability.ListUnavailability(ctx, participant.UserID)
		if err == nil {
			for _, period := range unavailability {
				if overlaps(startAt, endAt, period.StartAt, period.EndAt) {
					userID := participant.UserID
					slot.Conflicts = append(slot.Conflicts, newGenericConflict("unavailability", period.Title, &userID, nil, nil))
					slot.Score = 0
				}
			}
		}

		events, err := s.availability.ListEvents(ctx, participant.UserID, startAt, endAt)
		if err == nil {
			for _, event := range events {
				if !overlaps(startAt, endAt, event.StartAt, event.EndAt) {
					continue
				}
				conflict := buildEventConflict(event, participant, meeting.OrganizerUserID)
				slot.Conflicts = append(slot.Conflicts, conflict)
				if isHardEventConflict(event.Priority) {
					slot.Score = 0
					continue
				}
				slot.Score -= penaltyForPriority(event.Priority)
				if strings.EqualFold(event.Priority, "low") && event.IsReschedulable {
					slot.Conflicts = append(slot.Conflicts, SlotConflict{
						ID:           uuid.New(),
						UserID:       uuidPointer(participant.UserID),
						EventID:      uuidPointer(event.ID),
						ConflictType: "reschedule_suggestion",
						VisibleTitle: "Can be rescheduled",
					})
				}
			}
		}
	}

	for _, resource := range meeting.Resources {
		available, err := s.availability.ResourceAvailable(ctx, resource.ResourceID, startAt, endAt)
		if err == nil && !available {
			resourceID := resource.ResourceID
			slot.Conflicts = append(slot.Conflicts, newGenericConflict("resource_unavailable", "Resource unavailable", nil, nil, &resourceID))
			slot.Score = 0
		}
	}

	if slot.Score < 0 {
		slot.Score = 0
	}
	return slot
}

func validateMeetingAggregate(meeting *MeetingAggregate) error {
	if meeting == nil {
		return newServiceError(ErrCodeValidation, "meeting is required", nil)
	}
	if meeting.DurationMinutes <= 0 {
		return newServiceError(ErrCodeValidation, "meeting duration_minutes must be greater than zero", nil)
	}
	if meeting.SearchRangeStart.IsZero() || meeting.SearchRangeEnd.IsZero() || !meeting.SearchRangeStart.Before(meeting.SearchRangeEnd) {
		return newServiceError(ErrCodeValidation, "meeting search range is invalid", nil)
	}
	return nil
}

func buildCandidateStarts(meeting *MeetingAggregate) []time.Time {
	start := roundUpToStep(meeting.SearchRangeStart.UTC(), slotStep)
	endLimit := meeting.SearchRangeEnd.UTC().Add(-time.Duration(meeting.DurationMinutes) * time.Minute)
	result := make([]time.Time, 0)
	for current := start; !current.After(endLimit); current = current.Add(slotStep) {
		result = append(result, current)
	}
	return result
}

func roundUpToStep(value time.Time, step time.Duration) time.Time {
	truncated := value.Truncate(step)
	if truncated.Equal(value) {
		return truncated
	}
	return truncated.Add(step)
}

func withinMeetingRange(meeting *MeetingAggregate, startAt, endAt time.Time) bool {
	return !startAt.Before(meeting.SearchRangeStart) && !endAt.After(meeting.SearchRangeEnd)
}

func withinMeetingDailyBounds(meeting *MeetingAggregate, startAt, endAt time.Time) bool {
	if meeting.EarliestStartTime != nil {
		earliest, err := time.Parse("15:04", *meeting.EarliestStartTime)
		if err == nil {
			bound := time.Date(startAt.Year(), startAt.Month(), startAt.Day(), earliest.Hour(), earliest.Minute(), 0, 0, startAt.Location())
			if startAt.Before(bound) {
				return false
			}
		}
	}
	if meeting.LatestStartTime != nil {
		latest, err := time.Parse("15:04", *meeting.LatestStartTime)
		if err == nil {
			bound := time.Date(endAt.Year(), endAt.Month(), endAt.Day(), latest.Hour(), latest.Minute(), 0, 0, endAt.Location())
			if endAt.After(bound) {
				return false
			}
		}
	}
	return true
}

// withinWorkingHours проверяет, попадает ли интервал [startAt, endAt]
// в рабочие часы пользователя для соответствующего дня недели.
// Если рабочие часы не заданы, интервал считается допустимым.
// Если день помечен как нерабочий (IsWorkingDay == false), возвращает false.
func withinWorkingHours(items []WorkingHoursWindow, startAt, endAt time.Time) bool {
	if len(items) == 0 {
		return true
	}
	weekday := int(startAt.Weekday())
	for _, item := range items {
		if item.Weekday != weekday {
			continue
		}
		if !item.IsWorkingDay {
			return false
		}
		startBound, err := time.Parse("15:04", item.StartTime)
		if err != nil {
			continue
		}
		endBound, err := time.Parse("15:04", item.EndTime)
		if err != nil {
			continue
		}
		windowStart := time.Date(startAt.Year(), startAt.Month(), startAt.Day(), startBound.Hour(), startBound.Minute(), 0, 0, startAt.Location())
		windowEnd := time.Date(startAt.Year(), startAt.Month(), startAt.Day(), endBound.Hour(), endBound.Minute(), 0, 0, startAt.Location())
		if !startAt.Before(windowStart) && !endAt.After(windowEnd) {
			return true
		}
	}
	return false
}

// overlaps возвращает true, если интервалы [aStart, aEnd) и [bStart, bEnd)
// пересекаются (используется полуоткрытый интервал).
func overlaps(aStart, aEnd, bStart, bEnd time.Time) bool {
	return aStart.Before(bEnd) && aEnd.After(bStart)
}

func meetingParticipants(meeting *MeetingAggregate) []MeetingParticipantRef {
	if meeting == nil {
		return nil
	}
	if len(meeting.Participants) > 0 {
		return meeting.Participants
	}
	return []MeetingParticipantRef{{UserID: meeting.OrganizerUserID}}
}

func buildEventConflict(event CalendarEvent, participant MeetingParticipantRef, requesterUserID uuid.UUID) SlotConflict {
	visibleTitle := event.Title
	visiblePriority := stringPointer(event.Priority)
	if event.OwnerUserID != requesterUserID {
		if participant.VisibilityOverride != nil && *participant.VisibilityOverride != "" {
			visibleTitle = *participant.VisibilityOverride
			visiblePriority = nil
		} else if strings.EqualFold(event.VisibilityHint, "private") {
			visibleTitle = "Busy"
			visiblePriority = nil
		}
	}
	return SlotConflict{
		ID:              uuid.New(),
		UserID:          uuidPointer(participant.UserID),
		EventID:         uuidPointer(event.ID),
		ConflictType:    eventConflictType(event.Priority),
		VisibleTitle:    visibleTitle,
		VisiblePriority: visiblePriority,
	}
}

func eventConflictType(priority string) string {
	if isHardEventConflict(priority) {
		return "critical_event_conflict"
	}
	return "event_conflict"
}

func isHardEventConflict(priority string) bool {
	return strings.EqualFold(priority, "critical")
}

func penaltyForPriority(priority string) int {
	switch strings.ToLower(strings.TrimSpace(priority)) {
	case "low":
		return softPenaltyLow
	case "medium":
		return softPenaltyMedium
	case "high":
		return softPenaltyHigh
	default:
		return softPenaltyHigh
	}
}

func hasHardConflict(conflicts []SlotConflict) bool {
	for _, conflict := range conflicts {
		switch conflict.ConflictType {
		case "outside_search_range", "outside_meeting_hours", "outside_working_hours", "unavailability", "resource_unavailable", "critical_event_conflict":
			return true
		}
	}
	return false
}

func canAccessMeeting(meeting *MeetingAggregate, userID uuid.UUID) bool {
	if meeting == nil {
		return false
	}
	if meeting.OrganizerUserID == userID {
		return true
	}
	for _, participant := range meeting.Participants {
		if participant.UserID == userID {
			return true
		}
	}
	return false
}

func newGenericConflict(conflictType, title string, userID, eventID, resourceID *uuid.UUID) SlotConflict {
	return SlotConflict{
		ID:           uuid.New(),
		UserID:       userID,
		EventID:      eventID,
		ResourceID:   resourceID,
		ConflictType: conflictType,
		VisibleTitle: title,
	}
}

func uuidPointer(value uuid.UUID) *uuid.UUID {
	copyValue := value
	return &copyValue
}

func stringPointer(value string) *string {
	if value == "" {
		return nil
	}
	copyValue := value
	return &copyValue
}

func normalizeTimePointer(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	normalized := value.UTC()
	return &normalized
}
