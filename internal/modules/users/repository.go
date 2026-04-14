package users

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

func (r *Repository) GetUserByID(ctx context.Context, userID uuid.UUID) (*User, error) {
	var model UserModel
	if err := r.db.WithContext(ctx).Where("id = ?", userID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	user := mapUserModel(&model)
	return &user, nil
}

func (r *Repository) UpdateUserProfile(ctx context.Context, user *User) error {
	model := userToModel(user)
	if err := r.db.WithContext(ctx).
		Model(&UserModel{}).
		Where("id = ?", user.ID).
		Updates(map[string]any{
			"full_name": user.FullName,
			"timezone":  user.Timezone,
		}).Error; err != nil {
		return err
	}

	if err := r.db.WithContext(ctx).Where("id = ?", user.ID).First(model).Error; err != nil {
		return err
	}

	*user = mapUserModel(model)
	return nil
}

func (r *Repository) ReplaceWorkingHours(ctx context.Context, userID uuid.UUID, items []WorkingHours) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&WorkingHoursModel{}).Error; err != nil {
			return err
		}

		if len(items) == 0 {
			return nil
		}

		models := make([]WorkingHoursModel, 0, len(items))
		for _, item := range items {
			models = append(models, *workingHoursToModel(&item))
		}

		return tx.Create(&models).Error
	})
}

func (r *Repository) ListWorkingHours(ctx context.Context, userID uuid.UUID) ([]WorkingHours, error) {
	var models []WorkingHoursModel
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("weekday ASC").
		Find(&models).Error; err != nil {
		return nil, err
	}

	items := make([]WorkingHours, 0, len(models))
	for i := range models {
		items = append(items, mapWorkingHoursModel(&models[i]))
	}

	return items, nil
}

func (r *Repository) CreateUnavailabilityPeriod(ctx context.Context, period *UnavailabilityPeriod) error {
	model := unavailabilityToModel(period)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}

	*period = mapUnavailabilityModel(model)
	return nil
}

func (r *Repository) ListUnavailabilityPeriods(ctx context.Context, userID uuid.UUID) ([]UnavailabilityPeriod, error) {
	var models []UnavailabilityPeriodModel
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("start_at ASC").
		Find(&models).Error; err != nil {
		return nil, err
	}

	items := make([]UnavailabilityPeriod, 0, len(models))
	for i := range models {
		items = append(items, mapUnavailabilityModel(&models[i]))
	}

	return items, nil
}

func (r *Repository) GetUnavailabilityPeriodByID(ctx context.Context, id uuid.UUID) (*UnavailabilityPeriod, error) {
	var model UnavailabilityPeriodModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	item := mapUnavailabilityModel(&model)
	return &item, nil
}

func (r *Repository) DeleteUnavailabilityPeriod(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&UnavailabilityPeriodModel{}).Error
}

func (r *Repository) HasUnavailabilityOverlap(ctx context.Context, userID uuid.UUID, startAt, endAt time.Time) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&UnavailabilityPeriodModel{}).
		Where("user_id = ? AND start_at < ? AND end_at > ?", userID, endAt, startAt).
		Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func userToModel(user *User) *UserModel {
	if user == nil {
		return nil
	}

	return &UserModel{
		BaseModel: db.BaseModel{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
		Email:        user.Email,
		FullName:     user.FullName,
		Timezone:     user.Timezone,
		PasswordHash: user.PasswordHash,
	}
}

func mapUserModel(model *UserModel) User {
	return User{
		ID:           model.ID,
		Email:        model.Email,
		FullName:     model.FullName,
		Timezone:     model.Timezone,
		PasswordHash: model.PasswordHash,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}
}

func workingHoursToModel(item *WorkingHours) *WorkingHoursModel {
	if item == nil {
		return nil
	}

	return &WorkingHoursModel{
		BaseModel: db.BaseModel{
			ID:        item.ID,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		},
		UserID:       item.UserID,
		Weekday:      item.Weekday,
		StartTime:    item.StartTime,
		EndTime:      item.EndTime,
		IsWorkingDay: item.IsWorkingDay,
	}
}

func mapWorkingHoursModel(model *WorkingHoursModel) WorkingHours {
	return WorkingHours{
		ID:           model.ID,
		UserID:       model.UserID,
		Weekday:      model.Weekday,
		StartTime:    model.StartTime,
		EndTime:      model.EndTime,
		IsWorkingDay: model.IsWorkingDay,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}
}

func unavailabilityToModel(item *UnavailabilityPeriod) *UnavailabilityPeriodModel {
	if item == nil {
		return nil
	}

	return &UnavailabilityPeriodModel{
		BaseModel: db.BaseModel{
			ID:        item.ID,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		},
		UserID:  item.UserID,
		Type:    string(item.Type),
		Title:   item.Title,
		StartAt: item.StartAt,
		EndAt:   item.EndAt,
		Comment: item.Comment,
	}
}

func mapUnavailabilityModel(model *UnavailabilityPeriodModel) UnavailabilityPeriod {
	return UnavailabilityPeriod{
		ID:        model.ID,
		UserID:    model.UserID,
		Type:      UnavailabilityType(model.Type),
		Title:     model.Title,
		StartAt:   model.StartAt,
		EndAt:     model.EndAt,
		Comment:   model.Comment,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}
