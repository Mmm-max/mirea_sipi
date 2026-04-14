package calendarimport

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

type fakeImportRepository struct {
	jobs map[uuid.UUID]*ImportJob
}

func newFakeImportRepository() *fakeImportRepository {
	return &fakeImportRepository{jobs: make(map[uuid.UUID]*ImportJob)}
}

func (r *fakeImportRepository) Create(_ context.Context, job *ImportJob) error {
	copyValue := *job
	r.jobs[job.ID] = &copyValue
	return nil
}

func (r *fakeImportRepository) Update(_ context.Context, job *ImportJob) error {
	copyValue := *job
	r.jobs[job.ID] = &copyValue
	return nil
}

func (r *fakeImportRepository) GetByID(_ context.Context, id uuid.UUID) (*ImportJob, error) {
	job, ok := r.jobs[id]
	if !ok {
		return nil, ErrNotFound
	}
	copyValue := *job
	return &copyValue, nil
}

func (r *fakeImportRepository) ListByUser(_ context.Context, userID uuid.UUID) ([]ImportJob, error) {
	result := make([]ImportJob, 0)
	for _, job := range r.jobs {
		if job.UserID != userID {
			continue
		}
		result = append(result, *job)
	}
	return result, nil
}

type fakeICSParser struct {
	result *ParseResult
	err    error
}

func (p *fakeICSParser) Parse(_ []byte) (*ParseResult, error) {
	if p.err != nil {
		return nil, p.err
	}
	return p.result, nil
}

type fakeImportedEventStore struct {
	seen map[string]ImportedEvent
}

func newFakeImportedEventStore() *fakeImportedEventStore {
	return &fakeImportedEventStore{seen: make(map[string]ImportedEvent)}
}

func (s *fakeImportedEventStore) UpsertImportedEvent(_ context.Context, event ImportedEvent) (bool, error) {
	key := event.OwnerUserID.String() + ":" + event.ExternalUID
	if _, exists := s.seen[key]; exists {
		s.seen[key] = event
		return false, nil
	}
	s.seen[key] = event
	return true, nil
}

func TestServiceImportICSValidFile(t *testing.T) {
	t.Parallel()

	repo := newFakeImportRepository()
	parser := &fakeICSParser{
		result: &ParseResult{
			Events: []ImportedEvent{{
				ExternalUID: "event-1",
				Title:       "Imported Event",
			}},
		},
	}
	store := newFakeImportedEventStore()
	service := NewService(repo, parser, store)

	job, err := service.ImportICS(context.Background(), ImportICSCommand{
		UserID:           uuid.New(),
		OriginalFilename: "calendar.ics",
		Content:          []byte("BEGIN:VCALENDAR\nEND:VCALENDAR"),
	})
	if err != nil {
		t.Fatalf("ImportICS() error = %v", err)
	}
	if job.Status != ImportJobStatusCompleted {
		t.Fatalf("expected completed job, got %s", job.Status)
	}
	if job.ImportedCount != 1 || job.UpdatedCount != 0 || job.SkippedCount != 0 {
		t.Fatalf("unexpected counters: %+v", job)
	}
}

func TestServiceImportICSRepeatedImportNoDuplicates(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	repo := newFakeImportRepository()
	parser := &fakeICSParser{
		result: &ParseResult{
			Events: []ImportedEvent{{
				ExternalUID: "event-1",
				Title:       "Imported Event",
			}},
		},
	}
	store := newFakeImportedEventStore()
	service := NewService(repo, parser, store)

	_, err := service.ImportICS(context.Background(), ImportICSCommand{
		UserID:           userID,
		OriginalFilename: "calendar.ics",
		Content:          []byte("BEGIN:VCALENDAR\nEND:VCALENDAR"),
	})
	if err != nil {
		t.Fatalf("first ImportICS() error = %v", err)
	}

	job, err := service.ImportICS(context.Background(), ImportICSCommand{
		UserID:           userID,
		OriginalFilename: "calendar.ics",
		Content:          []byte("BEGIN:VCALENDAR\nEND:VCALENDAR"),
	})
	if err != nil {
		t.Fatalf("second ImportICS() error = %v", err)
	}
	if job.ImportedCount != 0 || job.UpdatedCount != 1 {
		t.Fatalf("expected second import to update existing event, got %+v", job)
	}
}

func TestServiceImportICSInvalidFile(t *testing.T) {
	t.Parallel()

	repo := newFakeImportRepository()
	parser := &fakeICSParser{err: errors.New("invalid")}
	store := newFakeImportedEventStore()
	service := NewService(repo, parser, store)

	_, err := service.ImportICS(context.Background(), ImportICSCommand{
		UserID:           uuid.New(),
		OriginalFilename: "calendar.txt",
		Content:          []byte("not-ical"),
	})
	if err == nil {
		t.Fatalf("expected validation error")
	}
	if len(repo.jobs) != 1 {
		t.Fatalf("expected failed import job to be stored")
	}
	for _, job := range repo.jobs {
		if job.Status != ImportJobStatusFailed {
			t.Fatalf("expected failed job, got %s", job.Status)
		}
		if job.ErrorMessage == nil || *job.ErrorMessage == "" {
			t.Fatalf("expected error message to be set")
		}
	}
}
