package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type RepositoryPort interface {
	CreateAccount(ctx context.Context, account *Account) error
	GetAccountByEmail(ctx context.Context, email string) (*Account, error)
	GetAccountByID(ctx context.Context, userID uuid.UUID) (*Account, error)
	CreateRefreshSession(ctx context.Context, session *RefreshSession) error
	GetRefreshSessionByTokenHash(ctx context.Context, tokenHash string) (*RefreshSession, error)
	RevokeRefreshSessionByID(ctx context.Context, sessionID uuid.UUID, revokedAt time.Time) error
	RevokeRefreshSessionByTokenHash(ctx context.Context, tokenHash string, revokedAt time.Time) error
	RevokeAllUserSessions(ctx context.Context, userID uuid.UUID, revokedAt time.Time) error
	RunInTransaction(ctx context.Context, fn func(repo RepositoryPort) error) error
}

type TokenManager interface {
	GenerateAccessToken(userID uuid.UUID) (string, error)
	GenerateRefreshToken(userID, sessionID uuid.UUID) (string, error)
	Parse(token string) (*Claims, error)
}

type ServicePort interface {
	Register(ctx context.Context, command RegisterCommand, meta SessionMeta) (*AuthResult, error)
	Login(ctx context.Context, command LoginCommand, meta SessionMeta) (*AuthResult, error)
	Refresh(ctx context.Context, command RefreshCommand, meta SessionMeta) (*AuthResult, error)
	Logout(ctx context.Context, command LogoutCommand) error
	LogoutAll(ctx context.Context, command LogoutAllCommand) error
}
