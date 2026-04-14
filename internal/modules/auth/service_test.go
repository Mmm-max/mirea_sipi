package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

type fakeAuthRepository struct {
	accountsByEmail map[string]*Account
	accountsByID    map[uuid.UUID]*Account
	sessionsByHash  map[string]*RefreshSession
}

func newFakeAuthRepository() *fakeAuthRepository {
	return &fakeAuthRepository{
		accountsByEmail: make(map[string]*Account),
		accountsByID:    make(map[uuid.UUID]*Account),
		sessionsByHash:  make(map[string]*RefreshSession),
	}
}

func (r *fakeAuthRepository) CreateAccount(_ context.Context, account *Account) error {
	if _, exists := r.accountsByEmail[account.Email]; exists {
		return ErrAlreadyExists
	}
	if account.ID == uuid.Nil {
		account.ID = uuid.New()
	}
	copyValue := *account
	r.accountsByEmail[account.Email] = &copyValue
	r.accountsByID[account.ID] = &copyValue
	return nil
}

func (r *fakeAuthRepository) GetAccountByEmail(_ context.Context, email string) (*Account, error) {
	account, ok := r.accountsByEmail[email]
	if !ok {
		return nil, ErrNotFound
	}
	copyValue := *account
	return &copyValue, nil
}

func (r *fakeAuthRepository) GetAccountByID(_ context.Context, userID uuid.UUID) (*Account, error) {
	account, ok := r.accountsByID[userID]
	if !ok {
		return nil, ErrNotFound
	}
	copyValue := *account
	return &copyValue, nil
}

func (r *fakeAuthRepository) CreateRefreshSession(_ context.Context, session *RefreshSession) error {
	copyValue := *session
	r.sessionsByHash[session.TokenHash] = &copyValue
	return nil
}

func (r *fakeAuthRepository) GetRefreshSessionByTokenHash(_ context.Context, tokenHash string) (*RefreshSession, error) {
	session, ok := r.sessionsByHash[tokenHash]
	if !ok {
		return nil, ErrNotFound
	}
	copyValue := *session
	return &copyValue, nil
}

func (r *fakeAuthRepository) RevokeRefreshSessionByID(_ context.Context, sessionID uuid.UUID, revokedAt time.Time) error {
	for key, session := range r.sessionsByHash {
		if session.ID == sessionID {
			copyValue := *session
			copyValue.RevokedAt = &revokedAt
			r.sessionsByHash[key] = &copyValue
			return nil
		}
	}
	return ErrNotFound
}

func (r *fakeAuthRepository) RevokeRefreshSessionByTokenHash(_ context.Context, tokenHash string, revokedAt time.Time) error {
	session, ok := r.sessionsByHash[tokenHash]
	if !ok {
		return ErrNotFound
	}
	copyValue := *session
	copyValue.RevokedAt = &revokedAt
	r.sessionsByHash[tokenHash] = &copyValue
	return nil
}

func (r *fakeAuthRepository) RevokeAllUserSessions(_ context.Context, userID uuid.UUID, revokedAt time.Time) error {
	for key, session := range r.sessionsByHash {
		if session.UserID != userID {
			continue
		}
		copyValue := *session
		copyValue.RevokedAt = &revokedAt
		r.sessionsByHash[key] = &copyValue
	}
	return nil
}

func (r *fakeAuthRepository) RunInTransaction(ctx context.Context, fn func(repo RepositoryPort) error) error {
	return fn(r)
}

type fakeTokenManager struct {
	claimsByToken map[string]*Claims
	refreshTokens []string
}

func newFakeTokenManager() *fakeTokenManager {
	return &fakeTokenManager{claimsByToken: make(map[string]*Claims)}
}

