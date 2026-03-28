package port

//go:generate go tool mockgen -destination=mock_${GOFILE} -package=${GOPACKAGE} -source=${GOFILE}

import "context"

// TokenPair holds the issued access and refresh tokens.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64 // access token TTL in seconds
}

// TokenClaims holds identity information extracted from a token.
type TokenClaims struct {
	CredentialID string
	Email        string
}

// TokenService defines the contract for issuing and validating JWT tokens.
type TokenService interface {
	// GenerateTokenPair creates a new access/refresh token pair.
	GenerateTokenPair(c context.Context, claims TokenClaims) (TokenPair, error)

	// ValidateAccessToken verifies an access token and returns its claims.
	ValidateAccessToken(c context.Context, token string) (TokenClaims, error)

	// ValidateRefreshToken verifies a refresh token and returns its claims.
	ValidateRefreshToken(c context.Context, token string) (TokenClaims, error)
}
