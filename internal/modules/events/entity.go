package events

import (
	"time"

	"github.com/google/uuid"
)

type SourceType string

const (
	SourceTypePersonalTask SourceType = "personal_task"
	SourceTypeMeeting      SourceType = "meeting"
	SourceTypeImported     SourceType = "imported"
	SourceTypeBlockedTime  SourceType = "blocked_time"
)

type Priority string

const (
	PriorityLow      Priority = "low"
	PriorityMedium   Priority = "medium"
	PriorityHigh     Priority = "high"
	PriorityCritical Priority = "critical"
)

type Event struct {
	ID              uuid.UUID
	OwnerUserID     uuid.UUID
	SourceType      SourceType
	ExternalUID     *string
	Title           string
	Description     string
	StartAt         time.Time
	EndAt           time.Time
	Priority        Priority
	IsReschedulable bool
	VisibilityHint  string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}
