package dto

import "time"

// RegisterInput carries data needed to register a new credential.
type RegisterInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginInput carries data needed to authenticate a credential.
type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RefreshInput carries data needed to refresh a token pair.
type RefreshInput struct {
	RefreshToken string `json:"refresh_token"`
}

// CredentialOutput represents a credential in API responses.
type CredentialOutput struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AuthOutput represents an authentication response with tokens.
type AuthOutput struct {
	Credential   CredentialOutput `json:"credential"`
	AccessToken  string           `json:"access_token"`
	RefreshToken string           `json:"refresh_token"`
	ExpiresIn    int64            `json:"expires_in"`
}
