package scheduling

import (
	"sipi/internal/platform/db"
	"time"

	"github.com/google/uuid"
)

type MeetingSlotModel struct {
	db.BaseModel
	MeetingID uuid.UUID `gorm:"type:uuid;not null;index"`
	StartAt   time.Time `gorm:"column:start_at;not null;index"`
	EndAt     time.Time `gorm:"column:end_at;not null;index"`
	Score     int       `gorm:"not null"`
	Rank      int       `gorm:"not null"`
}

func (MeetingSlotModel) TableName() string {
	return "meeting_slots"
}

type SlotConflictModel struct {
	db.BaseModel
	MeetingSlotID   uuid.UUID  `gorm:"type:uuid;not null;index"`
	UserID          *uuid.UUID `gorm:"type:uuid;index"`
	EventID         *uuid.UUID `gorm:"type:uuid;index"`
	ResourceID      *uuid.UUID `gorm:"type:uuid;index"`
	ConflictType    string     `gorm:"size:64;not null;index"`
	VisibleTitle    string     `gorm:"size:255;not null"`
	VisiblePriority *string    `gorm:"size:16"`
}

func (SlotConflictModel) TableName() string {
	return "slot_conflicts"
}
