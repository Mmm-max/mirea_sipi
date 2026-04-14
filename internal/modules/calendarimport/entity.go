package calendarimport

import (
	"time"

	"github.com/google/uuid"
)

type ImportJobStatus string

const (
	ImportJobStatusProcessing ImportJobStatus = "processing"
	ImportJobStatusCompleted  ImportJobStatus = "completed"
	ImportJobStatusFailed     ImportJobStatus = "failed"
)

type ImportJob struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	OriginalFilename string
	Status           ImportJobStatus
	ImportedCount    int
	UpdatedCount     int
	SkippedCount     int
	ErrorMessage     *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
