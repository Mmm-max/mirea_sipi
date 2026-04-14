package notifications

import (
	"context"
	"encoding/json"
	"errors"

	"sipi/internal/platform/db"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateMany(ctx context.Context, items []Notification) error {
	if len(items) == 0 {
		return nil
	}
	models := make([]NotificationModel, 0, len(items))
	for i := range items {
		models = append(models, *notificationToModel(&items[i]))
	}
	return r.db.WithContext(ctx).Create(&models).Error
}

func (r *Repository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]Notification, error) {
	var models []NotificationModel
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("is_read ASC, created_at DESC").
		Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]Notification, 0, len(models))
	for i := range models {
		result = append(result, mapNotificationModel(&models[i]))
	}
	return result, nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Notification, error) {
	var model NotificationModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	item := mapNotificationModel(&model)
	return &item, nil
}

func (r *Repository) MarkRead(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&NotificationModel{}).
		Where("id = ?", id).
		Update("is_read", true)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) MarkAllReadByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	result := r.db.WithContext(ctx).
		Model(&NotificationModel{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true)
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

func notificationToModel(item *Notification) *NotificationModel {
	if item == nil {
		return nil
	}
	return &NotificationModel{
		BaseModel: db.BaseModel{
			ID:        item.ID,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		},
		UserID:       item.UserID,
		Type:         string(item.Type),
		Title:        item.Title,
		Body:         item.Body,
		IsRead:       item.IsRead,
		MetadataJSON: cloneRawMessage(item.MetadataJSON),
	}
}

func mapNotificationModel(model *NotificationModel) Notification {
	return Notification{
		ID:           model.ID,
		UserID:       model.UserID,
		Type:         NotificationType(model.Type),
		Title:        model.Title,
		Body:         model.Body,
		IsRead:       model.IsRead,
		MetadataJSON: cloneRawMessage(model.MetadataJSON),
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}
}

func cloneRawMessage(value json.RawMessage) json.RawMessage {
	if len(value) == 0 {
		return nil
	}
	cloned := make([]byte, len(value))
	copy(cloned, value)
	return cloned
}
