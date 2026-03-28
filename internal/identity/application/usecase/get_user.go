package usecase

import (
	"context"
	"fmt"

	"github.com/blackhorseya/go-ddd/internal/identity/application/dto"
	"github.com/blackhorseya/go-ddd/internal/identity/application/mapper"
	"github.com/blackhorseya/go-ddd/internal/identity/domain/user"
	"github.com/blackhorseya/go-ddd/pkg/contextx"
)

// GetUserUseCase retrieves a user by ID.
type GetUserUseCase struct {
	repo user.Repository
}

// NewGetUserUseCase creates a GetUserUseCase.
func NewGetUserUseCase(repo user.Repository) *GetUserUseCase {
	return &GetUserUseCase{repo: repo}
}

// Execute fetches a user by ID and returns the output DTO.
func (uc *GetUserUseCase) Execute(c context.Context, id string) (dto.UserOutput, error) {
	ctx := contextx.From(c)
	ctx = ctx.WithOperation("GetUser")

	u, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return dto.UserOutput{}, fmt.Errorf("find user: %w", err)
	}

	return mapper.ToUserOutput(u), nil
}
