package scheduling

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

type fakeSchedulingRepository struct {
	replaced []MeetingSlot
}

func (r *fakeSchedulingRepository) ReplaceSlots(_ context.Context, _ uuid.UUID, slots []MeetingSlot) error {
	r.replaced = append([]MeetingSlot(nil), slots...)
	return nil
}

func (r *fakeSchedulingRepository) ListSlotsByMeetingID(_ context.Context, _ uuid.UUID) ([]MeetingSlot, error) {
	return append([]MeetingSlot(nil), r.replaced...), nil
}

func (r *fakeSchedulingRepository) GetSlotByID(_ context.Context, id uuid.UUID) (*MeetingSlot, error) {
	for _, slot := range r.replaced {
		if slot.ID == id {
			copyValue := slot
			return &copyValue, nil
		}
	}
	return nil, ErrNotFound
}

type fakeMeetingStore struct {
	meeting *MeetingAggregate
}

func (s *fakeMeetingStore) GetMeeting(_ context.Context, _ uuid.UUID) (*MeetingAggregate, error) {
	copyValue := *s.meeting
	copyValue.Participants = append([]MeetingParticipantRef(nil), s.meeting.Participants...)
	copyValue.Resources = append([]MeetingResourceRef(nil), s.meeting.Resources...)
	return &copyValue, nil
}

func (s *fakeMeetingStore) UpdateSelectedTime(_ context.Context, _ uuid.UUID, startAt, endAt time.Time) error {
	s.meeting.SelectedStartAt = &startAt
	s.meeting.SelectedEndAt = &endAt
	return nil
}

type fakeAvailabilityStore struct {
	workingHours   map[uuid.UUID][]WorkingHoursWindow
	unavailability map[uuid.UUID][]UnavailabilityWindow
	events         map[uuid.UUID][]CalendarEvent
	resourceBusy   map[uuid.UUID][]timeRange
}

type timeRange struct {
	start time.Time
	end   time.Time
}

func (s *fakeAvailabilityStore) ListWorkingHours(_ context.Context, userID uuid.UUID) ([]WorkingHoursWindow, error) {
	return append([]WorkingHoursWindow(nil), s.workingHours[userID]...), nil
}

func (s *fakeAvailabilityStore) ListUnavailability(_ context.Context, userID uuid.UUID) ([]UnavailabilityWindow, error) {
	return append([]UnavailabilityWindow(nil), s.unavailability[userID]...), nil
}

func (s *fakeAvailabilityStore) ListEvents(_ context.Context, userID uuid.UUID, _, _ time.Time) ([]CalendarEvent, error) {
	return append([]CalendarEvent(nil), s.events[userID]...), nil
}

func (s *fakeAvailabilityStore) ResourceAvailable(_ context.Context, resourceID uuid.UUID, startAt, endAt time.Time) (bool, error) {
	for _, busy := range s.resourceBusy[resourceID] {
		if overlaps(startAt, endAt, busy.start, busy.end) {
			return false, nil
		}
	}
	return true, nil
}

type fakeEventSync struct{}

func (fakeEventSync) SyncSelectedMeeting(context.Context, *MeetingAggregate, time.Time, time.Time) error {
	return nil
}

func TestServiceSearchSlotsFindsFreeSlot(t *testing.T) {
	t.Parallel()

	organizerID := uuid.New()
	meetingID := uuid.New()
	start := time.Date(2026, 3, 27, 9, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 27, 10, 0, 0, 0, time.UTC)

	service := NewService(
		&fakeSchedulingRepository{},
		&fakeMeetingStore{meeting: &MeetingAggregate{
			ID:               meetingID,
			OrganizerUserID:  organizerID,
			Title:            "Planning",
			DurationMinutes:  60,
			SearchRangeStart: start,
			SearchRangeEnd:   end,
			Participants:     []MeetingParticipantRef{{UserID: organizerID}},
		}},
		&fakeAvailabilityStore{},
		fakeEventSync{},
		nil,
	)

	result, err := service.SearchSlots(context.Background(), SearchSlotsCommand{
		MeetingID:       meetingID,
		OrganizerUserID: organizerID,
		TopN:            1,
	})
	if err != nil {
		t.Fatalf("SearchSlots() error = %v", err)
	}
	if len(result.Slots) != 1 {
		t.Fatalf("expected 1 slot, got %d", len(result.Slots))
	}
	if result.Slots[0].Score != 100 {
		t.Fatalf("expected free slot score 100, got %d", result.Slots[0].Score)
	}
	if len(result.Slots[0].Conflicts) != 0 {
		t.Fatalf("expected free slot without conflicts")
	}
}

