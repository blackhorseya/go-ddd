package usecase

import (
	"context"
	"fmt"

	"github.com/blackhorseya/go-ddd/internal/identity/application/dto"
	"github.com/blackhorseya/go-ddd/internal/identity/application/mapper"
	"github.com/blackhorseya/go-ddd/internal/identity/domain/credential"
	"github.com/blackhorseya/go-ddd/pkg/contextx"
)

// GetCredentialUseCase retrieves a credential by ID.
type GetCredentialUseCase struct {
	repo credential.Repository
}

// NewGetCredentialUseCase creates a GetCredentialUseCase.
func NewGetCredentialUseCase(repo credential.Repository) *GetCredentialUseCase {
	return &GetCredentialUseCase{repo: repo}
}

// Execute fetches a credential by ID and returns the output DTO.
func (uc *GetCredentialUseCase) Execute(c context.Context, id string) (dto.CredentialOutput, error) {
	ctx := contextx.From(c)
	ctx = ctx.WithOperation("GetCredential")

	cred, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return dto.CredentialOutput{}, fmt.Errorf("find credential: %w", err)
	}

	return mapper.ToCredentialOutput(cred), nil
}
