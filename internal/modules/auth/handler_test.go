package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"sipi/internal/testutil"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type fakeAuthService struct {
	register  func(ctx context.Context, command RegisterCommand, meta SessionMeta) (*AuthResult, error)
	login     func(ctx context.Context, command LoginCommand, meta SessionMeta) (*AuthResult, error)
	refresh   func(ctx context.Context, command RefreshCommand, meta SessionMeta) (*AuthResult, error)
	logout    func(ctx context.Context, command LogoutCommand) error
	logoutAll func(ctx context.Context, command LogoutAllCommand) error
}

func (f fakeAuthService) Register(ctx context.Context, command RegisterCommand, meta SessionMeta) (*AuthResult, error) {
	return f.register(ctx, command, meta)
}
func (f fakeAuthService) Login(ctx context.Context, command LoginCommand, meta SessionMeta) (*AuthResult, error) {
	return f.login(ctx, command, meta)
}
func (f fakeAuthService) Refresh(ctx context.Context, command RefreshCommand, meta SessionMeta) (*AuthResult, error) {
	return f.refresh(ctx, command, meta)
}
func (f fakeAuthService) Logout(ctx context.Context, command LogoutCommand) error {
	return f.logout(ctx, command)
}
func (f fakeAuthService) LogoutAll(ctx context.Context, command LogoutAllCommand) error {
	return f.logoutAll(ctx, command)
}

func TestHandlerRegister(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := fakeAuthService{
		register: func(_ context.Context, command RegisterCommand, _ SessionMeta) (*AuthResult, error) {
			return &AuthResult{
				User: AuthUser{ID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), Email: command.Email},
				Tokens: TokenPair{
					AccessToken:  "access-token",
					RefreshToken: "refresh-token",
					TokenType:    "Bearer",
				},
			}, nil
		},
		login:     func(context.Context, LoginCommand, SessionMeta) (*AuthResult, error) { return nil, nil },
		refresh:   func(context.Context, RefreshCommand, SessionMeta) (*AuthResult, error) { return nil, nil },
		logout:    func(context.Context, LogoutCommand) error { return nil },
		logoutAll: func(context.Context, LogoutAllCommand) error { return nil },
	}

	router := gin.New()
	handler := NewHandler(service)
	router.POST("/auth/register", handler.Register)

	rec := testutil.PerformJSONRequest(router, http.MethodPost, "/auth/register", []byte(`{"email":"user@example.com","password":"StrongPass123!"}`))
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d, body=%s", rec.Code, rec.Body.String())
	}

	var response struct {
		Success bool `json:"success"`
		Data    struct {
			User struct {
				Email string `json:"email"`
			} `json:"user"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if !response.Success || response.Data.User.Email != "user@example.com" {
		t.Fatalf("unexpected response: %s", rec.Body.String())
	}
}
