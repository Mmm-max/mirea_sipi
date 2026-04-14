package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repository   RepositoryPort
	tokenManager TokenManager
}

func NewService(repository RepositoryPort, tokenManager TokenManager) *Service {
	return &Service{
		repository:   repository,
		tokenManager: tokenManager,
	}
}

func (s *Service) Register(ctx context.Context, command RegisterCommand, meta SessionMeta) (*AuthResult, error) {
	email := normalizeEmail(command.Email)

	existingUser, err := s.repository.GetAccountByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, newServiceError(ErrCodeEmailAlreadyExists, "email already exists", nil)
	}

	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, newServiceError(ErrCodeInternal, "failed to create user", err)
	}

	passwordHash, err := HashPassword(command.Password)
	if err != nil {
		return nil, newServiceError(ErrCodeInternal, "failed to hash password", err)
	}

	user := &Account{
		Email:        email,
		Timezone:     "UTC",
		PasswordHash: passwordHash,
	}

	if err := s.repository.CreateAccount(ctx, user); err != nil {
		if errors.Is(err, ErrAlreadyExists) {
			return nil, newServiceError(ErrCodeEmailAlreadyExists, "email already exists", err)
		}

		return nil, newServiceError(ErrCodeInternal, "failed to create user", err)
	}

	return s.issueTokenPair(ctx, user, meta)
}

func (s *Service) Login(ctx context.Context, command LoginCommand, meta SessionMeta) (*AuthResult, error) {
	user, err := s.repository.GetAccountByEmail(ctx, normalizeEmail(command.Email))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, newServiceError(ErrCodeInvalidCredentials, "invalid email or password", err)
		}

		return nil, newServiceError(ErrCodeInternal, "failed to login", err)
	}

	if err := ComparePassword(user.PasswordHash, command.Password); err != nil {
		return nil, newServiceError(ErrCodeInvalidCredentials, "invalid email or password", err)
	}

	return s.issueTokenPair(ctx, user, meta)
}

func (s *Service) Refresh(ctx context.Context, command RefreshCommand, meta SessionMeta) (*AuthResult, error) {
	claims, err := s.tokenManager.Parse(command.RefreshToken)
	if err != nil || claims.Type != "refresh" {
		return nil, newServiceError(ErrCodeInvalidRefreshToken, "invalid refresh token", err)
	}

	tokenHash := HashRefreshToken(command.RefreshToken)
	session, err := s.repository.GetRefreshSessionByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, newServiceError(ErrCodeInvalidRefreshToken, "invalid refresh token", err)
		}

		return nil, newServiceError(ErrCodeInternal, "failed to refresh token", err)
	}

	if session.UserID != claims.UserID || session.ID.String() != claims.TokenID {
		return nil, newServiceError(ErrCodeInvalidRefreshToken, "invalid refresh token", nil)
	}

	now := time.Now()
	if session.RevokedAt != nil {
		if revokeErr := s.repository.RevokeAllUserSessions(ctx, session.UserID, now); revokeErr != nil {
			return nil, newServiceError(ErrCodeInternal, "failed to process refresh token reuse", revokeErr)
		}

		return nil, newServiceError(ErrCodeRefreshTokenReused, "refresh token has already been used", nil)
	}

	if session.ExpiresAt.Before(now) {
		_ = s.repository.RevokeRefreshSessionByID(ctx, session.ID, now)
		return nil, newServiceError(ErrCodeRefreshTokenExpired, "refresh token expired", nil)
	}

	var result *AuthResult
	err = s.repository.RunInTransaction(ctx, func(repo RepositoryPort) error {
		if err := repo.RevokeRefreshSessionByID(ctx, session.ID, now); err != nil {
			return err
		}

		user, err := repo.GetAccountByID(ctx, session.UserID)
		if err != nil {
			return err
		}

		issued, err := s.issueTokenPairWithRepository(ctx, repo, user, meta)
		if err != nil {
			return err
		}

		result = issued
		return nil
	})
	if err != nil {
		if serviceErr, ok := err.(*ServiceError); ok {
			return nil, serviceErr
		}

		return nil, newServiceError(ErrCodeInternal, "failed to refresh token", err)
	}

	return result, nil
}

func (s *Service) Logout(ctx context.Context, command LogoutCommand) error {
	claims, err := s.tokenManager.Parse(command.RefreshToken)
	if err != nil || claims.Type != "refresh" {
		return newServiceError(ErrCodeInvalidRefreshToken, "invalid refresh token", err)
	}

	tokenHash := HashRefreshToken(command.RefreshToken)
	session, err := s.repository.GetRefreshSessionByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return newServiceError(ErrCodeInvalidRefreshToken, "invalid refresh token", err)
		}

		return newServiceError(ErrCodeInternal, "failed to logout", err)
	}

	if session.UserID != command.UserID || claims.UserID != command.UserID {
		return newServiceError(ErrCodeForbidden, "refresh token does not belong to current user", nil)
	}

	if err := s.repository.RevokeRefreshSessionByTokenHash(ctx, tokenHash, time.Now()); err != nil {
		return newServiceError(ErrCodeInternal, "failed to logout", err)
	}

	return nil
}

func (s *Service) LogoutAll(ctx context.Context, command LogoutAllCommand) error {
	if err := s.repository.RevokeAllUserSessions(ctx, command.UserID, time.Now()); err != nil {
		return newServiceError(ErrCodeInternal, "failed to logout from all sessions", err)
	}

	return nil
}

func (s *Service) issueTokenPair(ctx context.Context, user *Account, meta SessionMeta) (*AuthResult, error) {
	return s.issueTokenPairWithRepository(ctx, s.repository, user, meta)
}

func (s *Service) issueTokenPairWithRepository(ctx context.Context, repo RepositoryPort, user *Account, meta SessionMeta) (*AuthResult, error) {
	sessionID := uuid.New()

	accessToken, err := s.tokenManager.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, newServiceError(ErrCodeInternal, "failed to generate access token", err)
	}

	refreshToken, err := s.tokenManager.GenerateRefreshToken(user.ID, sessionID)
	if err != nil {
		return nil, newServiceError(ErrCodeInternal, "failed to generate refresh token", err)
	}

	refreshClaims, err := s.tokenManager.Parse(refreshToken)
	if err != nil {
		return nil, newServiceError(ErrCodeInternal, "failed to parse refresh token", err)
	}

	session := &RefreshSession{
		ID:        sessionID,
		UserID:    user.ID,
		TokenHash: HashRefreshToken(refreshToken),
		ExpiresAt: refreshClaims.ExpiresAt,
		UserAgent: truncate(meta.UserAgent, 512),
		IPAddress: truncate(meta.IPAddress, 64),
	}

	if err := repo.CreateRefreshSession(ctx, session); err != nil {
		return nil, newServiceError(ErrCodeInternal, "failed to create refresh session", err)
	}

	return &AuthResult{
		User: AuthUser{
			ID:    user.ID,
			Email: user.Email,
		},
		Tokens: TokenPair{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			TokenType:    "Bearer",
		},
	}, nil
}

func normalizeEmail(email string) string {
	return strings.TrimSpace(strings.ToLower(email))
}

func truncate(value string, maxLength int) string {
	if len(value) <= maxLength {
		return value
	}

	return value[:maxLength]
}
