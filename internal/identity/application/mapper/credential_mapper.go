package mapper

import (
	"github.com/blackhorseya/go-ddd/internal/identity/application/dto"
	"github.com/blackhorseya/go-ddd/internal/identity/domain/credential"
)

// ToCredentialOutput converts a domain Credential to a CredentialOutput DTO.
func ToCredentialOutput(c *credential.Credential) dto.CredentialOutput {
	return dto.CredentialOutput{
		ID:        c.ID(),
		Email:     c.Email().Address(),
		Status:    string(c.Status()),
		CreatedAt: c.CreatedAt(),
		UpdatedAt: c.UpdatedAt(),
	}
}
