package auth

import "github.com/google/uuid"

type RegisterCommand struct {
	Email    string
	Password string
}

type LoginCommand struct {
	Email    string
	Password string
}

type RefreshCommand struct {
	RefreshToken string
}

type LogoutCommand struct {
	UserID       uuid.UUID
	RefreshToken string
}

type LogoutAllCommand struct {
	UserID uuid.UUID
}

type SessionMeta struct {
	UserAgent string
	IPAddress string
}

type AuthResult struct {
	User   AuthUser
	Tokens TokenPair
}

type AuthUser struct {
	ID    uuid.UUID
	Email string
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
}
