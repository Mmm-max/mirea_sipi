package auth

import (
	"time"

	"github.com/google/uuid"
)

type RefreshSession struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	UserAgent string
	IPAddress string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Account struct {
	ID           uuid.UUID
	Email        string
	FullName     string
	Timezone     string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
