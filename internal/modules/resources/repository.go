package resources

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

func (r *Repository) CreateResource(ctx context.Context, resource *Resource) error {
	model := resourceToModel(resource)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}

	*resource = mapResourceModel(model)
	return nil
}

func (r *Repository) GetResourceByID(ctx context.Context, id uuid.UUID) (*Resource, error) {
	var model ResourceModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	resource := mapResourceModel(&model)
	return &resource, nil
}

func (r *Repository) ListResources(ctx context.Context) ([]Resource, error) {
	var models []ResourceModel
	if err := r.db.WithContext(ctx).Order("name ASC").Find(&models).Error; err != nil {
		return nil, err
	}

	result := make([]Resource, 0, len(models))
	for i := range models {
		result = append(result, mapResourceModel(&models[i]))
	}

	return result, nil
}

func (r *Repository) UpdateResource(ctx context.Context, resource *Resource) error {
	model := resourceToModel(resource)
	if err := r.db.WithContext(ctx).
		Model(&ResourceModel{}).
		Where("id = ?", resource.ID).
		Updates(map[string]any{
			"name":          model.Name,
			"type":          model.Type,
			"description":   model.Description,
			"capacity":      model.Capacity,
			"location":      model.Location,
			"owner_user_id": model.OwnerUserID,
		}).Error; err != nil {
		return err
	}

	fresh, err := r.GetResourceByID(ctx, resource.ID)
	if err != nil {
		return err
	}

	*resource = *fresh
	return nil
}

func (r *Repository) DeleteResource(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&ResourceModel{}, "id = ?", id).Error
}

func (r *Repository) CreateBooking(ctx context.Context, booking *ResourceBooking) error {
	model := bookingToModel(booking)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}

	*booking = mapBookingModel(model)
	return nil
}

func (r *Repository) GetBookingByID(ctx context.Context, id uuid.UUID) (*ResourceBooking, error) {
	var model ResourceBookingModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	booking := mapBookingModel(&model)
	return &booking, nil
}

func (r *Repository) DeleteBooking(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&ResourceBookingModel{}, "id = ?", id).Error
}

func (r *Repository) ListBookingsByResource(ctx context.Context, resourceID uuid.UUID, dateFrom, dateTo time.Time) ([]ResourceBooking, error) {
	var models []ResourceBookingModel
	if err := r.db.WithContext(ctx).
		Where("resource_id = ? AND end_at >= ? AND start_at <= ?", resourceID, dateFrom, dateTo).
		Order("start_at ASC").
		Find(&models).Error; err != nil {
		return nil, err
	}

	result := make([]ResourceBooking, 0, len(models))
	for i := range models {
		result = append(result, mapBookingModel(&models[i]))
	}

	return result, nil
}

func (r *Repository) HasBookingOverlap(ctx context.Context, resourceID uuid.UUID, startAt, endAt time.Time, excludeID *uuid.UUID) (bool, error) {
	query := r.db.WithContext(ctx).
		Model(&ResourceBookingModel{}).
		Where("resource_id = ? AND start_at < ? AND end_at > ?", resourceID, endAt, startAt)

	if excludeID != nil {
		query = query.Where("id <> ?", *excludeID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func resourceToModel(resource *Resource) *ResourceModel {
	if resource == nil {
		return nil
	}

	model := &ResourceModel{
		BaseModel: db.BaseModel{
			ID:        resource.ID,
			CreatedAt: resource.CreatedAt,
			UpdatedAt: resource.UpdatedAt,
		},
		Name:        resource.Name,
		Type:        string(resource.Type),
		Description: resource.Description,
		Capacity:    resource.Capacity,
		Location:    resource.Location,
		OwnerUserID: resource.OwnerUserID,
	}
	if resource.DeletedAt != nil {
		model.DeletedAt = gorm.DeletedAt{Time: *resource.DeletedAt, Valid: true}
	}

	return model
}

func mapResourceModel(model *ResourceModel) Resource {
	var deletedAt *time.Time
	if model.DeletedAt.Valid {
		deletedAt = &model.DeletedAt.Time
	}

	return Resource{
		ID:          model.ID,
		Name:        model.Name,
		Type:        ResourceType(model.Type),
		Description: model.Description,
		Capacity:    model.Capacity,
		Location:    model.Location,
		OwnerUserID: model.OwnerUserID,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
		DeletedAt:   deletedAt,
	}
}

func bookingToModel(booking *ResourceBooking) *ResourceBookingModel {
	if booking == nil {
		return nil
	}

	model := &ResourceBookingModel{
		BaseModel: db.BaseModel{
			ID:        booking.ID,
			CreatedAt: booking.CreatedAt,
			UpdatedAt: booking.UpdatedAt,
		},
		ResourceID:     booking.ResourceID,
		EventID:        booking.EventID,
		BookedByUserID: booking.BookedByUserID,
		StartAt:        booking.StartAt,
		EndAt:          booking.EndAt,
		Title:          booking.Title,
	}
	if booking.DeletedAt != nil {
		model.DeletedAt = gorm.DeletedAt{Time: *booking.DeletedAt, Valid: true}
	}

	return model
}

func mapBookingModel(model *ResourceBookingModel) ResourceBooking {
	var deletedAt *time.Time
	if model.DeletedAt.Valid {
		deletedAt = &model.DeletedAt.Time
	}

	return ResourceBooking{
		ID:             model.ID,
		ResourceID:     model.ResourceID,
		EventID:        model.EventID,
		BookedByUserID: model.BookedByUserID,
		StartAt:        model.StartAt,
		EndAt:          model.EndAt,
		Title:          model.Title,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
		DeletedAt:      deletedAt,
	}
}
