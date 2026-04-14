package calendarimport

import (
	"context"
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

func (r *Repository) Create(ctx context.Context, job *ImportJob) error {
	model := toImportJobModel(job)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}

	*job = mapImportJobModel(model)
	return nil
}

func (r *Repository) Update(ctx context.Context, job *ImportJob) error {
	model := toImportJobModel(job)
	if err := r.db.WithContext(ctx).
		Model(&ImportJobModel{}).
		Where("id = ?", job.ID).
		Updates(map[string]any{
			"status":         model.Status,
			"imported_count": model.ImportedCount,
			"updated_count":  model.UpdatedCount,
			"skipped_count":  model.SkippedCount,
			"error_message":  model.ErrorMessage,
		}).Error; err != nil {
		return err
	}

	fresh, err := r.GetByID(ctx, job.ID)
	if err != nil {
		return err
	}

	*job = *fresh
	return nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*ImportJob, error) {
	var model ImportJobModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	job := mapImportJobModel(&model)
	return &job, nil
}

func (r *Repository) ListByUser(ctx context.Context, userID uuid.UUID) ([]ImportJob, error) {
	var models []ImportJobModel
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&models).Error; err != nil {
		return nil, err
	}

	result := make([]ImportJob, 0, len(models))
	for i := range models {
		result = append(result, mapImportJobModel(&models[i]))
	}

	return result, nil
}

func toImportJobModel(job *ImportJob) *ImportJobModel {
	if job == nil {
		return nil
	}

	return &ImportJobModel{
		BaseModel: db.BaseModel{
			ID:        job.ID,
			CreatedAt: job.CreatedAt,
			UpdatedAt: job.UpdatedAt,
		},
		UserID:           job.UserID,
		OriginalFilename: job.OriginalFilename,
		Status:           string(job.Status),
		ImportedCount:    job.ImportedCount,
		UpdatedCount:     job.UpdatedCount,
		SkippedCount:     job.SkippedCount,
		ErrorMessage:     job.ErrorMessage,
	}
}

func mapImportJobModel(model *ImportJobModel) ImportJob {
	return ImportJob{
		ID:               model.ID,
		UserID:           model.UserID,
		OriginalFilename: model.OriginalFilename,
		Status:           ImportJobStatus(model.Status),
		ImportedCount:    model.ImportedCount,
		UpdatedCount:     model.UpdatedCount,
		SkippedCount:     model.SkippedCount,
		ErrorMessage:     model.ErrorMessage,
		CreatedAt:        model.CreatedAt,
		UpdatedAt:        model.UpdatedAt,
	}
}
