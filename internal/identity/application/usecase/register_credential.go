package usecase

import (
	"context"
	"fmt"

	"github.com/blackhorseya/go-ddd/internal/identity/application/dto"
	"github.com/blackhorseya/go-ddd/internal/identity/application/mapper"
	"github.com/blackhorseya/go-ddd/internal/identity/application/port"
	"github.com/blackhorseya/go-ddd/internal/identity/domain/credential"
	"github.com/blackhorseya/go-ddd/internal/shared/domain/event"
	"github.com/blackhorseya/go-ddd/pkg/contextx"
)

// RegisterUseCase handles new credential registration.
type RegisterUseCase struct {
	repo  credential.Repository
	idGen port.IDGenerator
	bus   event.EventBus
}

// NewRegisterUseCase creates a RegisterUseCase.
func NewRegisterUseCase(repo credential.Repository, idGen port.IDGenerator, bus event.EventBus) *RegisterUseCase {
	return &RegisterUseCase{repo: repo, idGen: idGen, bus: bus}
}

// Execute registers a new credential and returns the created output.
func (uc *RegisterUseCase) Execute(c context.Context, input dto.RegisterInput) (dto.CredentialOutput, error) {
	ctx := contextx.From(c)
	ctx = ctx.WithOperation("Register")

	email, err := credential.NewEmail(input.Email)
	if err != nil {
		ctx.Warn("register failed: invalid email", "email", input.Email, "error", err)
		return dto.CredentialOutput{}, fmt.Errorf("invalid email: %w", err)
	}

	password, err := credential.NewHashedPassword(input.Password)
	if err != nil {
		ctx.Warn("register failed: invalid password", "error", err)
		return dto.CredentialOutput{}, fmt.Errorf("invalid password: %w", err)
	}

	// Check for duplicate email before creating the credential.
	if _, err := uc.repo.FindByEmail(ctx, email); err == nil {
		ctx.Warn("register failed: email already in use", "email", input.Email)
		return dto.CredentialOutput{}, credential.ErrEmailDuplicated
	}

	cred, err := credential.NewCredential(credential.NewCredentialParams{
		ID:       uc.idGen.Generate(),
		Email:    email,
		Password: password,
	})
	if err != nil {
		ctx.Error("register failed: create credential", "error", err)
		return dto.CredentialOutput{}, fmt.Errorf("create credential: %w", err)
	}

	if err := uc.repo.Save(ctx, cred); err != nil {
		ctx.Warn("register failed: save credential", "email", input.Email, "error", err)
		return dto.CredentialOutput{}, fmt.Errorf("save credential: %w", err)
	}

	// Publish domain event — failure is logged but does not fail the use case.
	evt := credential.NewCredentialRegistered(cred.ID(), cred.Email().Address())
	if pubErr := uc.bus.Publish(ctx, evt); pubErr != nil {
		ctx.Warn("failed to publish credential.registered event", "error", pubErr)
	}

	ctx.Info("credential registered", "credential_id", cred.ID(), "email", cred.Email().Address())

	return mapper.ToCredentialOutput(cred), nil
}
