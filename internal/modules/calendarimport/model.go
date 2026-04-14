package calendarimport

import (
	"strings"

	"sipi/internal/platform/db"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ImportJobModel struct {
	db.BaseModel
	UserID           uuid.UUID `gorm:"type:uuid;not null;index"`
	OriginalFilename string    `gorm:"size:255;not null"`
	Status           string    `gorm:"size:32;not null;index"`
	ImportedCount    int       `gorm:"not null;default:0"`
	UpdatedCount     int       `gorm:"not null;default:0"`
	SkippedCount     int       `gorm:"not null;default:0"`
	ErrorMessage     *string   `gorm:"type:text"`
}

func (ImportJobModel) TableName() string {
	return "import_jobs"
}

func (m *ImportJobModel) BeforeSave(_ *gorm.DB) error {
	m.OriginalFilename = strings.TrimSpace(m.OriginalFilename)
	if m.ErrorMessage != nil {
		value := strings.TrimSpace(*m.ErrorMessage)
		if value == "" {
			m.ErrorMessage = nil
		} else {
			m.ErrorMessage = &value
		}
	}

	return nil
}
