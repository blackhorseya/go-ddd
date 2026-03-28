package handler

import (
	"errors"

	"github.com/gin-gonic/gin"

	"github.com/blackhorseya/go-ddd/internal/identity/application/dto"
	"github.com/blackhorseya/go-ddd/internal/identity/application/usecase"
	"github.com/blackhorseya/go-ddd/internal/identity/domain/user"
	"github.com/blackhorseya/go-ddd/internal/shared/adapter/http/response"
)

// UserHandler handles user-related HTTP endpoints.
type UserHandler struct {
	register *usecase.RegisterUserUseCase
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(register *usecase.RegisterUserUseCase) *UserHandler {
	return &UserHandler{register: register}
}

// Register registers user routes on the given engine.
func (h *UserHandler) Register(r *gin.Engine) {
	v1 := r.Group("/api/v1")
	{
		v1.POST("/users", h.RegisterUser)
	}
}

// RegisterUser handles user registration.
//
//	@Summary		Register a new user
//	@Description	建立新使用者帳號
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.RegisterUserInput					true	"Registration payload"
//	@Success		201		{object}	response.Response{data=dto.UserOutput}
//	@Failure		400		{object}	response.Response{error=response.Error}
//	@Failure		409		{object}	response.Response{error=response.Error}
//	@Failure		500		{object}	response.Response{error=response.Error}
//	@Router			/api/v1/users [post]
func (h *UserHandler) RegisterUser(c *gin.Context) {
	var input dto.RegisterUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	out, err := h.register.Execute(c.Request.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrEmailDuplicated):
			response.Conflict(c, "email already in use")
		case errors.Is(err, user.ErrInvalidEmail),
			errors.Is(err, user.ErrEmptyName),
			errors.Is(err, user.ErrNameTooLong),
			errors.Is(err, user.ErrPasswordTooShort),
			errors.Is(err, user.ErrPasswordTooLong):
			response.BadRequest(c, err.Error())
		default:
			response.InternalError(c, "failed to register user")
		}
		return
	}

	response.Created(c, out)
}
