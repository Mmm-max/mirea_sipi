package notifications

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type NotificationResponse struct {
	ID           string         `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Type         string         `json:"type" example:"meeting_invitation"`
	Title        string         `json:"title" example:"New meeting invitation"`
	Body         string         `json:"body" example:"You were invited to Sprint Planning"`
	IsRead       bool           `json:"is_read" example:"false"`
	MetadataJSON map[string]any `json:"metadata_json,omitempty" swaggertype:"object"`
	CreatedAt    string         `json:"created_at" example:"2026-03-26T12:00:00Z"`
	UpdatedAt    string         `json:"updated_at" example:"2026-03-26T12:00:00Z"`
}

type NotificationsListResponse struct {
	Items []NotificationResponse `json:"items"`
}

type NotificationsListEnvelope struct {
	Success bool                      `json:"success" example:"true"`
	Data    NotificationsListResponse `json:"data"`
	Meta    ResponseMetaDTO           `json:"meta"`
}

type ReadAllResponse struct {
	Updated int64 `json:"updated" example:"3"`
}

type ReadAllEnvelope struct {
	Success bool            `json:"success" example:"true"`
	Data    ReadAllResponse `json:"data"`
	Meta    ResponseMetaDTO `json:"meta"`
}

type MessageData struct {
	Message string `json:"message" example:"notification marked as read"`
}

type MessageEnvelope struct {
	Success bool            `json:"success" example:"true"`
	Data    MessageData     `json:"data"`
	Meta    ResponseMetaDTO `json:"meta"`
}

type ResponseMetaDTO struct {
	RequestID string `json:"request_id" example:"e7cc5d4a-3d85-4938-bff3-d2a6cf8ca2bb"`
}

func ToListResponse(items []Notification) NotificationsListResponse {
	result := NotificationsListResponse{
		Items: make([]NotificationResponse, 0, len(items)),
	}
	for _, item := range items {
		result.Items = append(result.Items, toNotificationResponse(item))
	}
	return result
}

func toNotificationResponse(item Notification) NotificationResponse {
	return NotificationResponse{
		ID:           item.ID.String(),
		Type:         string(item.Type),
		Title:        item.Title,
		Body:         item.Body,
		IsRead:       item.IsRead,
		MetadataJSON: parseMetadata(item.MetadataJSON),
		CreatedAt:    item.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:    item.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func parseMetadata(raw json.RawMessage) map[string]any {
	if len(raw) == 0 {
		return nil
	}
	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil
	}
	return result
}

func NewMarkReadCommand(userID uuid.UUID, notificationID string) (MarkReadCommand, error) {
	id, err := uuid.Parse(notificationID)
	if err != nil {
		return MarkReadCommand{}, err
	}
	return MarkReadCommand{
		UserID:         userID,
		NotificationID: id,
	}, nil
}
