package credential

//go:generate go tool mockgen -destination=mock_${GOFILE} -package=${GOPACKAGE} -source=${GOFILE}

import (
	"context"

	domain "github.com/blackhorseya/go-ddd/internal/shared/domain"
)

// Repository defines the persistence contract for the Credential aggregate.
type Repository interface {
	// Save persists a credential. It should upsert (create or update).
	Save(c context.Context, cred *Credential) error

	// FindByID retrieves a credential by its unique identifier.
	FindByID(c context.Context, id string) (*Credential, error)

	// FindByEmail retrieves a credential by email address.
	FindByEmail(c context.Context, email Email) (*Credential, error)

	// Delete removes a credential by its unique identifier.
	Delete(c context.Context, id string) error

	// List retrieves a paginated list of credentials.
	List(c context.Context, req domain.PageRequest) (domain.PageResult[*Credential], error)
}
