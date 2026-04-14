package users

import (
	"time"

	"github.com/google/uuid"
)

type UpdateProfileRequest struct {
	FullName string `json:"full_name" binding:"omitempty,min=2,max=255" example:"Alice Johnson"`
	Timezone string `json:"timezone" binding:"omitempty,min=1,max=64" example:"Europe/Moscow"`
}

type ProfileResponse struct {
	ID        string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email     string `json:"email" example:"alice@example.com"`
	FullName  string `json:"full_name" example:"Alice Johnson"`
	Timezone  string `json:"timezone" example:"Europe/Moscow"`
	CreatedAt string `json:"created_at" example:"2026-03-26T12:00:00Z"`
	UpdatedAt string `json:"updated_at" example:"2026-03-26T13:00:00Z"`
}

type ProfileEnvelope struct {
	Success bool            `json:"success" example:"true"`
	Data    ProfileResponse `json:"data"`
	Meta    ResponseMetaDTO `json:"meta"`
}

type WorkingHoursItemRequest struct {
	Weekday      int    `json:"weekday" binding:"required,min=0,max=6" example:"1"`
	StartTime    string `json:"start_time" example:"09:00"`
	EndTime      string `json:"end_time" example:"18:00"`
	IsWorkingDay bool   `json:"is_working_day" example:"true"`
}

type ReplaceWorkingHoursRequest struct {
	Items []WorkingHoursItemRequest `json:"items" binding:"required,min=1,max=7,dive"`
}

type WorkingHoursResponse struct {
	ID           string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Weekday      int    `json:"weekday" example:"1"`
	StartTime    string `json:"start_time" example:"09:00"`
	EndTime      string `json:"end_time" example:"18:00"`
	IsWorkingDay bool   `json:"is_working_day" example:"true"`
}

type WorkingHoursListResponse struct {
	Items []WorkingHoursResponse `json:"items"`
}

type WorkingHoursEnvelope struct {
	Success bool                     `json:"success" example:"true"`
	Data    WorkingHoursListResponse `json:"data"`
	Meta    ResponseMetaDTO          `json:"meta"`
}

type CreateUnavailabilityRequest struct {
	Type    string `json:"type" binding:"required,oneof=vacation sick_leave business_trip custom" example:"vacation"`
	Title   string `json:"title" binding:"required,min=1,max=255" example:"Summer Vacation"`
	StartAt string `json:"start_at" binding:"required" example:"2026-07-01T00:00:00Z"`
	EndAt   string `json:"end_at" binding:"required" example:"2026-07-10T23:59:59Z"`
	Comment string `json:"comment" binding:"omitempty,max=1000" example:"Out of office"`
}

type UnavailabilityResponse struct {
	ID        string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Type      string `json:"type" example:"vacation"`
	Title     string `json:"title" example:"Summer Vacation"`
	StartAt   string `json:"start_at" example:"2026-07-01T00:00:00Z"`
	EndAt     string `json:"end_at" example:"2026-07-10T23:59:59Z"`
	Comment   string `json:"comment" example:"Out of office"`
	CreatedAt string `json:"created_at" example:"2026-03-26T12:00:00Z"`
	UpdatedAt string `json:"updated_at" example:"2026-03-26T12:00:00Z"`
}

type UnavailabilityListResponse struct {
	Items []UnavailabilityResponse `json:"items"`
}

type UnavailabilityEnvelope struct {
	Success bool                   `json:"success" example:"true"`
	Data    UnavailabilityResponse `json:"data"`
	Meta    ResponseMetaDTO        `json:"meta"`
}

type UnavailabilityListEnvelope struct {
	Success bool                       `json:"success" example:"true"`
	Data    UnavailabilityListResponse `json:"data"`
	Meta    ResponseMetaDTO            `json:"meta"`
}

type MessageData struct {
	Message string `json:"message" example:"operation completed successfully"`
}

type MessageEnvelope struct {
	Success bool            `json:"success" example:"true"`
	Data    MessageData     `json:"data"`
	Meta    ResponseMetaDTO `json:"meta"`
}

type ResponseMetaDTO struct {
	RequestID string `json:"request_id" example:"e7cc5d4a-3d85-4938-bff3-d2a6cf8ca2bb"`
}

func (r UpdateProfileRequest) ToCommand(userID uuid.UUID) UpdateProfileCommand {
	return UpdateProfileCommand{
		UserID:   userID,
		FullName: r.FullName,
		Timezone: r.Timezone,
	}
}

func (r ReplaceWorkingHoursRequest) ToCommand(userID uuid.UUID) ReplaceWorkingHoursCommand {
	items := make([]WorkingHoursInput, 0, len(r.Items))
	for _, item := range r.Items {
		items = append(items, WorkingHoursInput{
			Weekday:      item.Weekday,
			StartTime:    item.StartTime,
			EndTime:      item.EndTime,
			IsWorkingDay: item.IsWorkingDay,
		})
	}

	return ReplaceWorkingHoursCommand{
		UserID: userID,
		Items:  items,
	}
}

func (r CreateUnavailabilityRequest) ToCommand(userID uuid.UUID) (CreateUnavailabilityCommand, error) {
	startAt, err := time.Parse(time.RFC3339, r.StartAt)
	if err != nil {
		return CreateUnavailabilityCommand{}, err
	}

	endAt, err := time.Parse(time.RFC3339, r.EndAt)
	if err != nil {
		return CreateUnavailabilityCommand{}, err
	}

	return CreateUnavailabilityCommand{
		UserID:  userID,
		Type:    UnavailabilityType(r.Type),
		Title:   r.Title,
		StartAt: startAt,
		EndAt:   endAt,
		Comment: r.Comment,
	}, nil
}
