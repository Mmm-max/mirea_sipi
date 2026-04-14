package calendarimport

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

const MaxICSFileSizeBytes = 5 << 20

type Service struct {
	repository RepositoryPort
	parser     ParserPort
	eventStore EventStorePort
}

func NewService(repository RepositoryPort, parser ParserPort, eventStore EventStorePort) *Service {
	return &Service{
		repository: repository,
		parser:     parser,
		eventStore: eventStore,
	}
}

func (s *Service) ImportICS(ctx context.Context, command ImportICSCommand) (*ImportJob, error) {
	filename := strings.TrimSpace(command.OriginalFilename)
	if filename == "" {
		return nil, newServiceError(ErrCodeValidation, "file is required", nil)
	}

	job := &ImportJob{
		ID:               uuid.New(),
		UserID:           command.UserID,
		OriginalFilename: filename,
		Status:           ImportJobStatusProcessing,
	}
	if err := s.repository.Create(ctx, job); err != nil {
		return nil, newServiceError(ErrCodeInternal, "failed to create import job", err)
	}

	failJob := func(message string, err error, validation bool) (*ImportJob, error) {
		job.Status = ImportJobStatusFailed
		job.ErrorMessage = stringPointer(message)
		if updateErr := s.repository.Update(ctx, job); updateErr != nil {
			return nil, newServiceError(ErrCodeInternal, "failed to update import job", updateErr)
		}

		if validation {
			return nil, newServiceError(ErrCodeValidation, message, err)
		}

		return nil, newServiceError(ErrCodeInternal, message, err)
	}

	if err := validateImportCommand(command); err != nil {
		return failJob(err.Error(), err, true)
	}

	parseResult, err := s.parser.Parse(command.Content)
	if err != nil {
		return failJob("invalid iCalendar file", err, true)
	}

	job.SkippedCount += parseResult.SkippedCount

	for i := range parseResult.Events {
		importedEvent := parseResult.Events[i]
		importedEvent.OwnerUserID = command.UserID

		created, upsertErr := s.eventStore.UpsertImportedEvent(ctx, importedEvent)
		if upsertErr != nil {
			return failJob("failed to persist imported event", upsertErr, false)
		}
		if created {
			job.ImportedCount++
			continue
		}
		job.UpdatedCount++
	}

	job.Status = ImportJobStatusCompleted
	job.ErrorMessage = nil
	if err := s.repository.Update(ctx, job); err != nil {
		return nil, newServiceError(ErrCodeInternal, "failed to finalize import job", err)
	}

	return job, nil
}

func (s *Service) ListHistory(ctx context.Context, query ListImportJobsQuery) ([]ImportJob, error) {
	jobs, err := s.repository.ListByUser(ctx, query.UserID)
	if err != nil {
		return nil, newServiceError(ErrCodeInternal, "failed to list import history", err)
	}

	return jobs, nil
}

func (s *Service) GetHistoryByID(ctx context.Context, query GetImportJobQuery) (*ImportJob, error) {
	job, err := s.repository.GetByID(ctx, query.ID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, newServiceError(ErrCodeNotFound, "import job not found", err)
		}
		return nil, newServiceError(ErrCodeInternal, "failed to get import job", err)
	}

	if job.UserID != query.UserID {
		return nil, newServiceError(ErrCodeNotFound, "import job not found", nil)
	}

	return job, nil
}

func validateImportCommand(command ImportICSCommand) error {
	if command.UserID == uuid.Nil {
		return fmt.Errorf("user_id is required")
	}
	if len(command.Content) == 0 {
		return fmt.Errorf("file is empty")
	}
	if len(command.Content) > MaxICSFileSizeBytes {
		return fmt.Errorf("file is too large")
	}
	if !strings.EqualFold(filepath.Ext(strings.TrimSpace(command.OriginalFilename)), ".ics") {
		return fmt.Errorf("file must have .ics extension")
	}

	content := strings.ToUpper(string(command.Content))
	if !strings.Contains(content, "BEGIN:VCALENDAR") || !strings.Contains(content, "END:VCALENDAR") {
		return fmt.Errorf("file content is not a valid iCalendar payload")
	}

	return nil
}
