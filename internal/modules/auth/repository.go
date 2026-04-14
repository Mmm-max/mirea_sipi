package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	platformdb "sipi/internal/platform/db"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateAccount(ctx context.Context, account *Account) error {
	model := accountToModel(account)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		if isUniqueViolation(err) {
			return ErrAlreadyExists
		}
		return err
	}

	*account = mapAccountModel(model)
	return nil
}

func (r *Repository) GetAccountByEmail(ctx context.Context, email string) (*Account, error) {
	var model AccountModel
	err := r.db.WithContext(ctx).
		Where("email = ?", strings.TrimSpace(strings.ToLower(email))).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	account := mapAccountModel(&model)
	return &account, nil
}

func (r *Repository) GetAccountByID(ctx context.Context, userID uuid.UUID) (*Account, error) {
	var model AccountModel
	err := r.db.WithContext(ctx).
		Where("id = ?", userID).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	account := mapAccountModel(&model)
	return &account, nil
}

func (r *Repository) CreateRefreshSession(ctx context.Context, session *RefreshSession) error {
	model := refreshSessionToModel(session)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}

	*session = mapRefreshSessionModel(model)
	return nil
}

func (r *Repository) GetRefreshSessionByTokenHash(ctx context.Context, tokenHash string) (*RefreshSession, error) {
	var model RefreshSessionModel
	err := r.db.WithContext(ctx).
		Where("token_hash = ?", tokenHash).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	session := mapRefreshSessionModel(&model)
	return &session, nil
}

func (r *Repository) RevokeRefreshSessionByID(ctx context.Context, sessionID uuid.UUID, revokedAt time.Time) error {
	return r.db.WithContext(ctx).
		Model(&RefreshSessionModel{}).
		Where("id = ? AND revoked_at IS NULL", sessionID).
		Update("revoked_at", revokedAt).
		Error
}

func (r *Repository) RevokeRefreshSessionByTokenHash(ctx context.Context, tokenHash string, revokedAt time.Time) error {
	return r.db.WithContext(ctx).
		Model(&RefreshSessionModel{}).
		Where("token_hash = ? AND revoked_at IS NULL", tokenHash).
		Update("revoked_at", revokedAt).
		Error
}

func (r *Repository) RevokeAllUserSessions(ctx context.Context, userID uuid.UUID, revokedAt time.Time) error {
	return r.db.WithContext(ctx).
		Model(&RefreshSessionModel{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", revokedAt).
		Error
}

func (r *Repository) RunInTransaction(ctx context.Context, fn func(repo RepositoryPort) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&Repository{db: tx})
	})
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(strings.ToLower(err.Error()), "duplicate key value")
}

func accountToModel(account *Account) *AccountModel {
	if account == nil {
		return nil
	}

	return &AccountModel{
		BaseModel: platformdb.BaseModel{
			ID:        account.ID,
			CreatedAt: account.CreatedAt,
			UpdatedAt: account.UpdatedAt,
		},
		Email:        account.Email,
		FullName:     account.FullName,
		Timezone:     account.Timezone,
		PasswordHash: account.PasswordHash,
	}
}

func mapAccountModel(model *AccountModel) Account {
	return Account{
		ID:           model.ID,
		Email:        model.Email,
		FullName:     model.FullName,
		Timezone:     model.Timezone,
		PasswordHash: model.PasswordHash,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}
}

func refreshSessionToModel(session *RefreshSession) *RefreshSessionModel {
	if session == nil {
		return nil
	}

	return &RefreshSessionModel{
		BaseModel: platformdb.BaseModel{
			ID:        session.ID,
			CreatedAt: session.CreatedAt,
			UpdatedAt: session.UpdatedAt,
		},
		UserID:    session.UserID,
		TokenHash: session.TokenHash,
		ExpiresAt: session.ExpiresAt,
		RevokedAt: session.RevokedAt,
		UserAgent: session.UserAgent,
		IPAddress: session.IPAddress,
	}
}

func mapRefreshSessionModel(model *RefreshSessionModel) RefreshSession {
	return RefreshSession{
		ID:        model.ID,
		UserID:    model.UserID,
		TokenHash: model.TokenHash,
		ExpiresAt: model.ExpiresAt,
		RevokedAt: model.RevokedAt,
		UserAgent: model.UserAgent,
		IPAddress: model.IPAddress,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}
