package notifications

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
)

type Service struct {
	repository RepositoryPort
}

func NewService(repository RepositoryPort) *Service {
	return &Service{repository: repository}
}

func (s *Service) CreateMany(ctx context.Context, items []CreateNotificationInput) error {
	if len(items) == 0 {
		return nil
	}
	notifications := make([]Notification, 0, len(items))
	for _, item := range items {
		notification := Notification{
			ID:           uuid.New(),
			UserID:       item.UserID,
			Type:         item.Type,
			Title:        strings.TrimSpace(item.Title),
			Body:         strings.TrimSpace(item.Body),
			IsRead:       false,
			MetadataJSON: cloneRawMessage(item.MetadataJSON),
		}
		if err := validateNotification(&notification); err != nil {
			return err
		}
		notifications = append(notifications, notification)
	}
	if err := s.repository.CreateMany(ctx, notifications); err != nil {
		return newServiceError(ErrCodeInternal, "failed to create notifications", err)
	}
	return nil
}

func (s *Service) ListNotifications(ctx context.Context, query ListNotificationsQuery) ([]Notification, error) {
	items, err := s.repository.ListByUserID(ctx, query.UserID)
	if err != nil {
		return nil, newServiceError(ErrCodeInternal, "failed to list notifications", err)
	}
	return items, nil
}

func (s *Service) MarkRead(ctx context.Context, command MarkReadCommand) error {
	item, err := s.repository.GetByID(ctx, command.NotificationID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return newServiceError(ErrCodeNotFound, "notification not found", err)
		}
		return newServiceError(ErrCodeInternal, "failed to load notification", err)
	}
	if item.UserID != command.UserID {
		return newServiceError(ErrCodeNotFound, "notification not found", nil)
	}
	if item.IsRead {
		return nil
	}
	if err := s.repository.MarkRead(ctx, item.ID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return newServiceError(ErrCodeNotFound, "notification not found", err)
		}
		return newServiceError(ErrCodeInternal, "failed to mark notification as read", err)
	}
	return nil
}

func (s *Service) MarkAllRead(ctx context.Context, command MarkAllReadCommand) (int64, error) {
	updated, err := s.repository.MarkAllReadByUserID(ctx, command.UserID)
	if err != nil {
		return 0, newServiceError(ErrCodeInternal, "failed to mark notifications as read", err)
	}
	return updated, nil
}

func validateNotification(item *Notification) error {
	if item == nil {
		return newServiceError(ErrCodeValidation, "notification is required", nil)
	}
	if item.UserID == uuid.Nil {
		return newServiceError(ErrCodeValidation, "user_id is required", nil)
	}
	if strings.TrimSpace(string(item.Type)) == "" {
		return newServiceError(ErrCodeValidation, "type is required", nil)
	}
	if strings.TrimSpace(item.Title) == "" {
		return newServiceError(ErrCodeValidation, "title is required", nil)
	}
	if strings.TrimSpace(item.Body) == "" {
		return newServiceError(ErrCodeValidation, "body is required", nil)
	}
	return nil
}
