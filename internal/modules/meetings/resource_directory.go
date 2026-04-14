package meetings

import (
	"context"
	"errors"
	"time"

	"sipi/internal/modules/resources"

	"github.com/google/uuid"
)

type resourceLookupRepository interface {
	GetResourceByID(ctx context.Context, id uuid.UUID) (*resources.Resource, error)
	HasBookingOverlap(ctx context.Context, resourceID uuid.UUID, startAt, endAt time.Time, excludeID *uuid.UUID) (bool, error)
}

type ResourceDirectory struct {
	repository resourceLookupRepository
}

func NewResourceDirectory(repository resourceLookupRepository) *ResourceDirectory {
	return &ResourceDirectory{repository: repository}
}

func (d *ResourceDirectory) Exists(ctx context.Context, resourceID uuid.UUID) (bool, error) {
	_, err := d.repository.GetResourceByID(ctx, resourceID)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, resources.ErrNotFound) {
		return false, nil
	}
	return false, err
}

func (d *ResourceDirectory) IsAvailable(ctx context.Context, resourceID uuid.UUID, startAt, endAt time.Time) (bool, error) {
	hasOverlap, err := d.repository.HasBookingOverlap(ctx, resourceID, startAt, endAt, nil)
	if err != nil {
		return false, err
	}
	return !hasOverlap, nil
}
