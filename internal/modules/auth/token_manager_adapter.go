package auth

import (
	"sipi/internal/platform/jwt"

	"github.com/google/uuid"
)

type TokenManagerAdapter struct {
	manager *jwt.Manager
}

func NewTokenManagerAdapter(manager *jwt.Manager) *TokenManagerAdapter {
	return &TokenManagerAdapter{manager: manager}
}

func (a *TokenManagerAdapter) GenerateAccessToken(userID uuid.UUID) (string, error) {
	return a.manager.GenerateAccessToken(userID)
}

func (a *TokenManagerAdapter) GenerateRefreshToken(userID, sessionID uuid.UUID) (string, error) {
	return a.manager.GenerateRefreshToken(userID, sessionID)
}

func (a *TokenManagerAdapter) Parse(token string) (*Claims, error) {
	claims, err := a.manager.Parse(token)
	if err != nil {
		return nil, err
	}

	result := &Claims{
		UserID:  claims.UserID,
		Type:    claims.Type,
		TokenID: claims.ID,
	}

	if claims.ExpiresAt != nil {
		result.ExpiresAt = claims.ExpiresAt.Time
	}

	return result, nil
}
