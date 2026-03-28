package usecase

import (
	"context"
	"fmt"

	"github.com/blackhorseya/go-ddd/internal/identity/application/dto"
	"github.com/blackhorseya/go-ddd/internal/identity/application/mapper"
	"github.com/blackhorseya/go-ddd/internal/identity/application/port"
	"github.com/blackhorseya/go-ddd/internal/identity/domain/user"
	"github.com/blackhorseya/go-ddd/pkg/contextx"
)

// RegisterUserUseCase handles new user registration.
type RegisterUserUseCase struct {
	repo  user.Repository
	idGen port.IDGenerator
}

// NewRegisterUserUseCase creates a RegisterUserUseCase.
func NewRegisterUserUseCase(repo user.Repository, idGen port.IDGenerator) *RegisterUserUseCase {
	return &RegisterUserUseCase{repo: repo, idGen: idGen}
}

// Execute registers a new user and returns the created user output.
func (uc *RegisterUserUseCase) Execute(c context.Context, input dto.RegisterUserInput) (dto.UserOutput, error) {
	ctx := contextx.From(c)
	ctx = ctx.WithOperation("RegisterUser")

	email, err := user.NewEmail(input.Email)
	if err != nil {
		return dto.UserOutput{}, fmt.Errorf("invalid email: %w", err)
	}

	password, err := user.NewHashedPassword(input.Password)
	if err != nil {
		return dto.UserOutput{}, fmt.Errorf("invalid password: %w", err)
	}

	u, err := user.NewUser(user.NewUserParams{
		ID:       uc.idGen.Generate(),
		Email:    email,
		Password: password,
		Name:     input.Name,
	})
	if err != nil {
		return dto.UserOutput{}, fmt.Errorf("create user: %w", err)
	}

	if err := uc.repo.Save(ctx, u); err != nil {
		return dto.UserOutput{}, fmt.Errorf("save user: %w", err)
	}

	ctx.Info("user registered", "user_id", u.ID(), "email", u.Email().Address())

	return mapper.ToUserOutput(u), nil
}
