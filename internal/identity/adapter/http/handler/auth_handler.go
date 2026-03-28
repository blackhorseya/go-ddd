package handler

import (
	"errors"

	"github.com/gin-gonic/gin"

	"github.com/blackhorseya/go-ddd/internal/identity/application/dto"
	"github.com/blackhorseya/go-ddd/internal/identity/application/usecase"
	"github.com/blackhorseya/go-ddd/internal/identity/domain/credential"
	"github.com/blackhorseya/go-ddd/internal/shared/adapter/http/response"
)

// AuthHandler handles authentication-related HTTP endpoints.
type AuthHandler struct {
	register *usecase.RegisterUseCase
	login    *usecase.LoginUseCase
	refresh  *usecase.RefreshTokenUseCase
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(
	register *usecase.RegisterUseCase,
	login *usecase.LoginUseCase,
	refresh *usecase.RefreshTokenUseCase,
) *AuthHandler {
	return &AuthHandler{register: register, login: login, refresh: refresh}
}

// Register registers auth routes on the given engine.
func (h *AuthHandler) Register(r *gin.Engine) {
	v1 := r.Group("/api/v1")
	{
		v1.POST("/auth/register", h.RegisterCredential)
		v1.POST("/auth/login", h.Login)
		v1.POST("/auth/refresh", h.RefreshToken)
	}
}

// RegisterCredential handles credential registration.
//
//	@Summary		Register a new credential
//	@Description	建立新的認證憑證
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.RegisterInput	true	"Registration payload"
//	@Success		201		{object}	response.Response{data=dto.CredentialOutput}
//	@Failure		400		{object}	response.Response
//	@Failure		409		{object}	response.Response
//	@Failure		500		{object}	response.Response
//	@Router			/api/v1/auth/register [post]
func (h *AuthHandler) RegisterCredential(c *gin.Context) {
	var input dto.RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	out, err := h.register.Execute(c.Request.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, credential.ErrEmailDuplicated):
			response.Conflict(c, "email already in use")
		case errors.Is(err, credential.ErrInvalidEmail),
			errors.Is(err, credential.ErrPasswordRequired),
			errors.Is(err, credential.ErrPasswordTooShort),
			errors.Is(err, credential.ErrPasswordTooLong):
			response.BadRequest(c, err.Error())
		default:
			response.InternalError(c, "failed to register credential")
		}
		return
	}

	response.Created(c, out)
}

// Login handles credential authentication.
//
//	@Summary		Authenticate a credential
//	@Description	使用 email 和密碼登入，回傳 JWT token pair
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.LoginInput	true	"Login payload"
//	@Success		200		{object}	response.Response{data=dto.AuthOutput}
//	@Failure		400		{object}	response.Response
//	@Failure		401		{object}	response.Response
//	@Failure		500		{object}	response.Response
//	@Router			/api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var input dto.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	out, err := h.login.Execute(c.Request.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidCredentials),
			errors.Is(err, credential.ErrCredentialSuspended):
			response.Unauthorized(c, "invalid email or password")
		default:
			response.InternalError(c, "failed to authenticate")
		}
		return
	}

	response.OK(c, out)
}

// RefreshToken handles token refresh.
//
//	@Summary		Refresh token pair
//	@Description	使用 refresh token 取得新的 token pair
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.RefreshInput	true	"Refresh payload"
//	@Success		200		{object}	response.Response{data=dto.AuthOutput}
//	@Failure		400		{object}	response.Response
//	@Failure		401		{object}	response.Response
//	@Failure		500		{object}	response.Response
//	@Router			/api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var input dto.RefreshInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	out, err := h.refresh.Execute(c.Request.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidCredentials),
			errors.Is(err, credential.ErrCredentialSuspended),
			errors.Is(err, credential.ErrNotFound):
			response.Unauthorized(c, "invalid or expired refresh token")
		default:
			response.InternalError(c, "failed to refresh token")
		}
		return
	}

	response.OK(c, out)
}
