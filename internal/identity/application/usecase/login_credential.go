package usecase

import (
	"context"
	"errors"

	"github.com/blackhorseya/go-ddd/internal/identity/application/dto"
	"github.com/blackhorseya/go-ddd/internal/identity/application/mapper"
	"github.com/blackhorseya/go-ddd/internal/identity/application/port"
	"github.com/blackhorseya/go-ddd/internal/identity/domain/credential"
	"github.com/blackhorseya/go-ddd/pkg/contextx"
)

// ErrInvalidCredentials is returned when email or password is wrong.
// Intentionally vague to avoid leaking whether the email exists.
var ErrInvalidCredentials = errors.New("invalid email or password")

// LoginUseCase authenticates a credential and issues a token pair.
type LoginUseCase struct {
	repo     credential.Repository
	tokenSvc port.TokenService
}

// NewLoginUseCase creates a LoginUseCase.
func NewLoginUseCase(repo credential.Repository, tokenSvc port.TokenService) *LoginUseCase {
	return &LoginUseCase{repo: repo, tokenSvc: tokenSvc}
}

// Execute authenticates the user and returns an auth output with tokens.
func (uc *LoginUseCase) Execute(c context.Context, input dto.LoginInput) (dto.AuthOutput, error) {
	ctx := contextx.From(c)
	ctx = ctx.WithOperation("Login")

	email, err := credential.NewEmail(input.Email)
	if err != nil {
		ctx.Warn("login failed: invalid email format", "email", input.Email)
		return dto.AuthOutput{}, ErrInvalidCredentials
	}

	cred, err := uc.repo.FindByEmail(ctx, email)
	if err != nil {
		ctx.Warn("login failed: email not found", "email", input.Email)
		return dto.AuthOutput{}, ErrInvalidCredentials
	}

	if !cred.Password().Verify(input.Password) {
		ctx.Warn("login failed: wrong password", "credential_id", cred.ID())
		return dto.AuthOutput{}, ErrInvalidCredentials
	}

	if cred.IsSuspended() {
		ctx.Warn("login failed: credential suspended", "credential_id", cred.ID())
		return dto.AuthOutput{}, credential.ErrCredentialSuspended
	}

	pair, err := uc.tokenSvc.GenerateTokenPair(ctx, port.TokenClaims{
		CredentialID: cred.ID(),
		Email:        cred.Email().Address(),
	})
	if err != nil {
		ctx.Error("login failed: token generation error", "credential_id", cred.ID(), "error", err)
		return dto.AuthOutput{}, err
	}

	ctx.Info("credential authenticated", "credential_id", cred.ID(), "email", cred.Email().Address())

	return dto.AuthOutput{
		Credential:   mapper.ToCredentialOutput(cred),
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresIn:    pair.ExpiresIn,
	}, nil
}
