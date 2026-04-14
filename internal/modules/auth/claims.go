package auth

import (
	"time"

	"github.com/google/uuid"
)

type Claims struct {
	UserID    uuid.UUID
	Type      string
	TokenID   string
	ExpiresAt time.Time
}