func TestServiceSearchSlotsHardConflictExcludedFromTopResult(t *testing.T) {
	t.Parallel()

	organizerID := uuid.New()
	meetingID := uuid.New()
	searchStart := time.Date(2026, 3, 27, 9, 0, 0, 0, time.UTC)

	service := NewService(
		&fakeSchedulingRepository{},
		&fakeMeetingStore{meeting: &MeetingAggregate{
			ID:               meetingID,
			OrganizerUserID:  organizerID,
			Title:            "Planning",
			DurationMinutes:  15,
			SearchRangeStart: searchStart,
			SearchRangeEnd:   searchStart.Add(30 * time.Minute),
			Participants:     []MeetingParticipantRef{{UserID: organizerID}},
		}},
		&fakeAvailabilityStore{
			unavailability: map[uuid.UUID][]UnavailabilityWindow{
				organizerID: {{
					Title:   "Busy",
					StartAt: searchStart,
					EndAt:   searchStart.Add(15 * time.Minute),
				}},
			},
		},
		fakeEventSync{},
		nil,
	)

	result, err := service.SearchSlots(context.Background(), SearchSlotsCommand{
		MeetingID:       meetingID,
		OrganizerUserID: organizerID,
		TopN:            1,
	})
	if err != nil {
		t.Fatalf("SearchSlots() error = %v", err)
	}
	if len(result.Slots) != 1 {
		t.Fatalf("expected 1 slot, got %d", len(result.Slots))
	}
	if !result.Slots[0].StartAt.Equal(searchStart.Add(15 * time.Minute)) {
		t.Fatalf("expected non-conflicting slot to be ranked first, got %s", result.Slots[0].StartAt)
	}
}

func TestServiceSearchSlotsLowPriorityRanksAboveHighPriority(t *testing.T) {
	t.Parallel()

	organizerID := uuid.New()
	meetingID := uuid.New()
	searchStart := time.Date(2026, 3, 27, 9, 0, 0, 0, time.UTC)

	service := NewService(
		&fakeSchedulingRepository{},
		&fakeMeetingStore{meeting: &MeetingAggregate{
			ID:               meetingID,
			OrganizerUserID:  organizerID,
			Title:            "Planning",
			DurationMinutes:  15,
			SearchRangeStart: searchStart,
			SearchRangeEnd:   searchStart.Add(30 * time.Minute),
			Participants:     []MeetingParticipantRef{{UserID: organizerID}},
		}},
		&fakeAvailabilityStore{
			events: map[uuid.UUID][]CalendarEvent{
				organizerID: {
					{
						ID:          uuid.New(),
						OwnerUserID: organizerID,
						Title:       "Low priority task",
						StartAt:     searchStart,
						EndAt:       searchStart.Add(15 * time.Minute),
						Priority:    "low",
					},
					{
						ID:          uuid.New(),
						OwnerUserID: organizerID,
						Title:       "High priority task",
						StartAt:     searchStart.Add(15 * time.Minute),
						EndAt:       searchStart.Add(30 * time.Minute),
						Priority:    "high",
					},
				},
			},
		},
		fakeEventSync{},
		nil,
	)

	result, err := service.SearchSlots(context.Background(), SearchSlotsCommand{
		MeetingID:       meetingID,
		OrganizerUserID: organizerID,
		TopN:            2,
	})
	if err != nil {
		t.Fatalf("SearchSlots() error = %v", err)
	}
	if len(result.Slots) != 2 {
		t.Fatalf("expected 2 slots, got %d", len(result.Slots))
	}
	if result.Slots[0].Score <= result.Slots[1].Score {
		t.Fatalf("expected low-priority conflict slot to rank above high-priority slot: %+v", result.Slots)
	}
	if !result.Slots[0].StartAt.Equal(searchStart) {
		t.Fatalf("expected low-priority slot first, got %s", result.Slots[0].StartAt)
	}
}
