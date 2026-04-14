package resources

import (
	"time"

	"github.com/google/uuid"
)

type CreateResourceRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=255" example:"Room A"`
	Type        string  `json:"type" binding:"required,oneof=meeting_room equipment shared_space other" example:"meeting_room"`
	Description string  `json:"description" binding:"omitempty,max=5000" example:"Main meeting room on floor 3"`
	Capacity    *int    `json:"capacity,omitempty" example:"8"`
	Location    *string `json:"location,omitempty" example:"Floor 3"`
}

type UpdateResourceRequest struct {
	Name          *string `json:"name,omitempty" binding:"omitempty,min=1,max=255" example:"Room A Updated"`
	Type          *string `json:"type,omitempty" binding:"omitempty,oneof=meeting_room equipment shared_space other" example:"shared_space"`
	Description   *string `json:"description,omitempty" binding:"omitempty,max=5000" example:"Updated description"`
	Capacity      *int    `json:"capacity,omitempty" example:"10"`
	ClearCapacity *bool   `json:"clear_capacity,omitempty" example:"false"`
	Location      *string `json:"location,omitempty" example:"Floor 4"`
	ClearLocation *bool   `json:"clear_location,omitempty" example:"false"`
}

type CreateBookingRequest struct {
	EventID *string `json:"event_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	StartAt string  `json:"start_at" binding:"required" example:"2026-03-27T09:00:00Z"`
	EndAt   string  `json:"end_at" binding:"required" example:"2026-03-27T10:00:00Z"`
	Title   string  `json:"title" binding:"required,min=1,max=255" example:"Sprint planning room booking"`
}

type ResourceResponse struct {
	ID          string  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name        string  `json:"name" example:"Room A"`
	Type        string  `json:"type" example:"meeting_room"`
	Description string  `json:"description" example:"Main meeting room on floor 3"`
	Capacity    *int    `json:"capacity,omitempty" example:"8"`
	Location    *string `json:"location,omitempty" example:"Floor 3"`
	OwnerUserID *string `json:"owner_user_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	CreatedAt   string  `json:"created_at" example:"2026-03-26T12:00:00Z"`
	UpdatedAt   string  `json:"updated_at" example:"2026-03-26T12:00:00Z"`
}

type BookingResponse struct {
	ID             string  `json:"id" example:"550e8400-e29b-41d4-a716-446655440001"`
	ResourceID     string  `json:"resource_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	EventID        *string `json:"event_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440002"`
	BookedByUserID string  `json:"booked_by_user_id" example:"550e8400-e29b-41d4-a716-446655440003"`
	StartAt        string  `json:"start_at" example:"2026-03-27T09:00:00Z"`
	EndAt          string  `json:"end_at" example:"2026-03-27T10:00:00Z"`
	Title          string  `json:"title" example:"Sprint planning room booking"`
	CreatedAt      string  `json:"created_at" example:"2026-03-26T12:00:00Z"`
	UpdatedAt      string  `json:"updated_at" example:"2026-03-26T12:00:00Z"`
}

type AvailabilityResponse struct {
	ResourceID  string            `json:"resource_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	DateFrom    string            `json:"date_from" example:"2026-03-27T00:00:00Z"`
	DateTo      string            `json:"date_to" example:"2026-03-28T00:00:00Z"`
	IsAvailable bool              `json:"is_available" example:"false"`
	Bookings    []BookingResponse `json:"bookings"`
}

type ResourceEnvelope struct {
	Success bool             `json:"success" example:"true"`
	Data    ResourceResponse `json:"data"`
	Meta    ResponseMetaDTO  `json:"meta"`
}

type ResourceListResponse struct {
	Items []ResourceResponse `json:"items"`
}

type ResourceListEnvelope struct {
	Success bool                 `json:"success" example:"true"`
	Data    ResourceListResponse `json:"data"`
	Meta    ResponseMetaDTO      `json:"meta"`
}

type BookingEnvelope struct {
	Success bool            `json:"success" example:"true"`
	Data    BookingResponse `json:"data"`
	Meta    ResponseMetaDTO `json:"meta"`
}

type AvailabilityEnvelope struct {
	Success bool                 `json:"success" example:"true"`
	Data    AvailabilityResponse `json:"data"`
	Meta    ResponseMetaDTO      `json:"meta"`
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

func (r CreateResourceRequest) ToCommand(ownerUserID uuid.UUID) CreateResourceCommand {
	return CreateResourceCommand{
		Name:        r.Name,
		Type:        ResourceType(r.Type),
		Description: r.Description,
		Capacity:    r.Capacity,
		Location:    r.Location,
		OwnerUserID: &ownerUserID,
	}
}

func (r UpdateResourceRequest) ToCommand(id uuid.UUID) UpdateResourceCommand {
	command := UpdateResourceCommand{
		ID:          id,
		Name:        r.Name,
		Description: r.Description,
		Capacity:    r.Capacity,
		Location:    r.Location,
	}
	if r.Type != nil {
		value := ResourceType(*r.Type)
		command.Type = &value
	}
	if r.ClearCapacity != nil {
		command.ClearCapacity = *r.ClearCapacity
		command.CapacitySet = command.CapacitySet || *r.ClearCapacity
	}
	if r.Capacity != nil {
		command.CapacitySet = true
	}
	if r.ClearLocation != nil {
		command.ClearLocation = *r.ClearLocation
		command.LocationSet = command.LocationSet || *r.ClearLocation
	}
	if r.Location != nil {
		command.LocationSet = true
	}

	return command
}

func (r CreateBookingRequest) ToCommand(resourceID, userID uuid.UUID) (CreateBookingCommand, error) {
	startAt, err := time.Parse(time.RFC3339, r.StartAt)
	if err != nil {
		return CreateBookingCommand{}, err
	}
	endAt, err := time.Parse(time.RFC3339, r.EndAt)
	if err != nil {
		return CreateBookingCommand{}, err
	}

	var eventID *uuid.UUID
	if r.EventID != nil && *r.EventID != "" {
		parsed, err := uuid.Parse(*r.EventID)
		if err != nil {
			return CreateBookingCommand{}, err
		}
		eventID = &parsed
	}

	return CreateBookingCommand{
		ResourceID:     resourceID,
		BookedByUserID: userID,
		EventID:        eventID,
		StartAt:        startAt,
		EndAt:          endAt,
		Title:          r.Title,
	}, nil
}
