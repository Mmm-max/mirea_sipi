package jwt

import (
	"fmt"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Manager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

type TokenClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Type   string    `json:"type"`
	jwtv5.RegisteredClaims
}

func NewManager(secret string, accessTTL, refreshTTL time.Duration) *Manager {
	return &Manager{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (m *Manager) GenerateAccessToken(userID uuid.UUID) (string, error) {
	return m.generateToken(userID, uuid.New(), "access", m.accessTTL)
}

func (m *Manager) GenerateRefreshToken(userID, sessionID uuid.UUID) (string, error) {
	return m.generateToken(userID, sessionID, "refresh", m.refreshTTL)
}

func (m *Manager) Parse(token string) (*TokenClaims, error) {
	parsed, err := jwtv5.ParseWithClaims(token, &TokenClaims{}, func(parsedToken *jwtv5.Token) (any, error) {
		if _, ok := parsedToken.Method.(*jwtv5.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}

		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := parsed.Claims.(*TokenClaims)
	if !ok || !parsed.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

func (m *Manager) generateToken(userID, tokenID uuid.UUID, tokenType string, ttl time.Duration) (string, error) {
	now := time.Now()

	claims := TokenClaims{
		UserID: userID,
		Type:   tokenType,
		RegisteredClaims: jwtv5.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwtv5.NewNumericDate(now),
			ExpiresAt: jwtv5.NewNumericDate(now.Add(ttl)),
			ID:        tokenID.String(),
		},
	}

	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims)

	return token.SignedString(m.secret)
}
