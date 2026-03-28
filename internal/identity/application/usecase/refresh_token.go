package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/blackhorseya/go-ddd/internal/identity/application/dto"
	"github.com/blackhorseya/go-ddd/internal/identity/application/mapper"
	"github.com/blackhorseya/go-ddd/internal/identity/application/port"
	"github.com/blackhorseya/go-ddd/internal/identity/domain/credential"
	"github.com/blackhorseya/go-ddd/pkg/contextx"
)

// RefreshTokenUseCase refreshes a token pair using a valid refresh token.
type RefreshTokenUseCase struct {
	repo     credential.Repository
	tokenSvc port.TokenService
}

// NewRefreshTokenUseCase creates a RefreshTokenUseCase.
func NewRefreshTokenUseCase(repo credential.Repository, tokenSvc port.TokenService) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{repo: repo, tokenSvc: tokenSvc}
}

// Execute validates the refresh token, checks credential status, and issues a new token pair.
func (uc *RefreshTokenUseCase) Execute(c context.Context, input dto.RefreshInput) (dto.AuthOutput, error) {
	ctx := contextx.From(c)
	ctx = ctx.WithOperation("RefreshToken")

	claims, err := uc.tokenSvc.ValidateRefreshToken(ctx, input.RefreshToken)
	if err != nil {
		ctx.Warn("refresh failed: invalid refresh token", "error", err)
		return dto.AuthOutput{}, ErrInvalidCredentials
	}

	cred, err := uc.repo.FindByID(ctx, claims.CredentialID)
	if err != nil {
		if errors.Is(err, credential.ErrNotFound) {
			ctx.Warn("refresh failed: credential not found", "credential_id", claims.CredentialID)
			return dto.AuthOutput{}, ErrInvalidCredentials
		}
		ctx.Error("refresh failed: find credential", "credential_id", claims.CredentialID, "error", err)
		return dto.AuthOutput{}, fmt.Errorf("find credential: %w", err)
	}

	if cred.IsSuspended() {
		ctx.Warn("refresh failed: credential suspended", "credential_id", cred.ID())
		return dto.AuthOutput{}, credential.ErrCredentialSuspended
	}

	pair, err := uc.tokenSvc.GenerateTokenPair(ctx, port.TokenClaims{
		CredentialID: cred.ID(),
		Email:        cred.Email().Address(),
	})
	if err != nil {
		ctx.Error("refresh failed: token generation error", "credential_id", cred.ID(), "error", err)
		return dto.AuthOutput{}, err
	}

	ctx.Info("token refreshed", "credential_id", cred.ID())

	return dto.AuthOutput{
		Credential:   mapper.ToCredentialOutput(cred),
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresIn:    pair.ExpiresIn,
	}, nil
}
