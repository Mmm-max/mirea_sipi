package events

import (
	"strings"
	"time"

	"sipi/internal/platform/db"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EventModel struct {
	db.BaseModel
	db.SoftDeleteModel
	OwnerUserID     uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_events_owner_source_external_uid,priority:1"`
	SourceType      string    `gorm:"size:32;not null;index;uniqueIndex:idx_events_owner_source_external_uid,priority:2"`
	ExternalUID     *string   `gorm:"size:255;index;uniqueIndex:idx_events_owner_source_external_uid,priority:3"`
	Title           string    `gorm:"size:255;not null"`
	Description     string    `gorm:"type:text;not null;default:''"`
	StartAt         time.Time `gorm:"not null;index"`
	EndAt           time.Time `gorm:"not null;index"`
	Priority        string    `gorm:"size:16;not null"`
	IsReschedulable bool      `gorm:"not null;default:true"`
	VisibilityHint  string    `gorm:"size:64;not null;default:'default'"`
}

func (EventModel) TableName() string {
	return "events"
}

func (m *EventModel) BeforeSave(_ *gorm.DB) error {
	m.Title = strings.TrimSpace(m.Title)
	m.Description = strings.TrimSpace(m.Description)
	m.VisibilityHint = strings.TrimSpace(m.VisibilityHint)
	if m.ExternalUID != nil {
		value := strings.TrimSpace(*m.ExternalUID)
		if value == "" {
			m.ExternalUID = nil
		} else {
			m.ExternalUID = &value
		}
	}

	return nil
}
