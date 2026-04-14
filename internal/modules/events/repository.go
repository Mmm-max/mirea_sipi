package events

import (
	"context"
	"errors"
	"time"

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

func (r *Repository) Create(ctx context.Context, event *Event) error {
	model := eventToModel(event)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}

	*event = mapEventModel(model)
	return nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Event, error) {
	var model EventModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	event := mapEventModel(&model)
	return &event, nil
}

func (r *Repository) GetImportedByExternalUID(ctx context.Context, ownerUserID uuid.UUID, externalUID string) (*Event, error) {
	var model EventModel
	if err := r.db.WithContext(ctx).
		Where("owner_user_id = ? AND source_type = ? AND external_uid = ?", ownerUserID, string(SourceTypeImported), externalUID).
		First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	event := mapEventModel(&model)
	return &event, nil
}

func (r *Repository) ListByOwner(ctx context.Context, ownerUserID uuid.UUID, dateFrom, dateTo *time.Time) ([]Event, error) {
	query := r.db.WithContext(ctx).
		Model(&EventModel{}).
		Where("owner_user_id = ?", ownerUserID).
		Order("start_at ASC")

	if dateFrom != nil {
		query = query.Where("end_at >= ?", *dateFrom)
	}

	if dateTo != nil {
		query = query.Where("start_at <= ?", *dateTo)
	}

	var models []EventModel
	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	events := make([]Event, 0, len(models))
	for i := range models {
		events = append(events, mapEventModel(&models[i]))
	}

	return events, nil
}

func (r *Repository) Update(ctx context.Context, event *Event) error {
	model := eventToModel(event)
	if err := r.db.WithContext(ctx).
		Model(&EventModel{}).
		Where("id = ?", event.ID).
		Updates(map[string]any{
			"source_type":      model.SourceType,
			"external_uid":     model.ExternalUID,
			"title":            model.Title,
			"description":      model.Description,
			"start_at":         model.StartAt,
			"end_at":           model.EndAt,
			"priority":         model.Priority,
			"is_reschedulable": model.IsReschedulable,
			"visibility_hint":  model.VisibilityHint,
		}).Error; err != nil {
		return err
	}

	fresh, err := r.GetByID(ctx, event.ID)
	if err != nil {
		return err
	}

	*event = *fresh
	return nil
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&EventModel{}, "id = ?", id).Error
}

func (r *Repository) HasOverlap(ctx context.Context, ownerUserID uuid.UUID, startAt, endAt time.Time, excludeID *uuid.UUID) (bool, error) {
	query := r.db.WithContext(ctx).
		Model(&EventModel{}).
		Where("owner_user_id = ? AND start_at < ? AND end_at > ?", ownerUserID, endAt, startAt)

	if excludeID != nil {
		query = query.Where("id <> ?", *excludeID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func eventToModel(event *Event) *EventModel {
	if event == nil {
		return nil
	}

	model := &EventModel{
		BaseModel: db.BaseModel{
			ID:        event.ID,
			CreatedAt: event.CreatedAt,
			UpdatedAt: event.UpdatedAt,
		},
		OwnerUserID:     event.OwnerUserID,
		SourceType:      string(event.SourceType),
		ExternalUID:     event.ExternalUID,
		Title:           event.Title,
		Description:     event.Description,
		StartAt:         event.StartAt,
		EndAt:           event.EndAt,
		Priority:        string(event.Priority),
		IsReschedulable: event.IsReschedulable,
		VisibilityHint:  event.VisibilityHint,
	}

	if event.DeletedAt != nil {
		model.DeletedAt = gorm.DeletedAt{Time: *event.DeletedAt, Valid: true}
	}

	return model
}

func mapEventModel(model *EventModel) Event {
	var deletedAt *time.Time
	if model.DeletedAt.Valid {
		deletedAt = &model.DeletedAt.Time
	}

	return Event{
		ID:              model.ID,
		OwnerUserID:     model.OwnerUserID,
		SourceType:      SourceType(model.SourceType),
		ExternalUID:     model.ExternalUID,
		Title:           model.Title,
		Description:     model.Description,
		StartAt:         model.StartAt,
		EndAt:           model.EndAt,
		Priority:        Priority(model.Priority),
		IsReschedulable: model.IsReschedulable,
		VisibilityHint:  model.VisibilityHint,
		CreatedAt:       model.CreatedAt,
		UpdatedAt:       model.UpdatedAt,
		DeletedAt:       deletedAt,
	}
}
