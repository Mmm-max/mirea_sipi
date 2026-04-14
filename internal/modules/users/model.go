package users

import (
	"strings"
	"time"

	"sipi/internal/platform/db"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserModel struct {
	db.BaseModel
	Email        string `gorm:"size:320;not null;uniqueIndex" json:"email"`
	FullName     string `gorm:"size:255;not null;default:''" json:"full_name"`
	Timezone     string `gorm:"size:64;not null;default:'UTC'" json:"timezone"`
	PasswordHash string `gorm:"size:512;not null" json:"-"`
}

func (UserModel) TableName() string {
	return "users"
}

func (u *UserModel) BeforeSave(_ *gorm.DB) error {
	u.Email = strings.TrimSpace(strings.ToLower(u.Email))
	u.FullName = strings.TrimSpace(u.FullName)
	u.Timezone = strings.TrimSpace(u.Timezone)
	return nil
}

type WorkingHoursModel struct {
	db.BaseModel
	UserID       uuid.UUID `gorm:"type:uuid;not null;index:idx_working_hours_user_weekday,priority:1"`
	Weekday      int       `gorm:"not null;index:idx_working_hours_user_weekday,priority:2"`
	StartTime    string    `gorm:"size:8;not null"`
	EndTime      string    `gorm:"size:8;not null"`
	IsWorkingDay bool      `gorm:"not null;default:true"`
}

func (WorkingHoursModel) TableName() string {
	return "working_hours"
}

type UnavailabilityPeriodModel struct {
	db.BaseModel
	UserID  uuid.UUID `gorm:"type:uuid;not null;index"`
	Type    string    `gorm:"size:32;not null"`
	Title   string    `gorm:"size:255;not null"`
	StartAt time.Time `gorm:"not null;index"`
	EndAt   time.Time `gorm:"not null;index"`
	Comment string    `gorm:"size:1000;not null;default:''"`
}

func (UnavailabilityPeriodModel) TableName() string {
	return "unavailability_periods"
}
