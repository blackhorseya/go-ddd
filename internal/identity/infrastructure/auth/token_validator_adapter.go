package auth

import (
	"context"

	"github.com/blackhorseya/go-ddd/internal/identity/application/port"
	"github.com/blackhorseya/go-ddd/internal/shared/adapter/http/middleware"
)

// Compile-time check.
var _ middleware.AccessTokenValidator = (*TokenValidatorAdapter)(nil)

// TokenValidatorAdapter adapts port.TokenService to middleware.AccessTokenValidator.
type TokenValidatorAdapter struct {
	svc port.TokenService
}

// NewTokenValidatorAdapter creates a TokenValidatorAdapter.
func NewTokenValidatorAdapter(svc port.TokenService) *TokenValidatorAdapter {
	return &TokenValidatorAdapter{svc: svc}
}

// ValidateAccessToken validates the token and returns the credential ID.
func (a *TokenValidatorAdapter) ValidateAccessToken(c context.Context, token string) (string, error) {
	claims, err := a.svc.ValidateAccessToken(c, token)
	if err != nil {
		return "", err
	}

	return claims.CredentialID, nil
}
