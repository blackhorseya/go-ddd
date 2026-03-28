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
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(register *usecase.RegisterUseCase) *AuthHandler {
	return &AuthHandler{register: register}
}

// Register registers auth routes on the given engine.
func (h *AuthHandler) Register(r *gin.Engine) {
	v1 := r.Group("/api/v1")
	{
		v1.POST("/auth/register", h.RegisterCredential)
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
