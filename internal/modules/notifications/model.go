package notifications

import (
	"sipi/internal/platform/db"

	"github.com/google/uuid"
)

type NotificationModel struct {
	db.BaseModel
	UserID       uuid.UUID `gorm:"type:uuid;not null;index"`
	Type         string    `gorm:"size:64;not null;index"`
	Title        string    `gorm:"size:255;not null"`
	Body         string    `gorm:"type:text;not null"`
	IsRead       bool      `gorm:"not null;default:false;index"`
	MetadataJSON []byte    `gorm:"type:jsonb"`
}

func (NotificationModel) TableName() string {
	return "notifications"
}
