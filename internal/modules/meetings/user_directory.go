package meetings

import (
	"context"
	"errors"

	"sipi/internal/modules/users"

	"github.com/google/uuid"
)

type userLookupRepository interface {
	GetUserByID(ctx context.Context, userID uuid.UUID) (*users.User, error)
}

type UserDirectory struct {
	repository userLookupRepository
}

func NewUserDirectory(repository userLookupRepository) *UserDirectory {
	return &UserDirectory{repository: repository}
}

func (d *UserDirectory) Exists(ctx context.Context, userID uuid.UUID) (bool, error) {
	_, err := d.repository.GetUserByID(ctx, userID)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, users.ErrNotFound) {
		return false, nil
	}
	return false, err
}
