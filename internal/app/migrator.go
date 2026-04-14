package app

import (
	"fmt"

	"sipi/internal/modules/auth"
	"sipi/internal/modules/calendarimport"
	"sipi/internal/modules/events"
	"sipi/internal/modules/meetings"
	"sipi/internal/modules/notifications"
	"sipi/internal/modules/resources"
	"sipi/internal/modules/scheduling"
	"sipi/internal/modules/users"

	"gorm.io/gorm"
)

func migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&users.UserModel{},
		&users.WorkingHoursModel{},
		&users.UnavailabilityPeriodModel{},
		&events.EventModel{},
		&meetings.MeetingModel{},
		&meetings.MeetingParticipantModel{},
		&meetings.MeetingResourceModel{},
		&scheduling.MeetingSlotModel{},
		&scheduling.SlotConflictModel{},
		&resources.ResourceModel{},
		&resources.ResourceBookingModel{},
		&notifications.NotificationModel{},
		&calendarimport.ImportJobModel{},
		&auth.RefreshSessionModel{},
	); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}

	return nil
}
