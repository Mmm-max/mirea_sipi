package resources

import (
	"strings"
	"time"

	"sipi/internal/platform/db"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ResourceModel struct {
	db.BaseModel
	db.SoftDeleteModel
	Name        string     `gorm:"size:255;not null"`
	Type        string     `gorm:"size:32;not null;index"`
	Description string     `gorm:"type:text;not null;default:''"`
	Capacity    *int       `gorm:"default:null"`
	Location    *string    `gorm:"size:255"`
	OwnerUserID *uuid.UUID `gorm:"type:uuid;index"`
}

func (ResourceModel) TableName() string {
	return "resources"
}

func (m *ResourceModel) BeforeSave(_ *gorm.DB) error {
	m.Name = strings.TrimSpace(m.Name)
	m.Description = strings.TrimSpace(m.Description)
	if m.Location != nil {
		value := strings.TrimSpace(*m.Location)
		if value == "" {
			m.Location = nil
		} else {
			m.Location = &value
		}
	}

	return nil
}

type ResourceBookingModel struct {
	db.BaseModel
	db.SoftDeleteModel
	ResourceID     uuid.UUID  `gorm:"type:uuid;not null;index"`
	EventID        *uuid.UUID `gorm:"type:uuid;index"`
	BookedByUserID uuid.UUID  `gorm:"type:uuid;not null;index"`
	StartAt        time.Time  `gorm:"not null;index"`
	EndAt          time.Time  `gorm:"not null;index"`
	Title          string     `gorm:"size:255;not null"`
}

func (ResourceBookingModel) TableName() string {
	return "resource_bookings"
}

func (m *ResourceBookingModel) BeforeSave(_ *gorm.DB) error {
	m.Title = strings.TrimSpace(m.Title)
	return nil
}