func (m *fakeTokenManager) GenerateAccessToken(userID uuid.UUID) (string, error) {
	token := "access-" + userID.String()
	m.claimsByToken[token] = &Claims{
		UserID:    userID,
		Type:      "access",
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
	return token, nil
}

func (m *fakeTokenManager) GenerateRefreshToken(userID, sessionID uuid.UUID) (string, error) {
	token := "refresh-" + sessionID.String()
	m.claimsByToken[token] = &Claims{
		UserID:    userID,
		Type:      "refresh",
		TokenID:   sessionID.String(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	m.refreshTokens = append(m.refreshTokens, token)
	return token, nil
}

func (m *fakeTokenManager) Parse(token string) (*Claims, error) {
	claims, ok := m.claimsByToken[token]
	if !ok {
		return nil, errors.New("token not found")
	}
	copyValue := *claims
	return &copyValue, nil
}

func TestServiceRegister(t *testing.T) {
	t.Parallel()

	repo := newFakeAuthRepository()
	tokens := newFakeTokenManager()
	service := NewService(repo, tokens)

	result, err := service.Register(context.Background(), RegisterCommand{
		Email:    "  USER@Example.com ",
		Password: "StrongPass123!",
	}, SessionMeta{UserAgent: "ua", IPAddress: "127.0.0.1"})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if result.User.Email != "user@example.com" {
		t.Fatalf("expected normalized email, got %q", result.User.Email)
	}
	account, err := repo.GetAccountByEmail(context.Background(), "user@example.com")
	if err != nil {
		t.Fatalf("GetAccountByEmail() error = %v", err)
	}
	if account.PasswordHash == "" || account.PasswordHash == "StrongPass123!" {
		t.Fatalf("expected hashed password to be stored")
	}
	if result.Tokens.RefreshToken == "" || result.Tokens.AccessToken == "" {
		t.Fatalf("expected token pair to be issued")
	}
	if _, err := repo.GetRefreshSessionByTokenHash(context.Background(), HashRefreshToken(result.Tokens.RefreshToken)); err != nil {
		t.Fatalf("expected refresh session to be stored: %v", err)
	}
}

func TestServiceLogin(t *testing.T) {
	t.Parallel()

	repo := newFakeAuthRepository()
	tokens := newFakeTokenManager()
	service := NewService(repo, tokens)

	hash, err := HashPassword("StrongPass123!")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	user := &Account{ID: uuid.New(), Email: "user@example.com", PasswordHash: hash, Timezone: "UTC"}
	if err := repo.CreateAccount(context.Background(), user); err != nil {
		t.Fatalf("CreateAccount() error = %v", err)
	}

	result, err := service.Login(context.Background(), LoginCommand{
		Email:    "USER@example.com",
		Password: "StrongPass123!",
	}, SessionMeta{})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	if result.User.ID != user.ID {
		t.Fatalf("expected user ID %s, got %s", user.ID, result.User.ID)
	}
	if _, err := repo.GetRefreshSessionByTokenHash(context.Background(), HashRefreshToken(result.Tokens.RefreshToken)); err != nil {
		t.Fatalf("expected refresh session to be stored: %v", err)
	}
}

func TestServiceRefresh(t *testing.T) {
	t.Parallel()

	repo := newFakeAuthRepository()
	tokens := newFakeTokenManager()
	service := NewService(repo, tokens)

	hash, err := HashPassword("StrongPass123!")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	user := &Account{ID: uuid.New(), Email: "user@example.com", PasswordHash: hash, Timezone: "UTC"}
	if err := repo.CreateAccount(context.Background(), user); err != nil {
		t.Fatalf("CreateAccount() error = %v", err)
	}

	oldSessionID := uuid.New()
	oldRefreshToken := "refresh-" + oldSessionID.String()
	tokens.claimsByToken[oldRefreshToken] = &Claims{
		UserID:    user.ID,
		Type:      "refresh",
		TokenID:   oldSessionID.String(),
		ExpiresAt: time.Now().Add(2 * time.Hour),
	}
	if err := repo.CreateRefreshSession(context.Background(), &RefreshSession{
		ID:        oldSessionID,
		UserID:    user.ID,
		TokenHash: HashRefreshToken(oldRefreshToken),
		ExpiresAt: time.Now().Add(2 * time.Hour),
	}); err != nil {
		t.Fatalf("CreateRefreshSession() error = %v", err)
	}

	result, err := service.Refresh(context.Background(), RefreshCommand{RefreshToken: oldRefreshToken}, SessionMeta{})
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}

	oldSession, err := repo.GetRefreshSessionByTokenHash(context.Background(), HashRefreshToken(oldRefreshToken))
	if err != nil {
		t.Fatalf("GetRefreshSessionByTokenHash() error = %v", err)
	}
	if oldSession.RevokedAt == nil {
		t.Fatalf("expected old refresh session to be revoked")
	}
	if result.Tokens.RefreshToken == oldRefreshToken {
		t.Fatalf("expected new refresh token to be issued")
	}
	if _, err := repo.GetRefreshSessionByTokenHash(context.Background(), HashRefreshToken(result.Tokens.RefreshToken)); err != nil {
		t.Fatalf("expected new refresh session to be stored: %v", err)
	}
}

func TestServiceLogout(t *testing.T) {
	t.Parallel()

	repo := newFakeAuthRepository()
	tokens := newFakeTokenManager()
	service := NewService(repo, tokens)

	userID := uuid.New()
	sessionID := uuid.New()
	refreshToken := "refresh-" + sessionID.String()
	tokens.claimsByToken[refreshToken] = &Claims{
		UserID:    userID,
		Type:      "refresh",
		TokenID:   sessionID.String(),
		ExpiresAt: time.Now().Add(2 * time.Hour),
	}
	if err := repo.CreateRefreshSession(context.Background(), &RefreshSession{
		ID:        sessionID,
		UserID:    userID,
		TokenHash: HashRefreshToken(refreshToken),
		ExpiresAt: time.Now().Add(2 * time.Hour),
	}); err != nil {
		t.Fatalf("CreateRefreshSession() error = %v", err)
	}

	if err := service.Logout(context.Background(), LogoutCommand{
		UserID:       userID,
		RefreshToken: refreshToken,
	}); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}

	session, err := repo.GetRefreshSessionByTokenHash(context.Background(), HashRefreshToken(refreshToken))
	if err != nil {
		t.Fatalf("GetRefreshSessionByTokenHash() error = %v", err)
	}
	if session.RevokedAt == nil {
		t.Fatalf("expected refresh session to be revoked")
	}
}
