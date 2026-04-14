package meetings

import (
	"strings"
	"time"

	"sipi/internal/platform/db"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MeetingModel struct {
	db.BaseModel
	db.SoftDeleteModel
	OrganizerUserID   uuid.UUID  `gorm:"type:uuid;not null;index"`
	Title             string     `gorm:"size:255;not null"`
	Description       string     `gorm:"type:text;not null;default:''"`
	DurationMinutes   int        `gorm:"not null"`
	Priority          string     `gorm:"size:16;not null"`
	Status            string     `gorm:"size:32;not null;index"`
	SearchRangeStart  time.Time  `gorm:"not null;index"`
	SearchRangeEnd    time.Time  `gorm:"not null;index"`
	EarliestStartTime *string    `gorm:"size:8"`
	LatestStartTime   *string    `gorm:"size:8"`
	RecurrenceRule    *string    `gorm:"size:512"`
	SelectedStartAt   *time.Time `gorm:"index"`
	SelectedEndAt     *time.Time `gorm:"index"`
}

func (MeetingModel) TableName() string {
	return "meetings"
}

func (m *MeetingModel) BeforeSave(_ *gorm.DB) error {
	m.Title = strings.TrimSpace(m.Title)
	m.Description = strings.TrimSpace(m.Description)
	m.EarliestStartTime = normalizeOptionalString(m.EarliestStartTime)
	m.LatestStartTime = normalizeOptionalString(m.LatestStartTime)
	m.RecurrenceRule = normalizeOptionalString(m.RecurrenceRule)
	return nil
}

type MeetingParticipantModel struct {
	db.BaseModel
	MeetingID          uuid.UUID  `gorm:"type:uuid;not null;index;uniqueIndex:idx_meeting_participant_user,priority:1"`
	UserID             uuid.UUID  `gorm:"type:uuid;not null;index;uniqueIndex:idx_meeting_participant_user,priority:2"`
	ResponseStatus     string     `gorm:"size:16;not null;default:'pending'"`
	VisibilityOverride *string    `gorm:"size:64"`
	AlternativeStartAt *time.Time `gorm:"index"`
	AlternativeEndAt   *time.Time `gorm:"index"`
	AlternativeComment *string    `gorm:"size:1000"`
}

func (MeetingParticipantModel) TableName() string {
	return "meeting_participants"
}

func (m *MeetingParticipantModel) BeforeSave(_ *gorm.DB) error {
	m.VisibilityOverride = normalizeOptionalString(m.VisibilityOverride)
	m.AlternativeComment = normalizeOptionalString(m.AlternativeComment)
	return nil
}

type MeetingResourceModel struct {
	db.BaseModel
	MeetingID  uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_meeting_resource_unique,priority:1"`
	ResourceID uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_meeting_resource_unique,priority:2"`
}

func (MeetingResourceModel) TableName() string {
	return "meeting_resources"
}
