package calendarimport

import (
	"time"

	"github.com/google/uuid"
)

type ImportedEvent struct {
	OwnerUserID     uuid.UUID
	ExternalUID     string
	Title           string
	Description     string
	StartAt         time.Time
	EndAt           time.Time
	Priority        string
	IsReschedulable bool
	VisibilityHint  string
}

type ParseResult struct {
	Events       []ImportedEvent
	SkippedCount int
}
