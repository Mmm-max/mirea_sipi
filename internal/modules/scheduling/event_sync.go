package scheduling

import (
	"context"
	"time"
)

type NoopEventSync struct{}

func NewNoopEventSync() *NoopEventSync {
	return &NoopEventSync{}
}

func (s *NoopEventSync) SyncSelectedMeeting(_ context.Context, _ *MeetingAggregate, _, _ time.Time) error {
	return nil
}
