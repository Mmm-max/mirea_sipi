package auth

import "github.com/google/uuid"

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email,max=320" example:"alice@example.com"`
	Password string `json:"password" binding:"required,min=8,max=128" example:"StrongPassword123!"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email,max=320" example:"alice@example.com"`
	Password string `json:"password" binding:"required,min=8,max=128" example:"StrongPassword123!"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

type Tokens struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	TokenType    string `json:"token_type" example:"Bearer"`
}

type AuthResponseData struct {
	User   UserResponse `json:"user"`
	Tokens Tokens       `json:"tokens"`
}

type UserResponse struct {
	ID    string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email string `json:"email" example:"alice@example.com"`
}

type AuthResponse struct {
	Success bool             `json:"success" example:"true"`
	Data    AuthResponseData `json:"data"`
	Meta    ResponseMetaDTO  `json:"meta"`
}

type MessageData struct {
	Message string `json:"message" example:"logged out successfully"`
}

type MessageEnvelope struct {
	Success bool            `json:"success" example:"true"`
	Data    MessageData     `json:"data"`
	Meta    ResponseMetaDTO `json:"meta"`
}

type ResponseMetaDTO struct {
	RequestID string `json:"request_id" example:"e7cc5d4a-3d85-4938-bff3-d2a6cf8ca2bb"`
}

func (r RegisterRequest) ToCommand() RegisterCommand {
	return RegisterCommand{
		Email:    r.Email,
		Password: r.Password,
	}
}

func (r LoginRequest) ToCommand() LoginCommand {
	return LoginCommand{
		Email:    r.Email,
		Password: r.Password,
	}
}

func (r RefreshRequest) ToCommand() RefreshCommand {
	return RefreshCommand{
		RefreshToken: r.RefreshToken,
	}
}

func (r LogoutRequest) ToCommand(userID uuid.UUID) LogoutCommand {
	return LogoutCommand{
		UserID:       userID,
		RefreshToken: r.RefreshToken,
	}
}
