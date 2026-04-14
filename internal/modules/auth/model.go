package auth

import (
	"strings"
	"time"

	"sipi/internal/platform/db"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AccountModel struct {
	db.BaseModel
	Email        string `gorm:"size:320;not null;uniqueIndex"`
	FullName     string `gorm:"size:255;not null;default:''"`
	Timezone     string `gorm:"size:64;not null;default:'UTC'"`
	PasswordHash string `gorm:"size:512;not null"`
}

func (AccountModel) TableName() string {
	return "users"
}

func (m *AccountModel) BeforeSave(_ *gorm.DB) error {
	m.Email = strings.TrimSpace(strings.ToLower(m.Email))
	m.FullName = strings.TrimSpace(m.FullName)
	m.Timezone = strings.TrimSpace(m.Timezone)
	return nil
}

type RefreshSessionModel struct {
	db.BaseModel
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	TokenHash string     `gorm:"size:255;not null;uniqueIndex" json:"-"`
	ExpiresAt time.Time  `gorm:"not null" json:"expires_at"`
	RevokedAt *time.Time `gorm:"index" json:"revoked_at,omitempty"`
	UserAgent string     `gorm:"size:512" json:"user_agent,omitempty"`
	IPAddress string     `gorm:"size:64" json:"ip_address,omitempty"`
}

func (RefreshSessionModel) TableName() string {
	return "refresh_sessions"
}
