package dto

import "time"

// RegisterInput carries data needed to register a new credential.
type RegisterInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// CredentialOutput represents a credential in API responses.
type CredentialOutput struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
