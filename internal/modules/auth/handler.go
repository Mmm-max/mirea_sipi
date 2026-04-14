package auth

import (
	"net/http"

	httpmiddleware "sipi/internal/http/middleware"
	"sipi/internal/platform/httpx"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service ServicePort
}

func NewHandler(service ServicePort) *Handler {
	return &Handler{service: service}
}

// Register godoc
// @Summary Register user
// @Description Creates a new user account and returns access/refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Register request"
// @Success 201 {object} AuthResponse
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 409 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var request RegisterRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpx.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	result, err := h.service.Register(c.Request.Context(), request.ToCommand(), sessionMetaFromContext(c))
	if err != nil {
		httpx.HandleServiceError(c, err)
		return
	}

	httpx.Created(c, newAuthResponseData(result))
}

// Login godoc
// @Summary Login user
// @Description Authenticates user and returns access/refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login request"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var request LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpx.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	result, err := h.service.Login(c.Request.Context(), request.ToCommand(), sessionMetaFromContext(c))
	if err != nil {
		httpx.HandleServiceError(c, err)
		return
	}

	httpx.OK(c, newAuthResponseData(result))
}

// Refresh godoc
// @Summary Refresh token pair
// @Description Invalidates the current refresh token and returns a new access/refresh token pair
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshRequest true "Refresh request"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /auth/refresh [post]
func (h *Handler) Refresh(c *gin.Context) {
	var request RefreshRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpx.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	result, err := h.service.Refresh(c.Request.Context(), request.ToCommand(), sessionMetaFromContext(c))
	if err != nil {
		httpx.HandleServiceError(c, err)
		return
	}

	httpx.OK(c, newAuthResponseData(result))
}

// Logout godoc
// @Summary Logout from current session
// @Description Revokes the provided refresh token for the current authenticated user
// @Tags auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body LogoutRequest true "Logout request"
// @Success 200 {object} MessageEnvelope
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 403 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	userID, ok := httpmiddleware.RequireUserID(c)
	if !ok {
		return
	}

	var request LogoutRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpx.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	if err := h.service.Logout(c.Request.Context(), request.ToCommand(userID)); err != nil {
		httpx.HandleServiceError(c, err)
		return
	}

	httpx.OK(c, MessageData{Message: "logged out successfully"})
}

// LogoutAll godoc
// @Summary Logout from all sessions
// @Description Revokes all refresh sessions for the current authenticated user
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} MessageEnvelope
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /auth/logout-all [post]
func (h *Handler) LogoutAll(c *gin.Context) {
	userID, ok := httpmiddleware.RequireUserID(c)
	if !ok {
		return
	}

	if err := h.service.LogoutAll(c.Request.Context(), LogoutAllCommand{UserID: userID}); err != nil {
		httpx.HandleServiceError(c, err)
		return
	}

	httpx.OK(c, MessageData{Message: "logged out from all sessions successfully"})
}

func sessionMetaFromContext(c *gin.Context) SessionMeta {
	return SessionMeta{
		UserAgent: c.Request.UserAgent(),
		IPAddress: c.ClientIP(),
	}
}

func newAuthResponseData(result *AuthResult) AuthResponseData {
	return AuthResponseData{
		User: UserResponse{
			ID:    result.User.ID.String(),
			Email: result.User.Email,
		},
		Tokens: Tokens{
			AccessToken:  result.Tokens.AccessToken,
			RefreshToken: result.Tokens.RefreshToken,
			TokenType:    result.Tokens.TokenType,
		},
	}
}
