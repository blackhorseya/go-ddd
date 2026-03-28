package usecase

import (
	"context"
	"fmt"

	"github.com/blackhorseya/go-ddd/internal/identity/application/dto"
	"github.com/blackhorseya/go-ddd/internal/identity/application/mapper"
	"github.com/blackhorseya/go-ddd/internal/identity/application/port"
	"github.com/blackhorseya/go-ddd/internal/identity/domain/credential"
	"github.com/blackhorseya/go-ddd/pkg/contextx"
)

// RegisterUseCase handles new credential registration.
type RegisterUseCase struct {
	repo  credential.Repository
	idGen port.IDGenerator
}

// NewRegisterUseCase creates a RegisterUseCase.
func NewRegisterUseCase(repo credential.Repository, idGen port.IDGenerator) *RegisterUseCase {
	return &RegisterUseCase{repo: repo, idGen: idGen}
}

// Execute registers a new credential and returns the created output.
func (uc *RegisterUseCase) Execute(c context.Context, input dto.RegisterInput) (dto.CredentialOutput, error) {
	ctx := contextx.From(c)
	ctx = ctx.WithOperation("Register")

	email, err := credential.NewEmail(input.Email)
	if err != nil {
		return dto.CredentialOutput{}, fmt.Errorf("invalid email: %w", err)
	}

	password, err := credential.NewHashedPassword(input.Password)
	if err != nil {
		return dto.CredentialOutput{}, fmt.Errorf("invalid password: %w", err)
	}

	cred, err := credential.NewCredential(credential.NewCredentialParams{
		ID:       uc.idGen.Generate(),
		Email:    email,
		Password: password,
	})
	if err != nil {
		return dto.CredentialOutput{}, fmt.Errorf("create credential: %w", err)
	}

	if err := uc.repo.Save(ctx, cred); err != nil {
		return dto.CredentialOutput{}, fmt.Errorf("save credential: %w", err)
	}

	ctx.Info("credential registered", "credential_id", cred.ID(), "email", cred.Email().Address())

	return mapper.ToCredentialOutput(cred), nil
}
