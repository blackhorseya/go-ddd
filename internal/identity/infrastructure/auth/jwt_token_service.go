package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/blackhorseya/go-ddd/internal/identity/application/port"
)

var (
	errInvalidToken = errors.New("invalid token")
	errTokenExpired = errors.New("token expired")
)

// tokenType distinguishes access tokens from refresh tokens.
type tokenType string

const (
	tokenTypeAccess  tokenType = "access"
	tokenTypeRefresh tokenType = "refresh"
)

type customClaims struct {
	jwt.RegisteredClaims

	CredentialID string    `json:"cid"`
	Email        string    `json:"email,omitempty"`
	TokenType    tokenType `json:"typ"`
}

// JWTConfig holds JWT signing parameters.
type JWTConfig struct {
	Secret          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

// JWTTokenService implements port.TokenService using golang-jwt.
type JWTTokenService struct {
	cfg JWTConfig
}

// NewJWTTokenService creates a JWTTokenService.
func NewJWTTokenService(cfg JWTConfig) port.TokenService {
	if cfg.AccessTokenTTL == 0 {
		cfg.AccessTokenTTL = 15 * time.Minute
	}

	if cfg.RefreshTokenTTL == 0 {
		cfg.RefreshTokenTTL = 7 * 24 * time.Hour
	}

	return &JWTTokenService{cfg: cfg}
}

func (s *JWTTokenService) GenerateTokenPair(_ context.Context, claims port.TokenClaims) (port.TokenPair, error) {
	now := time.Now()

	accessToken, err := s.signToken(customClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   claims.CredentialID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.AccessTokenTTL)),
		},
		CredentialID: claims.CredentialID,
		Email:        claims.Email,
		TokenType:    tokenTypeAccess,
	})
	if err != nil {
		return port.TokenPair{}, err
	}

	refreshToken, err := s.signToken(customClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   claims.CredentialID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.RefreshTokenTTL)),
		},
		CredentialID: claims.CredentialID,
		TokenType:    tokenTypeRefresh,
	})
	if err != nil {
		return port.TokenPair{}, err
	}

	return port.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.cfg.AccessTokenTTL.Seconds()),
	}, nil
}

func (s *JWTTokenService) ValidateAccessToken(_ context.Context, token string) (port.TokenClaims, error) {
	return s.validateToken(token, tokenTypeAccess)
}

func (s *JWTTokenService) ValidateRefreshToken(_ context.Context, token string) (port.TokenClaims, error) {
	return s.validateToken(token, tokenTypeRefresh)
}

func (s *JWTTokenService) signToken(claims customClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.Secret))
}

func (s *JWTTokenService) validateToken(raw string, expectedType tokenType) (port.TokenClaims, error) {
	token, err := jwt.ParseWithClaims(raw, &customClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errInvalidToken
		}
		return []byte(s.cfg.Secret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return port.TokenClaims{}, errTokenExpired
		}
		return port.TokenClaims{}, errInvalidToken
	}

	claims, ok := token.Claims.(*customClaims)
	if !ok || !token.Valid {
		return port.TokenClaims{}, errInvalidToken
	}

	if claims.TokenType != expectedType {
		return port.TokenClaims{}, errInvalidToken
	}

	return port.TokenClaims{
		CredentialID: claims.CredentialID,
		Email:        claims.Email,
	}, nil
}
