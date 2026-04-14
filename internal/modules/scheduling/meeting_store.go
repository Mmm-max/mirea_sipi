package scheduling

import (
	"context"
	"errors"
	"time"

	"sipi/internal/modules/meetings"

	"github.com/google/uuid"
)

type meetingReaderWriter interface {
	GetMeetingByID(ctx context.Context, id uuid.UUID) (*meetings.Meeting, error)
	UpdateMeeting(ctx context.Context, meeting *meetings.Meeting) error
}

type MeetingStore struct {
	repository meetingReaderWriter
}

func NewMeetingStore(repository meetingReaderWriter) *MeetingStore {
	return &MeetingStore{repository: repository}
}

func (s *MeetingStore) GetMeeting(ctx context.Context, id uuid.UUID) (*MeetingAggregate, error) {
	meeting, err := s.repository.GetMeetingByID(ctx, id)
	if err != nil {
		if errors.Is(err, meetings.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	result := &MeetingAggregate{
		ID:                meeting.ID,
		OrganizerUserID:   meeting.OrganizerUserID,
		Title:             meeting.Title,
		DurationMinutes:   meeting.DurationMinutes,
		SearchRangeStart:  meeting.SearchRangeStart,
		SearchRangeEnd:    meeting.SearchRangeEnd,
		EarliestStartTime: meeting.EarliestStartTime,
		LatestStartTime:   meeting.LatestStartTime,
		SelectedStartAt:   meeting.SelectedStartAt,
		SelectedEndAt:     meeting.SelectedEndAt,
	}
	for _, participant := range meeting.Participants {
		result.Participants = append(result.Participants, MeetingParticipantRef{
			UserID:             participant.UserID,
			VisibilityOverride: participant.VisibilityOverride,
		})
	}
	for _, resource := range meeting.Resources {
		result.Resources = append(result.Resources, MeetingResourceRef{ResourceID: resource.ResourceID})
	}
	return result, nil
}

func (s *MeetingStore) UpdateSelectedTime(ctx context.Context, meetingID uuid.UUID, startAt, endAt time.Time) error {
	meeting, err := s.repository.GetMeetingByID(ctx, meetingID)
	if err != nil {
		if errors.Is(err, meetings.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}
	meeting.SelectedStartAt = normalizeTimePointer(&startAt)
	meeting.SelectedEndAt = normalizeTimePointer(&endAt)
	meeting.Status = meetings.MeetingStatusScheduled
	return s.repository.UpdateMeeting(ctx, meeting)
}
